// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v3.12.4
// source: device-manager.proto

package generated

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

const (
	DeviceManager_GetToken_FullMethodName          = "/device_manager.DeviceManager/GetToken"
	DeviceManager_ReturnToken_FullMethodName       = "/device_manager.DeviceManager/ReturnToken"
	DeviceManager_GetMemoryQuota_FullMethodName    = "/device_manager.DeviceManager/GetMemoryQuota"
	DeviceManager_ReturnMemoryQuota_FullMethodName = "/device_manager.DeviceManager/ReturnMemoryQuota"
)

// DeviceManagerClient is the client API for DeviceManager service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type DeviceManagerClient interface {
	GetToken(ctx context.Context, in *GetTokenRequest, opts ...grpc.CallOption) (*GetTokenReply, error)
	ReturnToken(ctx context.Context, in *ReturnTokenRequest, opts ...grpc.CallOption) (*ReturnTokenReply, error)
	GetMemoryQuota(ctx context.Context, in *GetMemoryQuotaRequest, opts ...grpc.CallOption) (*GetMemoryQuotaReply, error)
	ReturnMemoryQuota(ctx context.Context, in *ReturnMemoryQuotaRequest, opts ...grpc.CallOption) (*ReturnMemoryQuotaReply, error)
}

type deviceManagerClient struct {
	cc grpc.ClientConnInterface
}

func NewDeviceManagerClient(cc grpc.ClientConnInterface) DeviceManagerClient {
	return &deviceManagerClient{cc}
}

func (c *deviceManagerClient) GetToken(ctx context.Context, in *GetTokenRequest, opts ...grpc.CallOption) (*GetTokenReply, error) {
	out := new(GetTokenReply)
	err := c.cc.Invoke(ctx, DeviceManager_GetToken_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *deviceManagerClient) ReturnToken(ctx context.Context, in *ReturnTokenRequest, opts ...grpc.CallOption) (*ReturnTokenReply, error) {
	out := new(ReturnTokenReply)
	err := c.cc.Invoke(ctx, DeviceManager_ReturnToken_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *deviceManagerClient) GetMemoryQuota(ctx context.Context, in *GetMemoryQuotaRequest, opts ...grpc.CallOption) (*GetMemoryQuotaReply, error) {
	out := new(GetMemoryQuotaReply)
	err := c.cc.Invoke(ctx, DeviceManager_GetMemoryQuota_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *deviceManagerClient) ReturnMemoryQuota(ctx context.Context, in *ReturnMemoryQuotaRequest, opts ...grpc.CallOption) (*ReturnMemoryQuotaReply, error) {
	out := new(ReturnMemoryQuotaReply)
	err := c.cc.Invoke(ctx, DeviceManager_ReturnMemoryQuota_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// DeviceManagerServer is the server API for DeviceManager service.
// All implementations must embed UnimplementedDeviceManagerServer
// for forward compatibility
type DeviceManagerServer interface {
	GetToken(context.Context, *GetTokenRequest) (*GetTokenReply, error)
	ReturnToken(context.Context, *ReturnTokenRequest) (*ReturnTokenReply, error)
	GetMemoryQuota(context.Context, *GetMemoryQuotaRequest) (*GetMemoryQuotaReply, error)
	ReturnMemoryQuota(context.Context, *ReturnMemoryQuotaRequest) (*ReturnMemoryQuotaReply, error)
	mustEmbedUnimplementedDeviceManagerServer()
}

// UnimplementedDeviceManagerServer must be embedded to have forward compatible implementations.
type UnimplementedDeviceManagerServer struct {
}

func (UnimplementedDeviceManagerServer) GetToken(context.Context, *GetTokenRequest) (*GetTokenReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetToken not implemented")
}
func (UnimplementedDeviceManagerServer) ReturnToken(context.Context, *ReturnTokenRequest) (*ReturnTokenReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ReturnToken not implemented")
}
func (UnimplementedDeviceManagerServer) GetMemoryQuota(context.Context, *GetMemoryQuotaRequest) (*GetMemoryQuotaReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetMemoryQuota not implemented")
}
func (UnimplementedDeviceManagerServer) ReturnMemoryQuota(context.Context, *ReturnMemoryQuotaRequest) (*ReturnMemoryQuotaReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ReturnMemoryQuota not implemented")
}
func (UnimplementedDeviceManagerServer) mustEmbedUnimplementedDeviceManagerServer() {}

// UnsafeDeviceManagerServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to DeviceManagerServer will
// result in compilation errors.
type UnsafeDeviceManagerServer interface {
	mustEmbedUnimplementedDeviceManagerServer()
}

func RegisterDeviceManagerServer(s grpc.ServiceRegistrar, srv DeviceManagerServer) {
	s.RegisterService(&DeviceManager_ServiceDesc, srv)
}

func _DeviceManager_GetToken_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetTokenRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DeviceManagerServer).GetToken(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: DeviceManager_GetToken_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DeviceManagerServer).GetToken(ctx, req.(*GetTokenRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _DeviceManager_ReturnToken_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ReturnTokenRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DeviceManagerServer).ReturnToken(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: DeviceManager_ReturnToken_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DeviceManagerServer).ReturnToken(ctx, req.(*ReturnTokenRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _DeviceManager_GetMemoryQuota_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetMemoryQuotaRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DeviceManagerServer).GetMemoryQuota(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: DeviceManager_GetMemoryQuota_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DeviceManagerServer).GetMemoryQuota(ctx, req.(*GetMemoryQuotaRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _DeviceManager_ReturnMemoryQuota_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ReturnMemoryQuotaRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DeviceManagerServer).ReturnMemoryQuota(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: DeviceManager_ReturnMemoryQuota_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DeviceManagerServer).ReturnMemoryQuota(ctx, req.(*ReturnMemoryQuotaRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// DeviceManager_ServiceDesc is the grpc.ServiceDesc for DeviceManager service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var DeviceManager_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "device_manager.DeviceManager",
	HandlerType: (*DeviceManagerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetToken",
			Handler:    _DeviceManager_GetToken_Handler,
		},
		{
			MethodName: "ReturnToken",
			Handler:    _DeviceManager_ReturnToken_Handler,
		},
		{
			MethodName: "GetMemoryQuota",
			Handler:    _DeviceManager_GetMemoryQuota_Handler,
		},
		{
			MethodName: "ReturnMemoryQuota",
			Handler:    _DeviceManager_ReturnMemoryQuota_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "device-manager.proto",
}