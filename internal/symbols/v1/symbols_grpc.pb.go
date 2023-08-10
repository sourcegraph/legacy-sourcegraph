// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             (unknown)
// source: symbols.proto

package v1

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
	SymbolsService_Search_FullMethodName         = "/symbols.v1.SymbolsService/Search"
	SymbolsService_LocalCodeIntel_FullMethodName = "/symbols.v1.SymbolsService/LocalCodeIntel"
	SymbolsService_ListLanguages_FullMethodName  = "/symbols.v1.SymbolsService/ListLanguages"
	SymbolsService_SymbolInfo_FullMethodName     = "/symbols.v1.SymbolsService/SymbolInfo"
	SymbolsService_Healthz_FullMethodName        = "/symbols.v1.SymbolsService/Healthz"
)

// SymbolsServiceClient is the client API for SymbolsService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type SymbolsServiceClient interface {
	Search(ctx context.Context, in *SearchRequest, opts ...grpc.CallOption) (*SearchResponse, error)
	LocalCodeIntel(ctx context.Context, in *LocalCodeIntelRequest, opts ...grpc.CallOption) (SymbolsService_LocalCodeIntelClient, error)
	ListLanguages(ctx context.Context, in *ListLanguagesRequest, opts ...grpc.CallOption) (*ListLanguagesResponse, error)
	SymbolInfo(ctx context.Context, in *SymbolInfoRequest, opts ...grpc.CallOption) (*SymbolInfoResponse, error)
	Healthz(ctx context.Context, in *HealthzRequest, opts ...grpc.CallOption) (*HealthzResponse, error)
}

type symbolsServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewSymbolsServiceClient(cc grpc.ClientConnInterface) SymbolsServiceClient {
	return &symbolsServiceClient{cc}
}

