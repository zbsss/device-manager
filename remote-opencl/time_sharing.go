package opencl

import (
	"log"
	"os"

	pb "github.com/zbsss/device-manager/generated"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var ClientId = os.Getenv("CLIENT_ID")
var DeviceId = os.Getenv("DEVICE_ID")

var Scheduler = initScheduler()

type scheduler = pb.DeviceManagerClient

func initScheduler() scheduler {
	var addr = os.Getenv("DEVICE_MANAGER_SERVICE_ADDR")
	if addr == "" {
		addr = "127.0.0.1:50051"
	}

	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	// defer conn.Close()

	return pb.NewDeviceManagerClient(conn)
}
