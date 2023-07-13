package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"sync"
	"time"
)

type TokenLease struct {
	PodId     string
	ExpiresAt time.Time
}

type TokenLeaseRequest struct {
	PodId    string
	Response chan *TokenLease
}

type LeaseHistoryEntry struct {
	PodId      string    `json:"podId"`
	LeasedAt   time.Time `json:"leasedAt"`
	ReturnedAt time.Time `json:"returnedAt"`
}

type PodQuota struct {
	PodId    string
	Requests float64
	Limit    float64
}

type scheduler struct {
	lock                sync.Mutex
	queue               []*TokenLeaseRequest
	currentLease        *TokenLease
	leaseHistory        []*LeaseHistoryEntry
	leaseHistoryLogFile string
	podQuota            map[string]*PodQuota
}

func (s *scheduler) run() {
	for {
		s.tryScheduleLease()
		s.tryTerminateExpiredLease()
	}
}

func (s *scheduler) tryScheduleLease() {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.currentLease == nil && len(s.queue) > 0 {
		// calculate used quota for each pod in current time window
		usedQuotaPerPod := s.calculateUsedQuotaPerPod()

		// split requests into 3 groups based on used quota
		usedLessThanRequests, usedLessThanLimit, usedLimit := s.splitRequestsByQuotaUsed(usedQuotaPerPod)

		// order by used quota
		var selected *TokenLeaseRequest

		if len(usedLessThanRequests) > 0 {
			sortAscByQuotaUsed(usedLessThanRequests, usedQuotaPerPod)
			selected = usedLessThanRequests[0]
		} else if len(usedLessThanLimit) > 0 {
			sortAscByQuotaUsed(usedLessThanLimit, usedQuotaPerPod)
			selected = usedLessThanLimit[0]
		} else {
			sortAscByQuotaUsed(usedLimit, usedQuotaPerPod)
			selected = usedLimit[0]
		}

		s.currentLease = &TokenLease{
			PodId:     selected.PodId,
			ExpiresAt: time.Now().Add(tokenDuration),
		}

		selected.Response <- s.currentLease

		// remove `selected` from s.queue
		for i, v := range s.queue {
			if v == selected {
				s.queue = append(s.queue[:i], s.queue[i+1:]...)
				break
			}
		}
	}
}

func (s *scheduler) tryTerminateExpiredLease() {
	s.lock.Lock()

	if s.currentLease != nil && time.Now().After(s.currentLease.ExpiresAt) {
		log.Printf("Lease for pod %s has expired\n", s.currentLease.PodId)

		s.lock.Unlock()
		s.ReturnLease(s.currentLease)
	} else {
		s.lock.Unlock()
	}
}

func (s *scheduler) calculateUsedQuotaPerPod() map[string]float64 {
	hist := []*LeaseHistoryEntry{}
	leaseDurationPerPod := map[string]time.Duration{}

	windowStart := time.Now().Add(-windowDuration)

	for _, entry := range s.leaseHistory {
		if (entry.LeasedAt.Before(windowStart) && entry.ReturnedAt.After(windowStart)) || entry.LeasedAt.After(windowStart) {
			if entry.LeasedAt.Before(windowStart) {
				entry.LeasedAt = windowStart
			}

			hist = append(hist, entry)
			leaseDurationPerPod[entry.PodId] += entry.ReturnedAt.Sub(entry.LeasedAt)
		}
	}

	s.leaseHistory = hist

	usedQuotaPerPod := map[string]float64{}
	for pod, duration := range leaseDurationPerPod {
		usedQuotaPerPod[pod] = duration.Seconds() / windowDuration.Seconds()
	}

	return usedQuotaPerPod
}

func (s *scheduler) splitRequestsByQuotaUsed(usedQuotaPerPod map[string]float64) ([]*TokenLeaseRequest, []*TokenLeaseRequest, []*TokenLeaseRequest) {
	usedLimit := []*TokenLeaseRequest{}
	usedLessThanRequests := []*TokenLeaseRequest{}
	usedLessThanLimit := []*TokenLeaseRequest{}

	for _, pod := range s.queue {
		if usedQuotaPerPod[pod.PodId] < s.podQuota[pod.PodId].Requests {
			usedLessThanRequests = append(usedLessThanRequests, pod)
		} else if usedQuotaPerPod[pod.PodId] < s.podQuota[pod.PodId].Limit {
			usedLessThanLimit = append(usedLessThanLimit, pod)
		} else {
			usedLimit = append(usedLimit, pod)
		}
	}

	return usedLessThanRequests, usedLessThanLimit, usedLimit
}

func sortAscByQuotaUsed(requests []*TokenLeaseRequest, usedQuotaPerPod map[string]float64) {
	sort.Slice(requests, func(i, j int) bool {
		return usedQuotaPerPod[requests[i].PodId] < usedQuotaPerPod[requests[j].PodId]
	})
}

func (s *scheduler) EnqueueLeaseRequest(req *TokenLeaseRequest) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.queue = append(s.queue, req)
}

func (s *scheduler) ReturnLease(lease *TokenLease) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.currentLease == nil {
		return fmt.Errorf("no lease to return")
	}

	if s.currentLease.PodId != lease.PodId {
		return fmt.Errorf("pod %s does not have a lease", lease.PodId)
	}

	newHistEntry := LeaseHistoryEntry{
		PodId:      lease.PodId,
		LeasedAt:   s.currentLease.ExpiresAt.Add(-tokenDuration),
		ReturnedAt: time.Now(),
	}

	s.leaseHistory = append([]*LeaseHistoryEntry{&newHistEntry}, s.leaseHistory...)

	s.currentLease = nil

	go s.saveLeaseHistoryEntry(newHistEntry)

	return nil
}

func (s *scheduler) saveLeaseHistoryEntry(entry LeaseHistoryEntry) {
	file, err := os.OpenFile(s.leaseHistoryLogFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		log.Printf("Failed to open file: %s\n", err)
	}
	defer file.Close()

	data, err := json.Marshal(entry)
	if err != nil {
		log.Printf("Failed to marshal entry: %s\n", err)
	}

	_, err = file.Write(data)
	if err != nil {
		log.Printf("Failed to write entry: %s\n", err)
	}

	_, _ = file.WriteString("\n")

	log.Printf("Saved entry: %s\n", data)
}

func (s *scheduler) UpdatePodQuota(podQuota *PodQuota) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.podQuota[podQuota.PodId] = podQuota
}
