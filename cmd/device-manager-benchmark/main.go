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

var devices = []*pb.RegisterDeviceRequest{
	{
		DeviceId: "dev-0",
		MemoryB:  1024,
	},
}

var clients = []*pb.ReservePodQuotaRequest{
	{
		DeviceId: "dev-0",
		PodId:    "client-0",

		Requests: 0.5,
		Limit:    0.75,
		Memory:   0.25,
	},
	{
		DeviceId: "dev-0",
		PodId:    "client-1",

		Requests: 0.25,
		Limit:    0.25,
		Memory:   0.25,
	},
	{
		DeviceId: "dev-0",
		PodId:    "client-2",

		Requests: 0.25,
		Limit:    0.25,
		Memory:   0.25,
	},
}

func waitRandom(min, max int) {
	time.Sleep(time.Duration(rand.Intn(max-min+1)+min) * time.Millisecond)
}

func worker(grpc pb.DeviceManagerClient, wg *sync.WaitGroup, deviceId, clientId string) {
	defer wg.Done()

	workTimeMin := 200
	workTimeMax := 225

	// inbetweenTimeMin := 0
	// inbetweenTimeMax := 100

	ctx := context.Background()

	logPrefix := fmt.Sprintf("[%s/%s] ", deviceId, clientId)
	infoLogger := log.New(os.Stdout, logPrefix, log.LstdFlags)

	for {
		infoLogger.Printf("Getting memory quota")
		_, err := grpc.AllocateMemory(ctx, &pb.AllocateMemoryRequest{
			DeviceId: deviceId,
			PodId:    clientId,
			MemoryB:  100,
		})
		if err != nil {
			infoLogger.Fatalf("could not get memory quota: %v", err)
		}

		infoLogger.Printf("Getting token")
		token, err := grpc.GetToken(ctx, &pb.GetTokenRequest{
			DeviceId: deviceId,
			PodId:    clientId,
		})
		if err != nil {
			log.Fatalf("%s could not get token: %v", logPrefix, err)
		}

		infoLogger.Printf("Got token: %v", token.ExpiresAt)

		// Simulate work
		waitRandom(workTimeMin, workTimeMax)

		infoLogger.Printf("Returning token")
		_, err = grpc.ReturnToken(ctx, &pb.ReturnTokenRequest{
			DeviceId: deviceId,
			PodId:    clientId,
		})
		if err != nil {
			infoLogger.Fatalf("could not return token: %v", err)
		}

		infoLogger.Printf("Returning memory quota")
		_, err = grpc.FreeMemory(ctx, &pb.FreeMemoryRequest{
			DeviceId: deviceId,
			PodId:    clientId,
			MemoryB:  100,
		})
		if err != nil {
			infoLogger.Fatalf("could not return memory quota: %v", err)
		}

		// waitRandom(inbetweenTimeMin, inbetweenTimeMax)
	}
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
		_, err := grpc.ReservePodQuota(ctx, client)
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
