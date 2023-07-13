package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"time"

	pb "github.com/zbsss/device-manager/generated"
	"github.com/zbsss/device-manager/internal/devicemanager"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	port          = flag.Int("port", 50051, "The server port")
	tokenLifetime = flag.Int("token-life", 200, "Lifetime of token in milliseconds")
	windowSize    = flag.Int("windowSize", 10000, "Window size in milliseconds")
)

var windowDuration = time.Duration(*windowSize) * time.Millisecond
var tokenDuration = time.Duration(*tokenLifetime) * time.Millisecond

func main() {
	flag.Parse()

	dm := devicemanager.NewTestDeviceManager(windowDuration, tokenDuration)

	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	reflection.Register(s)
	pb.RegisterDeviceManagerServer(s, dm)

	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
