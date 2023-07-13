package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	pb "github.com/zbsss/device-manager/generated"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func worker(grpc pb.DeviceManagerClient, wg *sync.WaitGroup, deviceId, clientId string) {
	defer wg.Done()

	workTimeMin := 50
	workTimeMax := 300

	inbetweenTimeMin := 50
	inbetweenTimeMax := 100

	ctx := context.Background()

	logPrefix := fmt.Sprintf("[%s/%s]", deviceId, clientId)

	for {
		log.Printf("%s Getting memory quota", logPrefix)
		_, err := grpc.GetMemoryQuota(ctx, &pb.GetMemoryQuotaRequest{
			Device: deviceId,
			Pod:    clientId,
			Memory: 100,
		})
		if err != nil {
			log.Fatalf("%s could not get memory quota: %v", logPrefix, err)
		}

		log.Printf("%s Getting token", logPrefix)
		token, err := grpc.GetToken(ctx, &pb.GetTokenRequest{
			Device: deviceId,
			Pod:    clientId,
		})
		if err != nil {
			log.Fatalf("%s could not get token: %v", logPrefix, err)
		}

		log.Printf("%s Got token: %v", logPrefix, token.ExpiresAt)

		// Simulate work
		time.Sleep(time.Duration(rand.Intn(workTimeMax-workTimeMin+1)+workTimeMin) * time.Millisecond)

		log.Printf("%s Returning token", logPrefix)
		_, err = grpc.ReturnToken(ctx, &pb.ReturnTokenRequest{
			Device: deviceId,
			Pod:    clientId,
		})
		if err != nil {
			log.Fatalf("%s could not return token: %v", logPrefix, err)
		}

		log.Printf("%s Returning memory quota", logPrefix)
		_, err = grpc.ReturnMemoryQuota(ctx, &pb.ReturnMemoryQuotaRequest{
			Device: deviceId,
			Pod:    clientId,
			Memory: 100,
		})
		if err != nil {
			log.Fatalf("%s could not return memory quota: %v", logPrefix, err)
		}

		time.Sleep(time.Duration(rand.Intn(inbetweenTimeMax-inbetweenTimeMin+1)+inbetweenTimeMin) * time.Millisecond)
	}
}

var devices = []*pb.RegisterDeviceRequest{
	{
		Device: "dev-0",
		Memory: 1024,
	},
}

var clients = []*pb.RegisterPodQuotaRequest{
	{
		Device: "dev-0",
		Pod:    "client-0",

		Requests: 0.5,
		Limit:    0.5,
		Memory:   0.5,
	},
	{
		Device: "dev-0",
		Pod:    "client-1",

		Requests: 0.25,
		Limit:    0.25,
		Memory:   0.25,
	},
	{
		Device: "dev-0",
		Pod:    "client-2",

		Requests: 0.25,
		Limit:    0.5,
		Memory:   0.25,
	},
}

func setupDeviceManager(grpc pb.DeviceManagerClient) {
	ctx := context.Background()

	for _, device := range devices {
		_, err := grpc.RegisterDevice(ctx, device)
		if err != nil {
			log.Fatalf("could not register device: %v", err)
		}
	}

	for _, client := range clients {
		_, err := grpc.RegisterPodQuota(ctx, client)
		if err != nil {
			log.Fatalf("could not register client: %v", err)
		}
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())

	addr := "127.0.0.1:50051"

	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	grpc := pb.NewDeviceManagerClient(conn)

	setupDeviceManager(grpc)

	var wg sync.WaitGroup

	for i := 0; i < len(clients); i++ {
		log.Println("Starting worker: ", i)

		wg.Add(1)

		go worker(
			grpc,
			&wg,
			"dev-0",
			fmt.Sprintf("client-%d", i),
		)
	}
	wg.Wait()
}
