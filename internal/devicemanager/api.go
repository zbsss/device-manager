package devicemanager

import (
	"context"
	"fmt"
	"log"
	"sync"

	pb "github.com/zbsss/device-manager/generated"
	"github.com/zbsss/device-manager/internal/scheduler"
)

func (dm *DeviceManager) GetAvailableDevices(ctx context.Context, in *pb.GetAvailableDevicesRequest) (*pb.GetAvailableDevicesReply, error) {
	log.Printf("Received: GetAvailableResources")

	var devices []*pb.FreeDeviceResources

	for _, device := range dm.devices {
		device.mut.RLock()
		if device.Vendor == in.Vendor && device.Model == in.Model {
			devices = append(devices, &pb.FreeDeviceResources{
				DeviceId: device.Id,
				Memory:   getAvailableMemory(device),
				Requests: dm.schedulerPerDevice[device.Id].GetAvailableQuota(),
			})
		}
		device.mut.RUnlock()
	}

	return &pb.GetAvailableDevicesReply{Free: devices}, nil
}

func (dm *DeviceManager) GetToken(ctx context.Context, in *pb.GetTokenRequest) (*pb.GetTokenReply, error) {
	log.Printf("Received: GetToken for device %s from pod %s", in.DeviceId, in.PodId)

	if in.DeviceId == "" {
		return nil, fmt.Errorf("device not specified")
	}
	if in.PodId == "" {
		return nil, fmt.Errorf("pod not specified")
	}

	req := &scheduler.TokenLeaseRequest{
		PodId:    in.PodId,
		Response: make(chan *scheduler.TokenLease),
	}

	dm.schedulerPerDevice[in.DeviceId].EnqueueLeaseRequest(req)
	token := <-req.Response

	log.Printf("Sending token")

	return &pb.GetTokenReply{ExpiresAt: token.ExpiresAt.Unix()}, nil
}

func (dm *DeviceManager) ReturnToken(ctx context.Context, in *pb.ReturnTokenRequest) (*pb.ReturnTokenReply, error) {
	log.Printf("Received: ReturnToken")

	if in.DeviceId == "" {
		return nil, fmt.Errorf("device not specified")
	}
	if in.PodId == "" {
		return nil, fmt.Errorf("pod not specified")
	}

	err := dm.schedulerPerDevice[in.DeviceId].ReturnLease(&scheduler.TokenLease{PodId: in.PodId})

	if err != nil {
		log.Printf("Error returning token: %s", err)
	} else {
		log.Printf("Token returned")
	}

	return &pb.ReturnTokenReply{}, nil
}

func (dm *DeviceManager) AllocateMemory(ctx context.Context, in *pb.AllocateMemoryRequest) (*pb.AllocateMemoryReply, error) {
	log.Printf("Received: GetMemoryQuota for device %s: %d", in.DeviceId, in.MemoryB)

	if in.DeviceId == "" {
		return nil, fmt.Errorf("device not specified")
	}
	if in.PodId == "" {
		return nil, fmt.Errorf("pod not specified")
	}
	if in.MemoryB <= 0 {
		return nil, fmt.Errorf("memory value is invalid")
	}

	var device *Device
	var pod *Pod
	var ok bool

	if device, ok = dm.devices[in.DeviceId]; !ok {
		return nil, fmt.Errorf("device %s not registered", in.DeviceId)
	}

	if pod, ok = device.Pods[in.PodId]; !ok {
		return nil, fmt.Errorf("pod %s not registered", in.PodId)
	}

	device.mut.Lock()
	defer device.mut.Unlock()

	if device.MemoryBUsed+in.MemoryB > device.MemoryBTotal || pod.MemoryBUsed+in.MemoryB > pod.MemoryBLimit {
		return nil, fmt.Errorf("OOM: memory limit exceeded")
	}

	device.MemoryBUsed += in.MemoryB
	pod.MemoryBUsed += in.MemoryB

	return &pb.AllocateMemoryReply{}, nil
}

