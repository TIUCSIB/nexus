package node

import (
	"context"
	"encoding/json"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/status"
)

// ---------------------------------------------------------------------------
// JSON codec for gRPC
// ---------------------------------------------------------------------------

type jsonCodec struct{}

func (jsonCodec) Marshal(v any) ([]byte, error)   { return json.Marshal(v) }
func (jsonCodec) Unmarshal(data []byte, v any) error { return json.Unmarshal(data, v) }
func (jsonCodec) Name() string                     { return "json" }

func init() {
	encoding.RegisterCodec(jsonCodec{})
}

// ---------------------------------------------------------------------------
// NodeServiceServer
// ---------------------------------------------------------------------------

// NodeServiceServer is the interface that gRPC servers must implement.
type NodeServiceServer interface {
	Register(context.Context, *RegisterRequest) (*RegisterResponse, error)
	Heartbeat(context.Context, *HeartbeatRequest) (*HeartbeatResponse, error)
	GetConfig(context.Context, *GetConfigRequest) (*GetConfigResponse, error)
	ReportTraffic(context.Context, *TrafficReport) (*TrafficReportResponse, error)
}

// UnimplementedNodeServiceServer can be embedded for forward-compatibility.
type UnimplementedNodeServiceServer struct{}

func (UnimplementedNodeServiceServer) Register(context.Context, *RegisterRequest) (*RegisterResponse, error) {
	return nil, status.Error(codes.Unimplemented, "Register is not implemented")
}

func (UnimplementedNodeServiceServer) Heartbeat(context.Context, *HeartbeatRequest) (*HeartbeatResponse, error) {
	return nil, status.Error(codes.Unimplemented, "Heartbeat is not implemented")
}

func (UnimplementedNodeServiceServer) GetConfig(context.Context, *GetConfigRequest) (*GetConfigResponse, error) {
	return nil, status.Error(codes.Unimplemented, "GetConfig is not implemented")
}

func (UnimplementedNodeServiceServer) ReportTraffic(context.Context, *TrafficReport) (*TrafficReportResponse, error) {
	return nil, status.Error(codes.Unimplemented, "ReportTraffic is not implemented")
}

// ---------------------------------------------------------------------------
// gRPC Service Descriptor & Registration
// ---------------------------------------------------------------------------

const serviceName = "nexus.node.NodeService"

// NodeService_ServiceDesc is the grpc.ServiceDesc for NodeService.
var NodeService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: serviceName,
	HandlerType: (*NodeServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Register",
			Handler: func(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
				in := new(RegisterRequest)
				if err := dec(in); err != nil {
					return nil, err
				}
				if interceptor == nil {
					return srv.(NodeServiceServer).Register(ctx, in)
				}
				info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/" + serviceName + "/Register"}
				handler := func(ctx context.Context, req any) (any, error) {
					return srv.(NodeServiceServer).Register(ctx, req.(*RegisterRequest))
				}
				return interceptor(ctx, in, info, handler)
			},
		},
		{
			MethodName: "Heartbeat",
			Handler: func(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
				in := new(HeartbeatRequest)
				if err := dec(in); err != nil {
					return nil, err
				}
				if interceptor == nil {
					return srv.(NodeServiceServer).Heartbeat(ctx, in)
				}
				info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/" + serviceName + "/Heartbeat"}
				handler := func(ctx context.Context, req any) (any, error) {
					return srv.(NodeServiceServer).Heartbeat(ctx, req.(*HeartbeatRequest))
				}
				return interceptor(ctx, in, info, handler)
			},
		},
		{
			MethodName: "GetConfig",
			Handler: func(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
				in := new(GetConfigRequest)
				if err := dec(in); err != nil {
					return nil, err
				}
				if interceptor == nil {
					return srv.(NodeServiceServer).GetConfig(ctx, in)
				}
				info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/" + serviceName + "/GetConfig"}
				handler := func(ctx context.Context, req any) (any, error) {
					return srv.(NodeServiceServer).GetConfig(ctx, req.(*GetConfigRequest))
				}
				return interceptor(ctx, in, info, handler)
			},
		},
		{
			MethodName: "ReportTraffic",
			Handler: func(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
				in := new(TrafficReport)
				if err := dec(in); err != nil {
					return nil, err
				}
				if interceptor == nil {
					return srv.(NodeServiceServer).ReportTraffic(ctx, in)
				}
				info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/" + serviceName + "/ReportTraffic"}
				handler := func(ctx context.Context, req any) (any, error) {
					return srv.(NodeServiceServer).ReportTraffic(ctx, req.(*TrafficReport))
				}
				return interceptor(ctx, in, info, handler)
			},
		},
	},
}

// RegisterNodeServiceServer registers the NodeServiceServer on the given gRPC server.
func RegisterNodeServiceServer(s *grpc.Server, srv NodeServiceServer) {
	s.RegisterService(&NodeService_ServiceDesc, srv)
}

// ---------------------------------------------------------------------------
// Client
// ---------------------------------------------------------------------------

// NodeServiceClient is the client API for NodeService.
type NodeServiceClient interface {
	Register(ctx context.Context, in *RegisterRequest, opts ...grpc.CallOption) (*RegisterResponse, error)
	Heartbeat(ctx context.Context, in *HeartbeatRequest, opts ...grpc.CallOption) (*HeartbeatResponse, error)
	GetConfig(ctx context.Context, in *GetConfigRequest, opts ...grpc.CallOption) (*GetConfigResponse, error)
	ReportTraffic(ctx context.Context, in *TrafficReport, opts ...grpc.CallOption) (*TrafficReportResponse, error)
}

type nodeServiceClient struct {
	cc grpc.ClientConnInterface
}

// NewNodeServiceClient creates a client that communicates using JSON encoding.
func NewNodeServiceClient(cc grpc.ClientConnInterface) NodeServiceClient {
	return &nodeServiceClient{cc}
}

func (c *nodeServiceClient) Register(ctx context.Context, in *RegisterRequest, opts ...grpc.CallOption) (*RegisterResponse, error) {
	out := new(RegisterResponse)
	if err := c.cc.Invoke(ctx, "/"+serviceName+"/Register", in, out, opts...); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *nodeServiceClient) Heartbeat(ctx context.Context, in *HeartbeatRequest, opts ...grpc.CallOption) (*HeartbeatResponse, error) {
	out := new(HeartbeatResponse)
	if err := c.cc.Invoke(ctx, "/"+serviceName+"/Heartbeat", in, out, opts...); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *nodeServiceClient) GetConfig(ctx context.Context, in *GetConfigRequest, opts ...grpc.CallOption) (*GetConfigResponse, error) {
	out := new(GetConfigResponse)
	if err := c.cc.Invoke(ctx, "/"+serviceName+"/GetConfig", in, out, opts...); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *nodeServiceClient) ReportTraffic(ctx context.Context, in *TrafficReport, opts ...grpc.CallOption) (*TrafficReportResponse, error) {
	out := new(TrafficReportResponse)
	if err := c.cc.Invoke(ctx, "/"+serviceName+"/ReportTraffic", in, out, opts...); err != nil {
		return nil, err
	}
	return out, nil
}

// DialNodeService dials the target with the JSON codec pre-configured.
func DialNodeService(target string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	opts = append(opts,
		grpc.WithDefaultCallOptions(grpc.ForceCodec(jsonCodec{})),
	)
	return grpc.NewClient(target, opts...)
}
