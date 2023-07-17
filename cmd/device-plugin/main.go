package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"path"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

const (
	resourceName = "example.com/mydev"
	endpoint     = "mydev.sock"
)

var (
	serverSock = path.Join(pluginapi.DevicePluginPath, endpoint)
)

type DummyDevicePlugin struct {
	pluginapi.UnimplementedDevicePluginServer
	devices []*pluginapi.Device
	server  *grpc.Server
}

func NewDummyDevicePlugin() *DummyDevicePlugin {
	return &DummyDevicePlugin{
		devices: []*pluginapi.Device{
			{ID: "device1", Health: pluginapi.Healthy},
			{ID: "device2", Health: pluginapi.Healthy},
		},
		server: grpc.NewServer([]grpc.ServerOption{}...),
	}
}

func main() {
	log.Printf("Starting dummy device plugin: %s", resourceName)

	plugin := NewDummyDevicePlugin()
	err := cleanup()
	if err != nil {
		log.Fatalf("Could not cleanup: %s", err)
	}

	sock, err := net.Listen("unix", serverSock)
	if err != nil {
		log.Fatalf("Could not listen on %s: %s", serverSock, err)
	}

	pluginapi.RegisterDevicePluginServer(plugin.server, plugin)

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		err := plugin.server.Serve(sock)
		if err != nil {
			log.Fatalf("Could not start device plugin server: %s", err)
		}
	}()

	err = dialDevicePlugin()
	if err != nil {
		plugin.Stop()
		log.Fatalf("Could not dial device plugin: %s", err)
	}

	err = plugin.Register(pluginapi.KubeletSocket, resourceName)
	if err != nil {
		plugin.Stop()
		log.Fatalf("Could not register device plugin: %s", err)
	}

	wg.Wait()
}

func dialDevicePlugin() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, serverSock,
		grpc.WithAuthority("localhost"),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
			return (&net.Dialer{}).DialContext(ctx, "unix", addr)
		}),
	)

	if err != nil {
		return fmt.Errorf("could not connect to %s: %s", serverSock, err)
	}

	defer conn.Close()
	return nil
}