func (dm *DeviceManager) FreeMemory(ctx context.Context, in *pb.FreeMemoryRequest) (*pb.FreeMemoryReply, error) {
	log.Printf("Received: ReturnMemoryQuota for device %s: %d", in.DeviceId, in.MemoryB)

	if in.DeviceId == "" {
		return nil, fmt.Errorf("device not specified")
	}
	if in.PodId == "" {
		return nil, fmt.Errorf("pod not specified")
	}
	if in.MemoryB <= 0 {
		return nil, fmt.Errorf("memory value is invalid")
	}

	var device *Device
	var pod *Pod
	var ok bool

	if device, ok = dm.devices[in.DeviceId]; !ok {
		return nil, fmt.Errorf("device %s not registered", in.DeviceId)
	}

	if pod, ok = device.Pods[in.PodId]; !ok {
		return nil, fmt.Errorf("pod %s not registered", in.PodId)
	}

	device.mut.Lock()
	defer device.mut.Unlock()

	device.MemoryBUsed -= in.MemoryB
	pod.MemoryBUsed -= in.MemoryB

	log.Println(device.MemoryBUsed)

	return &pb.FreeMemoryReply{}, nil
}

func (dm *DeviceManager) RegisterDevice(ctx context.Context, in *pb.RegisterDeviceRequest) (*pb.RegisterDeviceReply, error) {
	log.Printf("Received: RegisterDevice for device %s", in.DeviceId)

	if _, ok := dm.devices[in.DeviceId]; ok {
		return nil, fmt.Errorf("device already registered")
	}

	dm.devices[in.DeviceId] = &Device{
		mut:          &sync.RWMutex{},
		Id:           in.DeviceId,
		MemoryBTotal: in.MemoryB,
		MemoryBUsed:  0,
		Pods:         map[string]*Pod{},
	}

	dm.schedulerPerDevice[in.DeviceId] = dm.sf.StartScheduler(in.DeviceId)

	return &pb.RegisterDeviceReply{}, nil
}

func (dm *DeviceManager) ReservePodQuota(ctx context.Context, in *pb.ReservePodQuotaRequest) (*pb.ReservePodQuotaReply, error) {
	log.Printf("Received: RegisterPod for device %s and pod %s", in.DeviceId, in.PodId)

	if in.DeviceId == "" {
		return nil, fmt.Errorf("device not specified")
	}
	if in.PodId == "" {
		return nil, fmt.Errorf("pod not specified")
	}

	var device *Device
	var ok bool

	if device, ok = dm.devices[in.DeviceId]; !ok {
		return nil, fmt.Errorf("device %s not registered", in.DeviceId)
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

	device.mut.Lock()
	defer device.mut.Unlock()

	err := dm.schedulerPerDevice[in.DeviceId].ReservePodQuota(
		&scheduler.PodQuota{
			PodId: in.PodId, Requests: in.Requests, Limit: in.Limit,
		},
	)
	if err != nil {
		return nil, err
	}

	availableMemory := getAvailableMemory(device)
	if in.Memory > availableMemory {
		return nil, fmt.Errorf("OOM: memory limit exceeded")
	}

	device.Pods[in.PodId] = &Pod{
		Id:           in.PodId,
		MemoryQuota:  in.Memory,
		MemoryBLimit: uint64(in.Memory * float64(device.MemoryBTotal)),
		MemoryBUsed:  0,
	}

	return &pb.ReservePodQuotaReply{}, nil
}

func (dm *DeviceManager) UnreservePodQuota(ctx context.Context, in *pb.UnreservePodQuotaRequest) (*pb.UnreservePodQuotaQuotaReply, error) {
	log.Println("Received: UnreservePodQuota")

	if in.DeviceId == "" {
		return nil, fmt.Errorf("device not specified")
	}
	if in.PodId == "" {
		return nil, fmt.Errorf("pod not specified")
	}

	dm.unreservePodQuota(in.DeviceId, in.PodId)

	return &pb.UnreservePodQuotaQuotaReply{}, nil
}

func (dm *DeviceManager) unreservePodQuota(deviceId, podId string) {
	device := dm.devices[deviceId]
	if device == nil {
		return
	}

	device.mut.Lock()
	defer device.mut.Unlock()

	pod := device.Pods[podId]
	if pod == nil {
		return
	}

	dm.schedulerPerDevice[deviceId].UnreservePodQuota(podId)
	device.MemoryBUsed -= pod.MemoryBUsed
	delete(device.Pods, podId)
}

func getAvailableMemory(device *Device) float64 {
	availableQuota := 1.0
	for _, pod := range device.Pods {
		availableQuota -= pod.MemoryQuota
	}
	return availableQuota
}
