package devicemanager

import (
	"context"
	"fmt"
	"log"
	"sync"

	pb "github.com/zbsss/device-manager/generated"
	"github.com/zbsss/device-manager/internal/scheduler"
)

func (dm *DeviceManager) GetAvailableResources(ctx context.Context, in *pb.GetAvailableDevicesRequest) (*pb.GetAvailableDevicesReply, error) {
	log.Printf("Received: GetAvailableResources")

	var devices []*pb.FreeDeviceResources

	for _, device := range dm.devices {
		if device.Vendor == in.Vendor && device.Model == in.Model {
			devices = append(devices, &pb.FreeDeviceResources{
				DeviceId: device.Id,
				Memory:   float64(device.MemoryUsed) / float64(device.MemoryTotal),
				Requests: dm.schedulerPerDevice[device.Id].GetAvailableQuota(),
			})
		}
	}

	return &pb.GetAvailableDevicesReply{Free: devices}, nil
}

func (dm *DeviceManager) GetToken(ctx context.Context, in *pb.GetTokenRequest) (*pb.GetTokenReply, error) {
	log.Printf("Received: GetToken for device %s from pod %s", in.Device, in.Pod)

	if in.Device == "" {
		return nil, fmt.Errorf("device not specified")
	}
	if in.Pod == "" {
		return nil, fmt.Errorf("pod not specified")
	}

	req := &scheduler.TokenLeaseRequest{
		PodId:    in.Pod,
		Response: make(chan *scheduler.TokenLease),
	}

	dm.schedulerPerDevice[in.Device].EnqueueLeaseRequest(req)
	token := <-req.Response

	log.Printf("Sending token")

	return &pb.GetTokenReply{ExpiresAt: token.ExpiresAt.Unix()}, nil
}

func (dm *DeviceManager) ReturnToken(ctx context.Context, in *pb.ReturnTokenRequest) (*pb.ReturnTokenReply, error) {
	log.Printf("Received: ReturnToken")

	if in.Device == "" {
		return nil, fmt.Errorf("device not specified")
	}
	if in.Pod == "" {
		return nil, fmt.Errorf("pod not specified")
	}

	err := dm.schedulerPerDevice[in.Device].ReturnLease(&scheduler.TokenLease{PodId: in.Pod})

	if err != nil {
		log.Printf("Error returning token: %s", err)
	} else {
		log.Printf("Token returned")
	}

	return &pb.ReturnTokenReply{}, nil
}

func (dm *DeviceManager) GetMemoryQuota(ctx context.Context, in *pb.GetMemoryQuotaRequest) (*pb.GetMemoryQuotaReply, error) {
	log.Printf("Received: GetMemoryQuota for device %s: %d", in.Device, in.Memory)

	if in.Device == "" {
		return nil, fmt.Errorf("device not specified")
	}
	if in.Pod == "" {
		return nil, fmt.Errorf("pod not specified")
	}
	if in.Memory <= 0 {
		return nil, fmt.Errorf("memory value is invalid")
	}

	var device *Device
	var pod *Pod
	var ok bool

	if device, ok = dm.devices[in.Device]; !ok {
		return nil, fmt.Errorf("device %s not registered", in.Device)
	}

	if pod, ok = device.Pods[in.Pod]; !ok {
		return nil, fmt.Errorf("pod %s not registered", in.Pod)
	}

	device.mut.Lock()
	defer device.mut.Unlock()

	if device.MemoryUsed+in.Memory > device.MemoryTotal || pod.MemoryUsed+in.Memory > pod.MemoryLimit {
		return nil, fmt.Errorf("OOM: memory limit exceeded")
	}

	device.MemoryUsed += in.Memory
	pod.MemoryUsed += in.Memory

	return &pb.GetMemoryQuotaReply{}, nil
}

func (dm *DeviceManager) ReturnMemoryQuota(ctx context.Context, in *pb.ReturnMemoryQuotaRequest) (*pb.ReturnMemoryQuotaReply, error) {
	log.Printf("Received: ReturnMemoryQuota for device %s: %d", in.Device, in.Memory)

	if in.Device == "" {
		return nil, fmt.Errorf("device not specified")
	}
	if in.Pod == "" {
		return nil, fmt.Errorf("pod not specified")
	}
	if in.Memory <= 0 {
		return nil, fmt.Errorf("memory value is invalid")
	}

	var device *Device
	var pod *Pod
	var ok bool

	if device, ok = dm.devices[in.Device]; !ok {
		return nil, fmt.Errorf("device %s not registered", in.Device)
	}

	if pod, ok = device.Pods[in.Pod]; !ok {
		return nil, fmt.Errorf("pod %s not registered", in.Pod)
	}

	device.mut.Lock()
	defer device.mut.Unlock()

	device.MemoryUsed -= in.Memory
	pod.MemoryUsed -= in.Memory

	log.Println(device.MemoryUsed)

	return &pb.ReturnMemoryQuotaReply{}, nil
}

func (dm *DeviceManager) RegisterDevice(ctx context.Context, in *pb.RegisterDeviceRequest) (*pb.RegisterDeviceReply, error) {
	log.Printf("Received: RegisterDevice for device %s", in.Device)

	if _, ok := dm.devices[in.Device]; ok {
		return nil, fmt.Errorf("device already registered")
	}

	dm.devices[in.Device] = &Device{
		mut:         &sync.Mutex{},
		Id:          in.Device,
		MemoryTotal: in.Memory,
		MemoryUsed:  0,
		Pods:        map[string]*Pod{},
	}

	dm.schedulerPerDevice[in.Device] = dm.sf.StartScheduler(in.Device)

	return &pb.RegisterDeviceReply{}, nil
}

func (dm *DeviceManager) RegisterPodQuota(ctx context.Context, in *pb.RegisterPodQuotaRequest) (*pb.RegisterPodQuotaReply, error) {
	log.Printf("Received: RegisterPod for device %s and pod %s", in.Device, in.Pod)

	if in.Device == "" {
		return nil, fmt.Errorf("device not specified")
	}
	if in.Pod == "" {
		return nil, fmt.Errorf("pod not specified")
	}

	var device *Device
	var ok bool

	if device, ok = dm.devices[in.Device]; !ok {
		return nil, fmt.Errorf("device %s not registered", in.Device)
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

	err := dm.schedulerPerDevice[in.Device].AllocatePodQuota(
		&scheduler.PodQuota{
			PodId: in.Pod, Requests: in.Requests, Limit: in.Limit,
		},
	)
	if err != nil {
		return nil, err
	}

	device.Pods[in.Pod] = &Pod{
		Id:          in.Pod,
		MemoryQuota: in.Memory,
		MemoryLimit: uint64(in.Memory * float64(device.MemoryTotal)),
		MemoryUsed:  0,
	}

	return &pb.RegisterPodQuotaReply{}, nil
}
