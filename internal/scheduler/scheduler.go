package scheduler

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

type scheduler struct {
	isRunning           atomic.Bool
	lock                sync.RWMutex
	deviceId            string
	queue               []*TokenLeaseRequest
	currentLease        *TokenLease
	leaseHistory        []*LeaseHistoryEntry
	leaseHistoryLogFile string
	podQuota            map[string]*PodQuota
	windowDuration      time.Duration
	evictionPeriod      time.Duration
}

func startScheduler(deviceId string, windowDuration, evictionPeriod time.Duration) Scheduler {
	s := &scheduler{
		lock:         sync.RWMutex{},
		deviceId:     deviceId,
		queue:        []*TokenLeaseRequest{},
		currentLease: nil,
		leaseHistory: []*LeaseHistoryEntry{},
		// TODO: when running in Pod we need some sidecar to upload logs to S3?
		leaseHistoryLogFile: fmt.Sprintf("data/data-%s.json", time.Now().Format("2006-01-02-15-04-05")),
		podQuota:            map[string]*PodQuota{},
		windowDuration:      windowDuration,
		evictionPeriod:      evictionPeriod,
	}

	go s.run()
	return s
}

func (s *scheduler) run() {
	s.isRunning.Store(true)

	for s.isRunning.Load() {
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
			ExpiresAt: time.Now().Add(s.evictionPeriod),
		}

		selected.Response <- s.currentLease

		// remove `selected` from s.queue
		s.queue = s.queue[1:]
	}
}

func (s *scheduler) areOtherPodsInQueueNoLock(podId string) bool {
	for _, req := range s.queue {
		if req.PodId != podId {
			return true
		}
	}

	return false
}

func (s *scheduler) tryTerminateExpiredLease() {
	s.lock.Lock()
	defer s.lock.Unlock()

	// only evict Pod if there are other Pods waiting in the queue
	if s.currentLease != nil &&
		time.Now().After(s.currentLease.ExpiresAt) &&
		s.areOtherPodsInQueueNoLock(s.currentLease.PodId) {
		log.Printf("Lease for pod %s has expired\n", s.currentLease.PodId)

		err := evictPod(s.currentLease.PodId, "default")
		if err != nil {
			log.Printf("Failed to evict pod %s: %v\n", s.currentLease.PodId, err)
			return
		}

		// Once the Pod is deleted it will be garbage collected by the DeviceManager
		log.Printf("Pod %s evicted\n", s.currentLease.PodId)

		s.cancelLeaseNoLock()
	}
}

func (s *scheduler) calculateUsedQuotaPerPod() map[string]float64 {
	hist := []*LeaseHistoryEntry{}
	leaseDurationPerPod := map[string]time.Duration{}

	windowStart := time.Now().Add(-s.windowDuration)

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
		usedQuotaPerPod[pod] = duration.Seconds() / s.windowDuration.Seconds()
	}

	return usedQuotaPerPod
}

func (s *scheduler) cancelLeaseNoLock() {
	newHistEntry := LeaseHistoryEntry{
		PodId:      s.currentLease.PodId,
		LeasedAt:   s.currentLease.ExpiresAt.Add(-s.evictionPeriod),
		ReturnedAt: time.Now(),
	}

	s.leaseHistory = append([]*LeaseHistoryEntry{&newHistEntry}, s.leaseHistory...)

	s.currentLease = nil

	// go s.saveLeaseHistoryEntry(newHistEntry)
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

	// TODO: refactor
	_, err = file.Write(data)
	if err != nil {
		log.Printf("Failed to write entry: %s\n", err)
	}

	// TODO: refactor
	_, _ = file.WriteString("\n")

	log.Printf("Saved entry: %s\n", data)
}
