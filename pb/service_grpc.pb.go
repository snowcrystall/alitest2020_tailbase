// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package pb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion6

// AgentServiceClient is the client API for AgentService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type AgentServiceClient interface {
	NotifyTargetTraceids(ctx context.Context, opts ...grpc.CallOption) (AgentService_NotifyTargetTraceidsClient, error)
	NotifyAllFilterOver(ctx context.Context, in *Req, opts ...grpc.CallOption) (*Reply, error)
}

type agentServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewAgentServiceClient(cc grpc.ClientConnInterface) AgentServiceClient {
	return &agentServiceClient{cc}
}

func (c *agentServiceClient) NotifyTargetTraceids(ctx context.Context, opts ...grpc.CallOption) (AgentService_NotifyTargetTraceidsClient, error) {
	stream, err := c.cc.NewStream(ctx, &_AgentService_serviceDesc.Streams[0], "/pb.AgentService/NotifyTargetTraceids", opts...)
	if err != nil {
		return nil, err
	}
	x := &agentServiceNotifyTargetTraceidsClient{stream}
	return x, nil
}

type AgentService_NotifyTargetTraceidsClient interface {
	Send(*TargetInfo) error
	CloseAndRecv() (*Reply, error)
	grpc.ClientStream
}

type agentServiceNotifyTargetTraceidsClient struct {
	grpc.ClientStream
}

func (x *agentServiceNotifyTargetTraceidsClient) Send(m *TargetInfo) error {
	return x.ClientStream.SendMsg(m)
}

func (x *agentServiceNotifyTargetTraceidsClient) CloseAndRecv() (*Reply, error) {
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	m := new(Reply)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *agentServiceClient) NotifyAllFilterOver(ctx context.Context, in *Req, opts ...grpc.CallOption) (*Reply, error) {
	out := new(Reply)
	err := c.cc.Invoke(ctx, "/pb.AgentService/NotifyAllFilterOver", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// AgentServiceServer is the server API for AgentService service.
// All implementations must embed UnimplementedAgentServiceServer
// for forward compatibility
type AgentServiceServer interface {
	NotifyTargetTraceids(AgentService_NotifyTargetTraceidsServer) error
	NotifyAllFilterOver(context.Context, *Req) (*Reply, error)
	mustEmbedUnimplementedAgentServiceServer()
}

// UnimplementedAgentServiceServer must be embedded to have forward compatible implementations.
type UnimplementedAgentServiceServer struct {
}

func (*UnimplementedAgentServiceServer) NotifyTargetTraceids(AgentService_NotifyTargetTraceidsServer) error {
	return status.Errorf(codes.Unimplemented, "method NotifyTargetTraceids not implemented")
}
func (*UnimplementedAgentServiceServer) NotifyAllFilterOver(context.Context, *Req) (*Reply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method NotifyAllFilterOver not implemented")
}
func (*UnimplementedAgentServiceServer) mustEmbedUnimplementedAgentServiceServer() {}

func RegisterAgentServiceServer(s *grpc.Server, srv AgentServiceServer) {
	s.RegisterService(&_AgentService_serviceDesc, srv)
}

func _AgentService_NotifyTargetTraceids_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(AgentServiceServer).NotifyTargetTraceids(&agentServiceNotifyTargetTraceidsServer{stream})
}

type AgentService_NotifyTargetTraceidsServer interface {
	SendAndClose(*Reply) error
	Recv() (*TargetInfo, error)
	grpc.ServerStream
}

type agentServiceNotifyTargetTraceidsServer struct {
	grpc.ServerStream
}

func (x *agentServiceNotifyTargetTraceidsServer) SendAndClose(m *Reply) error {
	return x.ServerStream.SendMsg(m)
}

func (x *agentServiceNotifyTargetTraceidsServer) Recv() (*TargetInfo, error) {
	m := new(TargetInfo)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _AgentService_NotifyAllFilterOver_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Req)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AgentServiceServer).NotifyAllFilterOver(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.AgentService/NotifyAllFilterOver",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AgentServiceServer).NotifyAllFilterOver(ctx, req.(*Req))
	}
	return interceptor(ctx, in, info, handler)
}

