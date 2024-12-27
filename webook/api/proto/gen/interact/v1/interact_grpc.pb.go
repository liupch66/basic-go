// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             (unknown)
// source: interact/v1/interact.proto

package interactv1

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	InteractService_IncrReadCnt_FullMethodName = "/interact.v1.InteractService/IncrReadCnt"
	InteractService_Like_FullMethodName        = "/interact.v1.InteractService/Like"
	InteractService_CancelLike_FullMethodName  = "/interact.v1.InteractService/CancelLike"
	InteractService_Collect_FullMethodName     = "/interact.v1.InteractService/Collect"
	InteractService_Get_FullMethodName         = "/interact.v1.InteractService/Get"
	InteractService_GetByIds_FullMethodName    = "/interact.v1.InteractService/GetByIds"
)

// InteractServiceClient is the client API for InteractService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type InteractServiceClient interface {
	IncrReadCnt(ctx context.Context, in *IncrReadCntRequest, opts ...grpc.CallOption) (*IncrReadCntResponse, error)
	Like(ctx context.Context, in *LikeRequest, opts ...grpc.CallOption) (*LikeResponse, error)
	CancelLike(ctx context.Context, in *CancelLikeRequest, opts ...grpc.CallOption) (*CancelLikeResponse, error)
	Collect(ctx context.Context, in *CollectRequest, opts ...grpc.CallOption) (*CollectResponse, error)
	Get(ctx context.Context, in *GetRequest, opts ...grpc.CallOption) (*GetResponse, error)
	GetByIds(ctx context.Context, in *GetByIdsRequest, opts ...grpc.CallOption) (*GetByIdsResponse, error)
}

type interactServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewInteractServiceClient(cc grpc.ClientConnInterface) InteractServiceClient {
	return &interactServiceClient{cc}
}

func (c *interactServiceClient) IncrReadCnt(ctx context.Context, in *IncrReadCntRequest, opts ...grpc.CallOption) (*IncrReadCntResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(IncrReadCntResponse)
	err := c.cc.Invoke(ctx, InteractService_IncrReadCnt_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *interactServiceClient) Like(ctx context.Context, in *LikeRequest, opts ...grpc.CallOption) (*LikeResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(LikeResponse)
	err := c.cc.Invoke(ctx, InteractService_Like_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *interactServiceClient) CancelLike(ctx context.Context, in *CancelLikeRequest, opts ...grpc.CallOption) (*CancelLikeResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(CancelLikeResponse)
	err := c.cc.Invoke(ctx, InteractService_CancelLike_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *interactServiceClient) Collect(ctx context.Context, in *CollectRequest, opts ...grpc.CallOption) (*CollectResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(CollectResponse)
	err := c.cc.Invoke(ctx, InteractService_Collect_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *interactServiceClient) Get(ctx context.Context, in *GetRequest, opts ...grpc.CallOption) (*GetResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetResponse)
	err := c.cc.Invoke(ctx, InteractService_Get_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *interactServiceClient) GetByIds(ctx context.Context, in *GetByIdsRequest, opts ...grpc.CallOption) (*GetByIdsResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetByIdsResponse)
	err := c.cc.Invoke(ctx, InteractService_GetByIds_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// InteractServiceServer is the server API for InteractService service.
// All implementations must embed UnimplementedInteractServiceServer
// for forward compatibility.
type InteractServiceServer interface {
	IncrReadCnt(context.Context, *IncrReadCntRequest) (*IncrReadCntResponse, error)
	Like(context.Context, *LikeRequest) (*LikeResponse, error)
	CancelLike(context.Context, *CancelLikeRequest) (*CancelLikeResponse, error)
	Collect(context.Context, *CollectRequest) (*CollectResponse, error)
	Get(context.Context, *GetRequest) (*GetResponse, error)
	GetByIds(context.Context, *GetByIdsRequest) (*GetByIdsResponse, error)
	mustEmbedUnimplementedInteractServiceServer()
}

// UnimplementedInteractServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedInteractServiceServer struct{}

func (UnimplementedInteractServiceServer) IncrReadCnt(context.Context, *IncrReadCntRequest) (*IncrReadCntResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method IncrReadCnt not implemented")
}
func (UnimplementedInteractServiceServer) Like(context.Context, *LikeRequest) (*LikeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Like not implemented")
}
func (UnimplementedInteractServiceServer) CancelLike(context.Context, *CancelLikeRequest) (*CancelLikeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CancelLike not implemented")
}
func (UnimplementedInteractServiceServer) Collect(context.Context, *CollectRequest) (*CollectResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Collect not implemented")
}
func (UnimplementedInteractServiceServer) Get(context.Context, *GetRequest) (*GetResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Get not implemented")
}
func (UnimplementedInteractServiceServer) GetByIds(context.Context, *GetByIdsRequest) (*GetByIdsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetByIds not implemented")
}
func (UnimplementedInteractServiceServer) mustEmbedUnimplementedInteractServiceServer() {}
func (UnimplementedInteractServiceServer) testEmbeddedByValue()                         {}

// UnsafeInteractServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to InteractServiceServer will
// result in compilation errors.
type UnsafeInteractServiceServer interface {
	mustEmbedUnimplementedInteractServiceServer()
}

func RegisterInteractServiceServer(s grpc.ServiceRegistrar, srv InteractServiceServer) {
	// If the following call pancis, it indicates UnimplementedInteractServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&InteractService_ServiceDesc, srv)
}

func _InteractService_IncrReadCnt_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(IncrReadCntRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(InteractServiceServer).IncrReadCnt(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: InteractService_IncrReadCnt_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(InteractServiceServer).IncrReadCnt(ctx, req.(*IncrReadCntRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _InteractService_Like_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LikeRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(InteractServiceServer).Like(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: InteractService_Like_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(InteractServiceServer).Like(ctx, req.(*LikeRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _InteractService_CancelLike_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CancelLikeRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(InteractServiceServer).CancelLike(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: InteractService_CancelLike_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(InteractServiceServer).CancelLike(ctx, req.(*CancelLikeRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _InteractService_Collect_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CollectRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(InteractServiceServer).Collect(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: InteractService_Collect_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(InteractServiceServer).Collect(ctx, req.(*CollectRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _InteractService_Get_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(InteractServiceServer).Get(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: InteractService_Get_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(InteractServiceServer).Get(ctx, req.(*GetRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _InteractService_GetByIds_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetByIdsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(InteractServiceServer).GetByIds(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: InteractService_GetByIds_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(InteractServiceServer).GetByIds(ctx, req.(*GetByIdsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// InteractService_ServiceDesc is the grpc.ServiceDesc for InteractService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var InteractService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "interact.v1.InteractService",
	HandlerType: (*InteractServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "IncrReadCnt",
			Handler:    _InteractService_IncrReadCnt_Handler,
		},
		{
			MethodName: "Like",
			Handler:    _InteractService_Like_Handler,
		},
		{
			MethodName: "CancelLike",
			Handler:    _InteractService_CancelLike_Handler,
		},
		{
			MethodName: "Collect",
			Handler:    _InteractService_Collect_Handler,
		},
		{
			MethodName: "Get",
			Handler:    _InteractService_Get_Handler,
		},
		{
			MethodName: "GetByIds",
			Handler:    _InteractService_GetByIds_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "interact/v1/interact.proto",
}
