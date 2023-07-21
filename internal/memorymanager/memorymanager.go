package memorymanager

import (
	"fmt"
	"strings"
	"sync"
)

type PodMemory struct {
	Id           string
	MemoryQuota  float64
	MemoryBLimit uint64
	MemoryBUsed  uint64
}

type MemoryManager interface {
	AllocateMemory(podId string, memoryB uint64) error
	FreeMemory(podId string, memoryB uint64)

	GetAvailableQuota() float64
	ReservePodQuota(podId string, memoryQuota float64) error
	UnreservePodQuota(podId string)

	PrintState() string
}

type memoryManager struct {
	lock         *sync.RWMutex
	DeviceId     string
	MemoryBTotal uint64
	MemoryBUsed  uint64
	PodsMem      map[string]*PodMemory
}

func NewMemoryManager(deviceId string, memoryBTotal uint64) MemoryManager {
	return &memoryManager{
		lock:         &sync.RWMutex{},
		DeviceId:     deviceId,
		MemoryBTotal: memoryBTotal,
		MemoryBUsed:  0,
		PodsMem:      map[string]*PodMemory{},
	}
}

func (mm *memoryManager) GetAvailableQuota() float64 {
	mm.lock.RLock()
	defer mm.lock.RUnlock()

	availableQuota := 1.0
	for _, podMem := range mm.PodsMem {
		availableQuota -= podMem.MemoryQuota
	}
	return availableQuota
}

func (mm *memoryManager) ReservePodQuota(podId string, memoryQuota float64) error {
	availableQuota := mm.GetAvailableQuota()

	if memoryQuota > availableQuota {
		return fmt.Errorf("OOM: memory limit exceeded")
	}

	mm.lock.Lock()
	defer mm.lock.Unlock()

	mm.PodsMem[podId] = &PodMemory{
		Id:           podId,
		MemoryQuota:  memoryQuota,
		MemoryBLimit: uint64(memoryQuota * float64(mm.MemoryBTotal)),
		MemoryBUsed:  0,
	}

	return nil
}

func (mm *memoryManager) UnreservePodQuota(podId string) {
	mm.lock.Lock()
	defer mm.lock.Unlock()

	pod := mm.PodsMem[podId]
	if pod == nil {
		return
	}

	mm.MemoryBUsed -= mm.PodsMem[podId].MemoryBUsed
	delete(mm.PodsMem, podId)
}

func (mm *memoryManager) AllocateMemory(podId string, memoryB uint64) error {
	mm.lock.Lock()
	defer mm.lock.Unlock()

	pod := mm.PodsMem[podId]
	if pod == nil {
		return fmt.Errorf("pod %s not registered", podId)
	}

	if mm.MemoryBUsed+memoryB > mm.MemoryBTotal || pod.MemoryBUsed+memoryB > pod.MemoryBLimit {
		return fmt.Errorf("OOM: memory limit exceeded")
	}

	mm.MemoryBUsed += memoryB
	pod.MemoryBUsed += memoryB

	return nil
}

func (mm *memoryManager) FreeMemory(podId string, memoryB uint64) {
	mm.lock.Lock()
	defer mm.lock.Unlock()

	pod := mm.PodsMem[podId]
	if pod == nil {
		return
	}

	mm.MemoryBUsed -= memoryB
	pod.MemoryBUsed -= memoryB
}

func (mm *memoryManager) PrintState() string {
	mm.lock.RLock()
	defer mm.lock.RUnlock()

	var sb strings.Builder

	totalMemUsed := float64(mm.MemoryBUsed) / float64(mm.MemoryBTotal)
	sb.WriteString(fmt.Sprintf("\nDevice %s: %f", mm.DeviceId, totalMemUsed))

	for _, pod := range mm.PodsMem {
		podMemUsed := float64(pod.MemoryBUsed) / float64(pod.MemoryBLimit)
		sb.WriteString(fmt.Sprintf("\n\tPod %s: %f", pod.Id, podMemUsed))
	}

	return sb.String()
}
