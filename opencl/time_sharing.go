package opencl

import (
	"log"

	pb "github.com/zbsss/device-manager/generated"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// device-manager-service:80
const addr = "127.0.0.1:50051"

type scheduler = pb.DeviceManagerClient

func initScheduler() scheduler {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	// defer conn.Close()

	return pb.NewDeviceManagerClient(conn)
}

var Scheduler = initScheduler()
