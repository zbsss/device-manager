package devicemanager

import "sync"

type Pod struct {
	Id          string
	MemoryQuota float64
	MemoryLimit uint64
	MemoryUsed  uint64
}

type Device struct {
	mut         *sync.Mutex
	Id          string
	MemoryTotal uint64
	MemoryUsed  uint64
	Pods        map[string]*Pod
}
