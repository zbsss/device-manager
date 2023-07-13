package main

import (
	"fmt"
	"sync"
	"time"
)

var schedulers = map[string]*scheduler{}

func StartScheduler(deviceId string) {
	s := &scheduler{
		lock:                sync.Mutex{},
		queue:               []*TokenLeaseRequest{},
		currentLease:        nil,
		leaseHistory:        []*LeaseHistoryEntry{},
		// TODO: when running in Pod we need some sidecar to upload logs to S3?
		leaseHistoryLogFile: fmt.Sprintf("data/data-%s.json", time.Now().Format("2006-01-02-15-04-05")),
		podQuota:            map[string]*PodQuota{},
	}

	schedulers[deviceId] = s

	go s.run()
}

func GetScheduler(deviceId string) *scheduler {
	// TODO: handle when scheduler does not exist
	return schedulers[deviceId]
}
