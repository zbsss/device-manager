package devicemanager

import (
	"sync"

	"github.com/zbsss/device-manager/internal/memorymanager"
	"github.com/zbsss/device-manager/internal/scheduler"
)

type Device struct {
	lock      *sync.RWMutex
	mm        memorymanager.MemoryManager
	scheduler scheduler.Scheduler

	Id             string
	AllocatorPodId string
	Vendor         string
	Model          string
	Pods           map[string]bool
}