var _AgentService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "pb.AgentService",
	HandlerType: (*AgentServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "NotifyAllFilterOver",
			Handler:    _AgentService_NotifyAllFilterOver_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "NotifyTargetTraceids",
			Handler:       _AgentService_NotifyTargetTraceids_Handler,
			ClientStreams: true,
		},
	},
	Metadata: "pb/service.proto",
}

// ProcessServiceClient is the client API for ProcessService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ProcessServiceClient interface {
	SetTargetTraceid(ctx context.Context, in *TraceidRequest, opts ...grpc.CallOption) (*Reply, error)
	SendTraceData(ctx context.Context, opts ...grpc.CallOption) (ProcessService_SendTraceDataClient, error)
	SendTargetIds(ctx context.Context, opts ...grpc.CallOption) (ProcessService_SendTargetIdsClient, error)
	NotifyFilterOver(ctx context.Context, in *Addr, opts ...grpc.CallOption) (*Reply, error)
	NotifySendOver(ctx context.Context, in *Addr, opts ...grpc.CallOption) (*Reply, error)
}

type processServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewProcessServiceClient(cc grpc.ClientConnInterface) ProcessServiceClient {
	return &processServiceClient{cc}
}

func (c *processServiceClient) SetTargetTraceid(ctx context.Context, in *TraceidRequest, opts ...grpc.CallOption) (*Reply, error) {
	out := new(Reply)
	err := c.cc.Invoke(ctx, "/pb.ProcessService/SetTargetTraceid", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *processServiceClient) SendTraceData(ctx context.Context, opts ...grpc.CallOption) (ProcessService_SendTraceDataClient, error) {
	stream, err := c.cc.NewStream(ctx, &_ProcessService_serviceDesc.Streams[0], "/pb.ProcessService/SendTraceData", opts...)
	if err != nil {
		return nil, err
	}
	x := &processServiceSendTraceDataClient{stream}
	return x, nil
}

type ProcessService_SendTraceDataClient interface {
	Send(*TraceData) error
	CloseAndRecv() (*Reply, error)
	grpc.ClientStream
}

type processServiceSendTraceDataClient struct {
	grpc.ClientStream
}

func (x *processServiceSendTraceDataClient) Send(m *TraceData) error {
	return x.ClientStream.SendMsg(m)
}

func (x *processServiceSendTraceDataClient) CloseAndRecv() (*Reply, error) {
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	m := new(Reply)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *processServiceClient) SendTargetIds(ctx context.Context, opts ...grpc.CallOption) (ProcessService_SendTargetIdsClient, error) {
	stream, err := c.cc.NewStream(ctx, &_ProcessService_serviceDesc.Streams[1], "/pb.ProcessService/SendTargetIds", opts...)
	if err != nil {
		return nil, err
	}
	x := &processServiceSendTargetIdsClient{stream}
	return x, nil
}

type ProcessService_SendTargetIdsClient interface {
	Send(*TargetInfo) error
	CloseAndRecv() (*Reply, error)
	grpc.ClientStream
}

type processServiceSendTargetIdsClient struct {
	grpc.ClientStream
}

func (x *processServiceSendTargetIdsClient) Send(m *TargetInfo) error {
	return x.ClientStream.SendMsg(m)
}

func (x *processServiceSendTargetIdsClient) CloseAndRecv() (*Reply, error) {
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	m := new(Reply)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *processServiceClient) NotifyFilterOver(ctx context.Context, in *Addr, opts ...grpc.CallOption) (*Reply, error) {
	out := new(Reply)
	err := c.cc.Invoke(ctx, "/pb.ProcessService/NotifyFilterOver", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *processServiceClient) NotifySendOver(ctx context.Context, in *Addr, opts ...grpc.CallOption) (*Reply, error) {
	out := new(Reply)
	err := c.cc.Invoke(ctx, "/pb.ProcessService/NotifySendOver", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ProcessServiceServer is the server API for ProcessService service.
// All implementations must embed UnimplementedProcessServiceServer
// for forward compatibility
type ProcessServiceServer interface {
	SetTargetTraceid(context.Context, *TraceidRequest) (*Reply, error)
	SendTraceData(ProcessService_SendTraceDataServer) error
	SendTargetIds(ProcessService_SendTargetIdsServer) error
	NotifyFilterOver(context.Context, *Addr) (*Reply, error)
	NotifySendOver(context.Context, *Addr) (*Reply, error)
	mustEmbedUnimplementedProcessServiceServer()
}

// UnimplementedProcessServiceServer must be embedded to have forward compatible implementations.
type UnimplementedProcessServiceServer struct {
}

func (*UnimplementedProcessServiceServer) SetTargetTraceid(context.Context, *TraceidRequest) (*Reply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetTargetTraceid not implemented")
}
func (*UnimplementedProcessServiceServer) SendTraceData(ProcessService_SendTraceDataServer) error {
	return status.Errorf(codes.Unimplemented, "method SendTraceData not implemented")
}
func (*UnimplementedProcessServiceServer) SendTargetIds(ProcessService_SendTargetIdsServer) error {
	return status.Errorf(codes.Unimplemented, "method SendTargetIds not implemented")
}
func (*UnimplementedProcessServiceServer) NotifyFilterOver(context.Context, *Addr) (*Reply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method NotifyFilterOver not implemented")
}
func (*UnimplementedProcessServiceServer) NotifySendOver(context.Context, *Addr) (*Reply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method NotifySendOver not implemented")
}
func (*UnimplementedProcessServiceServer) mustEmbedUnimplementedProcessServiceServer() {}

func RegisterProcessServiceServer(s *grpc.Server, srv ProcessServiceServer) {
	s.RegisterService(&_ProcessService_serviceDesc, srv)
}

func _ProcessService_SetTargetTraceid_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TraceidRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ProcessServiceServer).SetTargetTraceid(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.ProcessService/SetTargetTraceid",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ProcessServiceServer).SetTargetTraceid(ctx, req.(*TraceidRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ProcessService_SendTraceData_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(ProcessServiceServer).SendTraceData(&processServiceSendTraceDataServer{stream})
}

type ProcessService_SendTraceDataServer interface {
	SendAndClose(*Reply) error
	Recv() (*TraceData, error)
	grpc.ServerStream
}

type processServiceSendTraceDataServer struct {
	grpc.ServerStream
}

func (x *processServiceSendTraceDataServer) SendAndClose(m *Reply) error {
	return x.ServerStream.SendMsg(m)
}

func (x *processServiceSendTraceDataServer) Recv() (*TraceData, error) {
	m := new(TraceData)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _ProcessService_SendTargetIds_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(ProcessServiceServer).SendTargetIds(&processServiceSendTargetIdsServer{stream})
}

type ProcessService_SendTargetIdsServer interface {
	SendAndClose(*Reply) error
	Recv() (*TargetInfo, error)
	grpc.ServerStream
}

type processServiceSendTargetIdsServer struct {
	grpc.ServerStream
}

func (x *processServiceSendTargetIdsServer) SendAndClose(m *Reply) error {
	return x.ServerStream.SendMsg(m)
}

func (x *processServiceSendTargetIdsServer) Recv() (*TargetInfo, error) {
	m := new(TargetInfo)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _ProcessService_NotifyFilterOver_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Addr)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ProcessServiceServer).NotifyFilterOver(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.ProcessService/NotifyFilterOver",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ProcessServiceServer).NotifyFilterOver(ctx, req.(*Addr))
	}
	return interceptor(ctx, in, info, handler)
}

func _ProcessService_NotifySendOver_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Addr)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ProcessServiceServer).NotifySendOver(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.ProcessService/NotifySendOver",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ProcessServiceServer).NotifySendOver(ctx, req.(*Addr))
	}
	return interceptor(ctx, in, info, handler)
}

var _ProcessService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "pb.ProcessService",
	HandlerType: (*ProcessServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SetTargetTraceid",
			Handler:    _ProcessService_SetTargetTraceid_Handler,
		},
		{
			MethodName: "NotifyFilterOver",
			Handler:    _ProcessService_NotifyFilterOver_Handler,
		},
		{
			MethodName: "NotifySendOver",
			Handler:    _ProcessService_NotifySendOver_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "SendTraceData",
			Handler:       _ProcessService_SendTraceData_Handler,
			ClientStreams: true,
		},
		{
			StreamName:    "SendTargetIds",
			Handler:       _ProcessService_SendTargetIds_Handler,
			ClientStreams: true,
		},
	},
	Metadata: "pb/service.proto",
}