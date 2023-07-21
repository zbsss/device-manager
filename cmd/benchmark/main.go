package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	pb "github.com/zbsss/device-manager/generated"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func waitRandom(min, max int) {
	time.Sleep(time.Duration(rand.Intn(max-min+1)+min) * time.Second)
}

func main() {
	rand.Seed(time.Now().UnixNano())

	port := "50051"

	clientId := os.Getenv("CLIENT_ID")
	deviceId := os.Getenv("DEVICE_ID")
	addr := os.Getenv("HOST_IP")

	conn, err := grpc.Dial(
		fmt.Sprintf("%s:%s", addr, port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	defer conn.Close()

	grpc := pb.NewDeviceManagerClient(conn)

	workTimeMin := 20
	workTimeMax := 25

	ctx := context.Background()

	for {
		_, err := grpc.AllocateMemory(ctx, &pb.AllocateMemoryRequest{
			DeviceId: deviceId,
			PodId:    clientId,
			MemoryB:  128,
		})
		if err != nil {
			log.Fatalf("could not get memory quota: %v", err)
		}

		_, err = grpc.GetToken(ctx, &pb.GetTokenRequest{
			DeviceId: deviceId,
			PodId:    clientId,
		})
		if err != nil {
			log.Fatalf("could not get token: %v", err)
		}

		// Simulate work
		waitRandom(workTimeMin, workTimeMax)

		_, err = grpc.ReturnToken(ctx, &pb.ReturnTokenRequest{
			DeviceId: deviceId,
			PodId:    clientId,
		})
		if err != nil {
			log.Fatalf("could not return token: %v", err)
		}

		_, err = grpc.FreeMemory(ctx, &pb.FreeMemoryRequest{
			DeviceId: deviceId,
			PodId:    clientId,
			MemoryB:  128,
		})
		if err != nil {
			log.Fatalf("could not return memory quota: %v", err)
		}
	}
}
