// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.1.0
// - protoc             v3.17.3
// source: chameleon/smelter/v1/crawl/session/service.proto

package session

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// SessionManagerClient is the client API for SessionManager service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type SessionManagerClient interface {
	// GetCookies
	GetCookies(ctx context.Context, in *GetCookiesRequest, opts ...grpc.CallOption) (*GetCookiesResponse, error)
	// SetCookies
	SetCookies(ctx context.Context, in *SetCookiesRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	// ClearCookies
	ClearCookies(ctx context.Context, in *ClearCookiesRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
}

type sessionManagerClient struct {
	cc grpc.ClientConnInterface
}

func NewSessionManagerClient(cc grpc.ClientConnInterface) SessionManagerClient {
	return &sessionManagerClient{cc}
}

func (c *sessionManagerClient) GetCookies(ctx context.Context, in *GetCookiesRequest, opts ...grpc.CallOption) (*GetCookiesResponse, error) {
	out := new(GetCookiesResponse)
	err := c.cc.Invoke(ctx, "/chameleon.smelter.v1.crawl.session.SessionManager/GetCookies", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *sessionManagerClient) SetCookies(ctx context.Context, in *SetCookiesRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/chameleon.smelter.v1.crawl.session.SessionManager/SetCookies", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *sessionManagerClient) ClearCookies(ctx context.Context, in *ClearCookiesRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/chameleon.smelter.v1.crawl.session.SessionManager/ClearCookies", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// SessionManagerServer is the server API for SessionManager service.
// All implementations must embed UnimplementedSessionManagerServer
// for forward compatibility
type SessionManagerServer interface {
	// GetCookies
	GetCookies(context.Context, *GetCookiesRequest) (*GetCookiesResponse, error)
	// SetCookies
	SetCookies(context.Context, *SetCookiesRequest) (*emptypb.Empty, error)
	// ClearCookies
	ClearCookies(context.Context, *ClearCookiesRequest) (*emptypb.Empty, error)
	mustEmbedUnimplementedSessionManagerServer()
}

// UnimplementedSessionManagerServer must be embedded to have forward compatible implementations.
type UnimplementedSessionManagerServer struct {
}

func (UnimplementedSessionManagerServer) GetCookies(context.Context, *GetCookiesRequest) (*GetCookiesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetCookies not implemented")
}
func (UnimplementedSessionManagerServer) SetCookies(context.Context, *SetCookiesRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetCookies not implemented")
}
func (UnimplementedSessionManagerServer) ClearCookies(context.Context, *ClearCookiesRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ClearCookies not implemented")
}
func (UnimplementedSessionManagerServer) mustEmbedUnimplementedSessionManagerServer() {}

// UnsafeSessionManagerServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to SessionManagerServer will
// result in compilation errors.
type UnsafeSessionManagerServer interface {
	mustEmbedUnimplementedSessionManagerServer()
}

func RegisterSessionManagerServer(s grpc.ServiceRegistrar, srv SessionManagerServer) {
	s.RegisterService(&SessionManager_ServiceDesc, srv)
}

func _SessionManager_GetCookies_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetCookiesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SessionManagerServer).GetCookies(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/chameleon.smelter.v1.crawl.session.SessionManager/GetCookies",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SessionManagerServer).GetCookies(ctx, req.(*GetCookiesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _SessionManager_SetCookies_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SetCookiesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SessionManagerServer).SetCookies(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/chameleon.smelter.v1.crawl.session.SessionManager/SetCookies",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SessionManagerServer).SetCookies(ctx, req.(*SetCookiesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _SessionManager_ClearCookies_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ClearCookiesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SessionManagerServer).ClearCookies(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/chameleon.smelter.v1.crawl.session.SessionManager/ClearCookies",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SessionManagerServer).ClearCookies(ctx, req.(*ClearCookiesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// SessionManager_ServiceDesc is the grpc.ServiceDesc for SessionManager service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var SessionManager_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "chameleon.smelter.v1.crawl.session.SessionManager",
	HandlerType: (*SessionManagerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetCookies",
			Handler:    _SessionManager_GetCookies_Handler,
		},
		{
			MethodName: "SetCookies",
			Handler:    _SessionManager_SetCookies_Handler,
		},
		{
			MethodName: "ClearCookies",
			Handler:    _SessionManager_ClearCookies_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "chameleon/smelter/v1/crawl/session/service.proto",
}
