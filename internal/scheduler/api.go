package scheduler

import (
	"fmt"
	"strings"
)

type Scheduler interface {
	Stop()

	EnqueueLeaseRequest(req *TokenLeaseRequest)
	ReturnLease(lease *TokenLease) error

	GetAvailableQuota() float64
	ReservePodQuota(podQuota *PodQuota) error
	UnreservePodQuota(podId string)

	PrintState() string
}

func (s *scheduler) Stop() {
	s.isRunning.Store(false)
}

func (s *scheduler) PrintState() string {
	s.lock.RLock()
	defer s.lock.RUnlock()

	var sb strings.Builder
	total := 0.0
	usedQuota := s.calculateUsedQuotaPerPod()
	for podId, podQuota := range s.podQuota {
		total += usedQuota[podId]
		sb.WriteString(fmt.Sprintf("\n\tPod %s: %f (req: %f, limit: %f)", podId, usedQuota[podId], podQuota.Requests, podQuota.Limit))
	}
	return fmt.Sprintf("\nDevice %s: %f", s.deviceId, total) + sb.String()
}

func (s *scheduler) GetAvailableQuota() float64 {
	s.lock.RLock()
	defer s.lock.RUnlock()

	availableQuota := 1.0
	for _, podQuota := range s.podQuota {
		availableQuota -= podQuota.Requests
	}

	return availableQuota
}

func (s *scheduler) ReservePodQuota(podQuota *PodQuota) error {
	availableQuota := s.GetAvailableQuota()
	if availableQuota <= 0 {
		return fmt.Errorf("not enough quota available")
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	s.podQuota[podQuota.PodId] = podQuota
	return nil
}

func (s *scheduler) UnreservePodQuota(podId string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.currentLease != nil && s.currentLease.PodId == podId {
		s.currentLease = nil
	}

	if len(s.queue) > 0 {
		for i, req := range s.queue {
			if req.PodId == podId {
				close(req.Response)
				s.queue = append(s.queue[:i], s.queue[i+1:]...)
			}
		}
	}

	delete(s.podQuota, podId)
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

	s.cancelLeaseNoLock()

	return nil
}
