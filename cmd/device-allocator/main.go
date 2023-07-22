package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	pb "github.com/zbsss/device-manager/pkg/devicemanager"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	port := "50051"
	addr := os.Getenv("HOST_IP")
	allocatorPodId := os.Getenv("ALLOCATOR_POD_ID")
	deviceId := os.Getenv("DEVICE_ID")
	vendor := os.Getenv("VENDOR")
	model := os.Getenv("MODEL")
	memory := os.Getenv("MEMORY")

	// parse memory into uint64
	memoryB, err := strconv.ParseUint(memory, 10, 64)
	if err != nil {
		log.Fatalf("could not parse memory: %v", err)
	}

	conn, err := grpc.Dial(
		fmt.Sprintf("%s:%s", addr, port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	defer conn.Close()

	grpc := pb.NewDeviceManagerClient(conn)
	ctx := context.Background()

	for {
		_, err = grpc.RegisterDevice(ctx, &pb.RegisterDeviceRequest{
			AllocatorPodId: allocatorPodId,
			Vendor:         vendor,
			Model:          model,
			DeviceId:       deviceId,
			MemoryB:        memoryB,
		})
		if err != nil {
			log.Printf("could not register device: %v", err)
		}

		free, err := grpc.GetAvailableDevices(ctx, &pb.GetAvailableDevicesRequest{
			Vendor: vendor,
			Model:  model,
		})
		if err != nil {
			log.Printf("could not get available devices: %v", err)
		} else {
			log.Printf("available devices: %v", free.Free)
		}

		time.Sleep(60 * time.Second)
	}
}
