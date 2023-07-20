package devicemanager

import "sync"

type Pod struct {
	Id           string
	MemoryQuota  float64
	MemoryBLimit uint64
	MemoryBUsed  uint64
}

type Device struct {
	mut          *sync.RWMutex
	Vendor       string
	Model        string
	Id           string
	MemoryBTotal uint64
	MemoryBUsed  uint64
	Pods         map[string]*Pod
}
