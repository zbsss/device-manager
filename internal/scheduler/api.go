package scheduler

import (
	"fmt"
)

type Scheduler interface {
	EnqueueLeaseRequest(req *TokenLeaseRequest)
	ReturnLease(lease *TokenLease) error
	AllocatePodQuota(podQuota *PodQuota) error
	GetAvailableQuota() float64
}

func (s *scheduler) GetAvailableQuota() float64 {
	s.lock.Lock()
	defer s.lock.Unlock()

	availableQuota := 1.0
	for _, podQuota := range s.podQuota {
		availableQuota -= podQuota.Requests
	}

	return availableQuota
}

func (s *scheduler) AllocatePodQuota(podQuota *PodQuota) error {
	availableQuota := s.GetAvailableQuota()
	if availableQuota <= 0 {
		return fmt.Errorf("not enough quota available")
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	s.podQuota[podQuota.PodId] = podQuota
	return nil
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
