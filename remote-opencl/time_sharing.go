package opencl

import (
	"fmt"
	"log"
	"os"

	pb "github.com/zbsss/device-manager/pkg/devicemanager"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const port = "50051"

var ClientId = os.Getenv("CLIENT_ID")
var DeviceId = os.Getenv("DEVICE_ID")

var Scheduler = initScheduler()

type scheduler = pb.DeviceManagerClient

func initScheduler() scheduler {
	var addr = os.Getenv("HOST_IP")
	if addr == "" {
		addr = "127.0.0.1"
	}

	conn, err := grpc.Dial(
		fmt.Sprintf("%s:%s", addr, port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	// defer conn.Close()

	return pb.NewDeviceManagerClient(conn)
}
