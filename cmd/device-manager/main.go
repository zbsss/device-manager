package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	pb "github.com/zbsss/device-manager/generated"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	port          = flag.Int("port", 50051, "The server port")
	tokenLifetime = flag.Int("token", 250, "Lifetime of token in milliseconds")
	windowSize    = flag.Int("windowSize", 10000, "Window size in milliseconds")
	verbose       = flag.Bool("verbose", true, "Window size in milliseconds")
)

type server struct {
	pb.UnimplementedDeviceManagerServer
}

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

var devices = map[string]*Device{
	"1": {
		mut:         &sync.Mutex{},
		Id:          "1",
		MemoryTotal: 1000000000,
		MemoryUsed:  0,
		Pods: map[string]*Pod{
			"pod-1": {
				Id:          "pod-1",
				MemoryQuota: 0.5,
				MemoryLimit: 500000000,
				MemoryUsed:  0,
			},
		},
	},
}

func (s *server) GetToken(ctx context.Context, in *pb.GetTokenRequest) (*pb.GetTokenReply, error) {
	log.Printf("Received: GetToken for device %s", in.Device)

	time.Sleep(5 * time.Second)

	log.Printf("Sending token")

	expiresAt := time.Now().Add(time.Duration(*tokenLifetime) * time.Millisecond).UnixNano()

	return &pb.GetTokenReply{ExpiresAt: expiresAt}, nil
}

func (s *server) ReturnToken(ctx context.Context, in *pb.ReturnTokenRequest) (*pb.ReturnTokenReply, error) {
	log.Printf("Received: ReturnToken")

	return &pb.ReturnTokenReply{}, nil
}

func (s *server) GetMemoryQuota(ctx context.Context, in *pb.GetMemoryQuotaRequest) (*pb.GetMemoryQuotaReply, error) {
	log.Printf("Received: GetMemoryQuota for device %s: %d", in.Device, in.Memory)

	podId := "pod-1"

	var device *Device
	var pod *Pod
	var ok bool

	if device, ok = devices[in.Device]; !ok {
		return nil, fmt.Errorf("device not registered")
	}

	if pod, ok = device.Pods[podId]; !ok {
		return nil, fmt.Errorf("pod not registered")
	}

	device.mut.Lock()

	if device.MemoryUsed+in.Memory > device.MemoryTotal || pod.MemoryUsed+in.Memory > pod.MemoryLimit {
		device.mut.Unlock()
		return nil, fmt.Errorf("OOM: memory limit exceeded")
	}

	device.MemoryUsed += in.Memory
	pod.MemoryUsed += in.Memory

	device.mut.Unlock()

	log.Println(device.MemoryUsed)

	return &pb.GetMemoryQuotaReply{}, nil
}

func (s *server) ReturnMemoryQuota(ctx context.Context, in *pb.ReturnMemoryQuotaRequest) (*pb.ReturnMemoryQuotaReply, error) {
	log.Printf("Received: ReturnMemoryQuota for device %s: %d", in.Device, in.Memory)

	podId := "pod-1"

	var device *Device
	var pod *Pod
	var ok bool

	if device, ok = devices[in.Device]; !ok {
		return nil, fmt.Errorf("device not registered")
	}

	if pod, ok = device.Pods[podId]; !ok {
		return nil, fmt.Errorf("pod not registered")
	}

	device.mut.Lock()

	device.MemoryUsed -= in.Memory
	pod.MemoryUsed -= in.Memory

	device.mut.Unlock()

	log.Println(device.MemoryUsed)

	return &pb.ReturnMemoryQuotaReply{}, nil
}

func (s *server) RegisterDevice(ctx context.Context, in *pb.RegisterDeviceRequest) (*pb.RegisterDeviceReply, error) {
	log.Printf("Received: RegisterDevice for device %s", in.Device)

	if _, ok := devices[in.Device]; ok {
		return nil, fmt.Errorf("device already registered")
	}

	devices[in.Device] = &Device{
		mut:         &sync.Mutex{},
		Id:          in.Device,
		MemoryTotal: in.Memory,
		MemoryUsed:  0,
		Pods:        map[string]*Pod{},
	}

	return &pb.RegisterDeviceReply{}, nil
}

func (s *server) RegisterPodQuota(ctx context.Context, in *pb.RegisterPodQuotaRequest) (*pb.RegisterPodQuotaReply, error) {
	log.Printf("Received: RegisterPod for device %s and pod %s", in.Device, in.Pod)

	var device *Device
	var ok bool

	if device, ok = devices[in.Device]; !ok {
		return nil, fmt.Errorf("device not registered")
	}

	device.mut.Lock()

	device.Pods[in.Pod] = &Pod{
		Id:          in.Pod,
		MemoryQuota: in.Memory,
		MemoryLimit: uint64(in.Memory * float64(device.MemoryTotal)),
		MemoryUsed:  0,
	}

	device.mut.Unlock()

	return &pb.RegisterPodQuotaReply{}, nil
}

func stateLoggerDaemon() {
	for {
		var sb strings.Builder
		sb.WriteString("\n===Current state===")
		for _, device := range devices {
			sb.WriteString(fmt.Sprintf("\nDevice %s: %d/%d", device.Id, device.MemoryUsed, device.MemoryTotal))

			for _, pod := range device.Pods {
				sb.WriteString(fmt.Sprintf("\n\tPod %s: %d/%d", pod.Id, pod.MemoryUsed, pod.MemoryLimit))
			}
		}
		sb.WriteString("\n===================\n")
		log.Println(sb.String())

		time.Sleep(10 * time.Second)
	}
}

func main() {
	flag.Parse()

	if *verbose {
		go stateLoggerDaemon()
	}

	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	reflection.Register(s)
	pb.RegisterDeviceManagerServer(s, &server{})

	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
