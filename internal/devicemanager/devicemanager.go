package devicemanager

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	pb "github.com/zbsss/device-manager/generated"
	"github.com/zbsss/device-manager/internal/scheduler"
)

type DeviceManager struct {
	pb.UnimplementedDeviceManagerServer
	devices            map[string]*Device
	schedulerPerDevice map[string]scheduler.Scheduler
	sf                 scheduler.SchedulerFactory
}

func NewDeviceManager(schedulerWindow, schedulerTokenExpiration time.Duration) *DeviceManager {
	dm := &DeviceManager{
		devices:            make(map[string]*Device),
		schedulerPerDevice: make(map[string]scheduler.Scheduler),
		sf:                 scheduler.NewSchedulerFactory(schedulerWindow, schedulerTokenExpiration),
	}

	go dm.stateLoggerDaemon()

	return dm
}

func NewTestDeviceManager(schedulerWindow, schedulerTokenExpiration time.Duration) *DeviceManager {
	dm := &DeviceManager{
		devices: map[string]*Device{
			"dev-1": {
				mut:          &sync.RWMutex{},
				Vendor:       "example.com",
				Model:        "mydev",
				Id:           "dev-1",
				MemoryBTotal: 10000000000000,
				MemoryBUsed:  0,
				Pods:         make(map[string]*Pod),
			},
		},
		schedulerPerDevice: make(map[string]scheduler.Scheduler),
		sf:                 scheduler.NewSchedulerFactory(schedulerWindow, schedulerTokenExpiration),
	}

	sch := dm.sf.StartScheduler("dev-1")
	sch.ReservePodQuota(&scheduler.PodQuota{
		PodId: "device", Requests: 0.5, Limit: 1.0,
	})
	dm.schedulerPerDevice["dev-1"] = sch

	go dm.stateLoggerDaemon()
	go dm.runGarbageCollector()

	return dm
}

func (dm *DeviceManager) stateLoggerDaemon() {
	for {
		var sb strings.Builder
		sb.WriteString("\n===Current state===")
		for _, device := range dm.devices {
			sb.WriteString(fmt.Sprintf("\nDevice %s: %d/%d", device.Id, device.MemoryBUsed, device.MemoryBTotal))

			for _, pod := range device.Pods {
				sb.WriteString(fmt.Sprintf("\n\tPod %s: %d/%d", pod.Id, pod.MemoryBUsed, pod.MemoryBLimit))
			}
		}
		sb.WriteString("\n===================\n")
		log.Println(sb.String())

		time.Sleep(30 * time.Second)
	}
}
