package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

func (p *DummyDevicePlugin) Stop() error {
	if p.server == nil {
		return nil
	}
	p.server.Stop()
	p.server = nil
	return cleanup()
}

func (p *DummyDevicePlugin) Register(masterSock string, resourceName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, masterSock,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
			return (&net.Dialer{}).DialContext(ctx, "unix", addr)
		}),
	)

	if err != nil {
		return fmt.Errorf("could not connect to %s: %s", masterSock, err)
	}
	defer conn.Close()

	client := pluginapi.NewRegistrationClient(conn)
	req := &pluginapi.RegisterRequest{
		Version:      pluginapi.Version,
		Endpoint:     endpoint,
		ResourceName: resourceName,
	}

	_, err = client.Register(context.Background(), req)
	if err != nil {
		return fmt.Errorf("could not register to kubelet service: %s", err)
	}

	return nil
}

func (p *DummyDevicePlugin) GetDevicePluginOptions(context.Context, *pluginapi.Empty) (*pluginapi.DevicePluginOptions, error) {
	return &pluginapi.DevicePluginOptions{
		PreStartRequired: false,
	}, nil
}

func (p *DummyDevicePlugin) Allocate(ctx context.Context, reqs *pluginapi.AllocateRequest) (*pluginapi.AllocateResponse, error) {
	log.Printf("Allocate request: %+v", reqs)

	resps := new(pluginapi.AllocateResponse)
	for _, req := range reqs.ContainerRequests {
		resp := new(pluginapi.ContainerAllocateResponse)
		for _, id := range req.DevicesIDs {
			dev := findDeviceByID(p, id)
			if dev == nil {
				return nil, nil
			}

			resp.Envs = map[string]string{
				"DEVICE_ID": id,
				"MEMORY":    "512",
			}
		}
		resps.ContainerResponses = append(resps.ContainerResponses, resp)
	}
	return resps, nil
}

func (p *DummyDevicePlugin) ListAndWatch(e *pluginapi.Empty, s pluginapi.DevicePlugin_ListAndWatchServer) error {
	s.Send(&pluginapi.ListAndWatchResponse{Devices: p.devices})
	tick := time.NewTicker(time.Second)
	defer tick.Stop()

	for {
		select {
		case <-tick.C:
			s.Send(&pluginapi.ListAndWatchResponse{Devices: p.devices})
		case <-s.Context().Done():
			return s.Context().Err()
		}
	}
}

func (p *DummyDevicePlugin) PreStartContainer(ctx context.Context, req *pluginapi.PreStartContainerRequest) (*pluginapi.PreStartContainerResponse, error) {
	return &pluginapi.PreStartContainerResponse{}, nil
}

func cleanup() error {
	if err := os.Remove(serverSock); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func findDeviceByID(p *DummyDevicePlugin, id string) *pluginapi.Device {
	for _, dev := range p.devices {
		if dev.ID == id {
			return dev
		}
	}
	return nil
}
