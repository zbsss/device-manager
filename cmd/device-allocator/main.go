package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"sync"
	"time"

	pb "github.com/zbsss/device-manager/generated"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	port := "50051"

	addr := os.Getenv("HOST_IP")
	deviceId := os.Getenv("DEVICE_ID")

	// TODO: add retries in case the Device Manager was restarted

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

	_, err = grpc.RegisterDevice(ctx, &pb.RegisterDeviceRequest{
		Vendor:   "example.com",
		Model:    "mydev",
		DeviceId: deviceId,
		MemoryB:  100000000,
	})
	if err != nil {
		log.Printf("could not register device: %v", err)
	}

	free, err := grpc.GetAvailableDevices(ctx, &pb.GetAvailableDevicesRequest{
		Vendor: "example.com",
		Model:  "mydev",
	})
	if err != nil {
		log.Printf("could not get available devices: %v", err)
	} else {
		log.Printf("available devices: %v", free.Free)
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	wg.Wait()
}
