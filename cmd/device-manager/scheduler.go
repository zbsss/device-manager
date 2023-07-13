package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
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

		for podId, podQuota := range s.podQuota {
			log.Printf("Pod %s: used quota %f, limit %f\n", podId, usedQuotaPerPod[podId], podQuota.Limit)
		}

		// order by used/requested quota
		sort.Slice(s.queue, func(i, j int) bool {
			iUsed := usedQuotaPerPod[s.queue[i].PodId]
			jUsed := usedQuotaPerPod[s.queue[j].PodId]
			iQuota := iUsed / s.podQuota[s.queue[i].PodId].Requests
			jQuota := jUsed / s.podQuota[s.queue[j].PodId].Requests
			if iUsed >= s.podQuota[s.queue[i].PodId].Limit {
				iQuota = math.MaxFloat64
			}
			if jUsed >= s.podQuota[s.queue[j].PodId].Limit {
				jQuota = math.MaxFloat64
			}
			return iQuota < jQuota
		})

		selected := s.queue[0]

		if usedQuotaPerPod[selected.PodId] >= s.podQuota[selected.PodId].Limit {
			log.Printf("Pod %s has reached its quota limit\n", selected.PodId)
			return
		}

		s.currentLease = &TokenLease{
			PodId:     selected.PodId,
			ExpiresAt: time.Now().Add(tokenDuration),
		}

		selected.Response <- s.currentLease

		// remove `selected` from s.queue
		s.queue = s.queue[1:]
	}
}

func (s *scheduler) tryTerminateExpiredLease() {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.currentLease != nil && time.Now().After(s.currentLease.ExpiresAt) {
		log.Printf("Lease for pod %s has expired\n", s.currentLease.PodId)

		err := EvictPod(s.currentLease.PodId, "default")
		if err != nil {
			log.Printf("Failed to evict pod %s: %v\n", s.currentLease.PodId, err)
			return
		}

		// TODO: release memory owned by this pod

		log.Printf("Pod %s evicted\n", s.currentLease.PodId)

		s.cancelLeaseNoLock()
	}
}

func (s *scheduler) calculateUsedQuotaPerPod() map[string]float64 {
	hist := []*LeaseHistoryEntry{}
	leaseDurationPerPod := map[string]time.Duration{}

	windowStart := time.Now().Add(-windowDuration)

	for _, entry := range s.leaseHistory {
		if entry.ReturnedAt.After(windowStart) {
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

func (s *scheduler) EnqueueLeaseRequest(req *TokenLeaseRequest) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.queue = append(s.queue, req)
}

func (s *scheduler) cancelLeaseNoLock() {
	newHistEntry := LeaseHistoryEntry{
		PodId:      s.currentLease.PodId,
		LeasedAt:   s.currentLease.ExpiresAt.Add(-tokenDuration),
		ReturnedAt: time.Now(),
	}

	s.leaseHistory = append([]*LeaseHistoryEntry{&newHistEntry}, s.leaseHistory...)

	s.currentLease = nil

	go s.saveLeaseHistoryEntry(newHistEntry)
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

	s.cancelLeaseNoLock()

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
