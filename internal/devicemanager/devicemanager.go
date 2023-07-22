package devicemanager

import (
	"log"
	"strings"
	"sync"
	"time"

	"github.com/zbsss/device-manager/internal/scheduler"
	pb "github.com/zbsss/device-manager/pkg/devicemanager"
)

type DeviceManager struct {
	pb.UnimplementedDeviceManagerServer
	lock    *sync.RWMutex
	devices map[string]*Device
	sf      scheduler.SchedulerFactory
}

func NewDeviceManager(schedulerWindow, schedulerTokenExpiration time.Duration) *DeviceManager {
	dm := &DeviceManager{
		lock:    &sync.RWMutex{},
		devices: make(map[string]*Device),
		sf:      scheduler.NewSchedulerFactory(schedulerWindow, schedulerTokenExpiration),
	}

	go dm.stateLoggerDaemon()
	go dm.runGarbageCollector()

	return dm
}

func (dm *DeviceManager) GetDev(deviceId string) *Device {
	dm.lock.RLock()
	defer dm.lock.RUnlock()

	return dm.devices[deviceId]
}

func (dm *DeviceManager) stateLoggerDaemon() {
	for {
		var sb strings.Builder
		sb.WriteString("\n===Current state===")

		sb.WriteString("\n---Devices---")

		dm.lock.RLock()
		for _, device := range dm.devices {
			sb.WriteString(
				"\nDeviceId: " + device.Id +
					"\nVendor: " + device.Vendor +
					"\nModel: " + device.Model +
					"\nAllocatorPodId: " + device.AllocatorPodId,
			)
		}

		sb.WriteString("\n\n---Memory Utilization---")

		for _, device := range dm.devices {
			sb.WriteString(device.mm.PrintState())
		}
		sb.WriteString("\n\n---CPU Utilization---")
		for _, device := range dm.devices {
			sb.WriteString(device.sch.PrintState())
		}
		dm.lock.RUnlock()

		sb.WriteString("\n===================")

		log.Println(sb.String())
		time.Sleep(30 * time.Second)
	}
}
