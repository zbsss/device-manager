package main

import "sync"

var schedulers = map[string]*scheduler{}

func StartScheduler(deviceId string) {
	s := &scheduler{
		lock:         sync.Mutex{},
		queue:        []*TokenLeaseRequest{},
		currentLease: nil,
		leaseHistory: []*LeaseHistoryEntry{},
	}

	schedulers[deviceId] = s

	go s.run()
}

func GetScheduler(deviceId string) *scheduler {
	return schedulers[deviceId]
}
