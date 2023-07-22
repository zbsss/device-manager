package devicemanager

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/zbsss/device-manager/internal/memorymanager"
	"github.com/zbsss/device-manager/internal/scheduler"
	pb "github.com/zbsss/device-manager/pkg/devicemanager"
)

func (dm *DeviceManager) GetAvailableDevices(ctx context.Context, in *pb.GetAvailableDevicesRequest) (*pb.GetAvailableDevicesReply, error) {
	log.Printf("Received: GetAvailableResources")

	var devices []*pb.FreeDeviceResources

	dm.lock.RLock()
	defer dm.lock.RUnlock()

	for _, device := range dm.devices {
		device.lock.RLock()
		if device.Vendor == in.Vendor && device.Model == in.Model {
			devices = append(devices, &pb.FreeDeviceResources{
				DeviceId: device.Id,
				Memory:   device.mm.GetAvailableQuota(),
				Requests: device.sch.GetAvailableQuota(),
			})
		}
		device.lock.RUnlock()
	}

	log.Printf("Returning %v devices", devices)

	return &pb.GetAvailableDevicesReply{Free: devices}, nil
}

func (dm *DeviceManager) GetToken(ctx context.Context, in *pb.GetTokenRequest) (*pb.GetTokenReply, error) {
	// log.Printf("Received: GetToken for device %s from pod %s", in.DeviceId, in.PodId)

	if in.DeviceId == "" {
		return nil, fmt.Errorf("device not specified")
	}
	if in.PodId == "" {
		return nil, fmt.Errorf("pod not specified")
	}

	device := dm.GetDev(in.DeviceId)
	if device == nil {
		return nil, fmt.Errorf("device %s not registered", in.DeviceId)
	}

	req := &scheduler.TokenLeaseRequest{
		PodId:    in.PodId,
		Response: make(chan *scheduler.TokenLease),
	}

	device.sch.EnqueueLeaseRequest(req)
	token := <-req.Response

	if token == nil {
		return nil, fmt.Errorf("token not available")
	}

	return &pb.GetTokenReply{ExpiresAt: token.ExpiresAt.Unix()}, nil
}

func (dm *DeviceManager) ReturnToken(ctx context.Context, in *pb.ReturnTokenRequest) (*pb.ReturnTokenReply, error) {
	// log.Printf("Received: ReturnToken")

	if in.DeviceId == "" {
		return nil, fmt.Errorf("device not specified")
	}
	if in.PodId == "" {
		return nil, fmt.Errorf("pod not specified")
	}

	device := dm.GetDev(in.DeviceId)
	if device == nil {
		return nil, fmt.Errorf("device %s not registered", in.DeviceId)
	}

	err := device.sch.ReturnLease(&scheduler.TokenLease{PodId: in.PodId})
	if err != nil {
		log.Printf("Error returning token: %s", err)
	}

	return &pb.ReturnTokenReply{}, nil
}

func (dm *DeviceManager) AllocateMemory(ctx context.Context, in *pb.AllocateMemoryRequest) (*pb.AllocateMemoryReply, error) {
	// log.Printf("Received: GetMemoryQuota for device %s: %d", in.DeviceId, in.MemoryB)

	if in.DeviceId == "" {
		return nil, fmt.Errorf("device not specified")
	}
	if in.PodId == "" {
		return nil, fmt.Errorf("pod not specified")
	}
	if in.MemoryB <= 0 {
		return nil, fmt.Errorf("memory value is invalid")
	}

	device := dm.GetDev(in.DeviceId)
	if device == nil {
		return nil, fmt.Errorf("device %s not registered", in.DeviceId)
	}

	err := device.mm.AllocateMemory(in.PodId, in.MemoryB)
	if err != nil {
		return nil, err
	}

	return &pb.AllocateMemoryReply{}, nil
}

func (dm *DeviceManager) FreeMemory(ctx context.Context, in *pb.FreeMemoryRequest) (*pb.FreeMemoryReply, error) {
	// log.Printf("Received: ReturnMemoryQuota for device %s: %d", in.DeviceId, in.MemoryB)

	if in.DeviceId == "" {
		return nil, fmt.Errorf("device not specified")
	}
	if in.PodId == "" {
		return nil, fmt.Errorf("pod not specified")
	}
	if in.MemoryB <= 0 {
		return nil, fmt.Errorf("memory value is invalid")
	}

	device := dm.GetDev(in.DeviceId)
	if device == nil {
		return nil, fmt.Errorf("device %s not registered", in.DeviceId)
	}

	device.mm.FreeMemory(in.PodId, in.MemoryB)

	return &pb.FreeMemoryReply{}, nil
}

func (dm *DeviceManager) RegisterDevice(ctx context.Context, in *pb.RegisterDeviceRequest) (*pb.RegisterDeviceReply, error) {
	log.Printf("Received: RegisterDevice for device %s", in.DeviceId)

	if in.AllocatorPodId == "" {
		return nil, fmt.Errorf("allocator pod not specified")
	}
	if in.DeviceId == "" {
		return nil, fmt.Errorf("device not specified")
	}
	if in.Vendor == "" {
		return nil, fmt.Errorf("vendor not specified")
	}
	if in.Model == "" {
		return nil, fmt.Errorf("model not specified")
	}

	if dm.GetDev(in.DeviceId) != nil {
		return nil, fmt.Errorf("device %s already registered", in.DeviceId)
	}

	dm.lock.Lock()
	defer dm.lock.Unlock()

	dm.devices[in.DeviceId] = &Device{
		lock:           &sync.RWMutex{},
		sch:            dm.sf.StartScheduler(in.DeviceId),
		mm:             memorymanager.NewMemoryManager(in.DeviceId, in.MemoryB),
		Id:             in.DeviceId,
		AllocatorPodId: in.AllocatorPodId,
		Vendor:         in.Vendor,
		Model:          in.Model,
		Pods:           map[string]bool{},
		LastUsedAt:     time.Now(),
	}

	return &pb.RegisterDeviceReply{}, nil
}

func (dm *DeviceManager) ReservePodQuota(ctx context.Context, in *pb.ReservePodQuotaRequest) (*pb.ReservePodQuotaReply, error) {
	// log.Printf("Received: RegisterPod for device %s and pod %s", in.DeviceId, in.PodId)

	if in.DeviceId == "" {
		return nil, fmt.Errorf("device not specified")
	}
	if in.PodId == "" {
		return nil, fmt.Errorf("pod not specified")
	}

	if in.Requests > in.Limit {
		return nil, fmt.Errorf("requests > limit")
	}
	if in.Requests < 0 || in.Limit < 0 || in.Memory < 0 {
		return nil, fmt.Errorf("requests, limit and memory must be positive")
	}
	if in.Limit == 0 {
		in.Limit = in.Requests
	}

	device := dm.GetDev(in.DeviceId)
	if device == nil {
		return nil, fmt.Errorf("device %s not registered", in.DeviceId)
	}

	device.lock.Lock()
	defer device.lock.Unlock()

	err := device.sch.ReservePodQuota(
		&scheduler.PodQuota{
			PodId: in.PodId, Requests: in.Requests, Limit: in.Limit,
		},
	)
	if err != nil {
		return nil, err
	}

	err = device.mm.ReservePodQuota(in.PodId, in.Memory)
	if err != nil {
		device.sch.UnreservePodQuota(in.PodId)
		return nil, err
	}

	device.Pods[in.PodId] = true

	return &pb.ReservePodQuotaReply{}, nil
}
