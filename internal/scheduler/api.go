package scheduler

import (
	"fmt"
)

type Scheduler interface {
	EnqueueLeaseRequest(req *TokenLeaseRequest)
	ReturnLease(lease *TokenLease) error
	UpdatePodQuota(podQuota *PodQuota)
}

func (s *scheduler) UpdatePodQuota(podQuota *PodQuota) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.podQuota[podQuota.PodId] = podQuota
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