func (c *symbolsServiceClient) Search(ctx context.Context, in *SearchRequest, opts ...grpc.CallOption) (*SearchResponse, error) {
	out := new(SearchResponse)
	err := c.cc.Invoke(ctx, SymbolsService_Search_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *symbolsServiceClient) LocalCodeIntel(ctx context.Context, in *LocalCodeIntelRequest, opts ...grpc.CallOption) (SymbolsService_LocalCodeIntelClient, error) {
	stream, err := c.cc.NewStream(ctx, &SymbolsService_ServiceDesc.Streams[0], SymbolsService_LocalCodeIntel_FullMethodName, opts...)
	if err != nil {
		return nil, err
	}
	x := &symbolsServiceLocalCodeIntelClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type SymbolsService_LocalCodeIntelClient interface {
	Recv() (*LocalCodeIntelResponse, error)
	grpc.ClientStream
}

type symbolsServiceLocalCodeIntelClient struct {
	grpc.ClientStream
}

func (x *symbolsServiceLocalCodeIntelClient) Recv() (*LocalCodeIntelResponse, error) {
	m := new(LocalCodeIntelResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *symbolsServiceClient) ListLanguages(ctx context.Context, in *ListLanguagesRequest, opts ...grpc.CallOption) (*ListLanguagesResponse, error) {
	out := new(ListLanguagesResponse)
	err := c.cc.Invoke(ctx, SymbolsService_ListLanguages_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *symbolsServiceClient) SymbolInfo(ctx context.Context, in *SymbolInfoRequest, opts ...grpc.CallOption) (*SymbolInfoResponse, error) {
	out := new(SymbolInfoResponse)
	err := c.cc.Invoke(ctx, SymbolsService_SymbolInfo_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *symbolsServiceClient) Healthz(ctx context.Context, in *HealthzRequest, opts ...grpc.CallOption) (*HealthzResponse, error) {
	out := new(HealthzResponse)
	err := c.cc.Invoke(ctx, SymbolsService_Healthz_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// SymbolsServiceServer is the server API for SymbolsService service.
// All implementations must embed UnimplementedSymbolsServiceServer
// for forward compatibility
type SymbolsServiceServer interface {
	Search(context.Context, *SearchRequest) (*SearchResponse, error)
	LocalCodeIntel(*LocalCodeIntelRequest, SymbolsService_LocalCodeIntelServer) error
	ListLanguages(context.Context, *ListLanguagesRequest) (*ListLanguagesResponse, error)
	SymbolInfo(context.Context, *SymbolInfoRequest) (*SymbolInfoResponse, error)
	Healthz(context.Context, *HealthzRequest) (*HealthzResponse, error)
	mustEmbedUnimplementedSymbolsServiceServer()
}

// UnimplementedSymbolsServiceServer must be embedded to have forward compatible implementations.
type UnimplementedSymbolsServiceServer struct {
}

func (UnimplementedSymbolsServiceServer) Search(context.Context, *SearchRequest) (*SearchResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Search not implemented")
}
func (UnimplementedSymbolsServiceServer) LocalCodeIntel(*LocalCodeIntelRequest, SymbolsService_LocalCodeIntelServer) error {
	return status.Errorf(codes.Unimplemented, "method LocalCodeIntel not implemented")
}
func (UnimplementedSymbolsServiceServer) ListLanguages(context.Context, *ListLanguagesRequest) (*ListLanguagesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListLanguages not implemented")
}
func (UnimplementedSymbolsServiceServer) SymbolInfo(context.Context, *SymbolInfoRequest) (*SymbolInfoResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SymbolInfo not implemented")
}
func (UnimplementedSymbolsServiceServer) Healthz(context.Context, *HealthzRequest) (*HealthzResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Healthz not implemented")
}
func (UnimplementedSymbolsServiceServer) mustEmbedUnimplementedSymbolsServiceServer() {}

// UnsafeSymbolsServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to SymbolsServiceServer will
// result in compilation errors.
type UnsafeSymbolsServiceServer interface {
	mustEmbedUnimplementedSymbolsServiceServer()
}

func RegisterSymbolsServiceServer(s grpc.ServiceRegistrar, srv SymbolsServiceServer) {
	s.RegisterService(&SymbolsService_ServiceDesc, srv)
}

func _SymbolsService_Search_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SearchRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SymbolsServiceServer).Search(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: SymbolsService_Search_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SymbolsServiceServer).Search(ctx, req.(*SearchRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _SymbolsService_LocalCodeIntel_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(LocalCodeIntelRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(SymbolsServiceServer).LocalCodeIntel(m, &symbolsServiceLocalCodeIntelServer{stream})
}

type SymbolsService_LocalCodeIntelServer interface {
	Send(*LocalCodeIntelResponse) error
	grpc.ServerStream
}

type symbolsServiceLocalCodeIntelServer struct {
	grpc.ServerStream
}

func (x *symbolsServiceLocalCodeIntelServer) Send(m *LocalCodeIntelResponse) error {
	return x.ServerStream.SendMsg(m)
}

func _SymbolsService_ListLanguages_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListLanguagesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SymbolsServiceServer).ListLanguages(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: SymbolsService_ListLanguages_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SymbolsServiceServer).ListLanguages(ctx, req.(*ListLanguagesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _SymbolsService_SymbolInfo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SymbolInfoRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SymbolsServiceServer).SymbolInfo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: SymbolsService_SymbolInfo_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SymbolsServiceServer).SymbolInfo(ctx, req.(*SymbolInfoRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _SymbolsService_Healthz_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HealthzRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SymbolsServiceServer).Healthz(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: SymbolsService_Healthz_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SymbolsServiceServer).Healthz(ctx, req.(*HealthzRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// SymbolsService_ServiceDesc is the grpc.ServiceDesc for SymbolsService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var SymbolsService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "symbols.v1.SymbolsService",
	HandlerType: (*SymbolsServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Search",
			Handler:    _SymbolsService_Search_Handler,
		},
		{
			MethodName: "ListLanguages",
			Handler:    _SymbolsService_ListLanguages_Handler,
		},
		{
			MethodName: "SymbolInfo",
			Handler:    _SymbolsService_SymbolInfo_Handler,
		},
		{
			MethodName: "Healthz",
			Handler:    _SymbolsService_Healthz_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "LocalCodeIntel",
			Handler:       _SymbolsService_LocalCodeIntel_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "symbols.proto",
}
