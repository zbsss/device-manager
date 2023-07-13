package scheduler

import "time"

type SchedulerFactory interface {
	StartScheduler(deviceId string) Scheduler
}

type schedulerFactory struct {
	windowDuration time.Duration
	evictionPeriod time.Duration
}

func NewSchedulerFactory(windowDuration, evictionPeriod time.Duration) *schedulerFactory {
	return &schedulerFactory{
		windowDuration: windowDuration,
		evictionPeriod: evictionPeriod,
	}
}

func (sf *schedulerFactory) StartScheduler(deviceId string) Scheduler {
	return startScheduler(deviceId, sf.windowDuration, sf.evictionPeriod)
}
