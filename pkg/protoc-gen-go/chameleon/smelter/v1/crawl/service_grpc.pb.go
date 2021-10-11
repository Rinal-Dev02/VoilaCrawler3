// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package crawl

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	anypb "google.golang.org/protobuf/types/known/anypb"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion7

// CrawlerNodeClient is the client API for CrawlerNode service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type CrawlerNodeClient interface {
	// Version
	Version(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*VersionResponse, error)
	// CrawlerOptions
	CrawlerOptions(ctx context.Context, in *CrawlerOptionsRequest, opts ...grpc.CallOption) (*CrawlerOptionsResponse, error)
	// AllowedDomains
	AllowedDomains(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*AllowedDomainsResponse, error)
	// CanonicalUrl
	CanonicalUrl(ctx context.Context, in *CanonicalUrlRequest, opts ...grpc.CallOption) (*CanonicalUrlResponse, error)
	// Parse
	Parse(ctx context.Context, in *Request, opts ...grpc.CallOption) (CrawlerNode_ParseClient, error)
	// Call used to get categories, brands
	// NOTE: this api may be non-real time.
	Call(ctx context.Context, in *CallRequest, opts ...grpc.CallOption) (*CallResponse, error)
}

type crawlerNodeClient struct {
	cc grpc.ClientConnInterface
}

func NewCrawlerNodeClient(cc grpc.ClientConnInterface) CrawlerNodeClient {
	return &crawlerNodeClient{cc}
}

func (c *crawlerNodeClient) Version(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*VersionResponse, error) {
	out := new(VersionResponse)
	err := c.cc.Invoke(ctx, "/chameleon.smelter.v1.crawl.CrawlerNode/Version", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *crawlerNodeClient) CrawlerOptions(ctx context.Context, in *CrawlerOptionsRequest, opts ...grpc.CallOption) (*CrawlerOptionsResponse, error) {
	out := new(CrawlerOptionsResponse)
	err := c.cc.Invoke(ctx, "/chameleon.smelter.v1.crawl.CrawlerNode/CrawlerOptions", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *crawlerNodeClient) AllowedDomains(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*AllowedDomainsResponse, error) {
	out := new(AllowedDomainsResponse)
	err := c.cc.Invoke(ctx, "/chameleon.smelter.v1.crawl.CrawlerNode/AllowedDomains", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *crawlerNodeClient) CanonicalUrl(ctx context.Context, in *CanonicalUrlRequest, opts ...grpc.CallOption) (*CanonicalUrlResponse, error) {
	out := new(CanonicalUrlResponse)
	err := c.cc.Invoke(ctx, "/chameleon.smelter.v1.crawl.CrawlerNode/CanonicalUrl", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *crawlerNodeClient) Parse(ctx context.Context, in *Request, opts ...grpc.CallOption) (CrawlerNode_ParseClient, error) {
	stream, err := c.cc.NewStream(ctx, &_CrawlerNode_serviceDesc.Streams[0], "/chameleon.smelter.v1.crawl.CrawlerNode/Parse", opts...)
	if err != nil {
		return nil, err
	}
	x := &crawlerNodeParseClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type CrawlerNode_ParseClient interface {
	Recv() (*anypb.Any, error)
	grpc.ClientStream
}

type crawlerNodeParseClient struct {
	grpc.ClientStream
}

func (x *crawlerNodeParseClient) Recv() (*anypb.Any, error) {
	m := new(anypb.Any)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *crawlerNodeClient) Call(ctx context.Context, in *CallRequest, opts ...grpc.CallOption) (*CallResponse, error) {
	out := new(CallResponse)
	err := c.cc.Invoke(ctx, "/chameleon.smelter.v1.crawl.CrawlerNode/Call", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// CrawlerNodeServer is the server API for CrawlerNode service.
// All implementations must embed UnimplementedCrawlerNodeServer
// for forward compatibility
type CrawlerNodeServer interface {
	// Version
	Version(context.Context, *emptypb.Empty) (*VersionResponse, error)
	// CrawlerOptions
	CrawlerOptions(context.Context, *CrawlerOptionsRequest) (*CrawlerOptionsResponse, error)
	// AllowedDomains
	AllowedDomains(context.Context, *emptypb.Empty) (*AllowedDomainsResponse, error)
	// CanonicalUrl
	CanonicalUrl(context.Context, *CanonicalUrlRequest) (*CanonicalUrlResponse, error)
	// Parse
	Parse(*Request, CrawlerNode_ParseServer) error
	// Call used to get categories, brands
	// NOTE: this api may be non-real time.
	Call(context.Context, *CallRequest) (*CallResponse, error)
	mustEmbedUnimplementedCrawlerNodeServer()
}

// UnimplementedCrawlerNodeServer must be embedded to have forward compatible implementations.
type UnimplementedCrawlerNodeServer struct {
}

func (UnimplementedCrawlerNodeServer) Version(context.Context, *emptypb.Empty) (*VersionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Version not implemented")
}
func (UnimplementedCrawlerNodeServer) CrawlerOptions(context.Context, *CrawlerOptionsRequest) (*CrawlerOptionsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CrawlerOptions not implemented")
}
func (UnimplementedCrawlerNodeServer) AllowedDomains(context.Context, *emptypb.Empty) (*AllowedDomainsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AllowedDomains not implemented")
}
func (UnimplementedCrawlerNodeServer) CanonicalUrl(context.Context, *CanonicalUrlRequest) (*CanonicalUrlResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CanonicalUrl not implemented")
}
func (UnimplementedCrawlerNodeServer) Parse(*Request, CrawlerNode_ParseServer) error {
	return status.Errorf(codes.Unimplemented, "method Parse not implemented")
}
func (UnimplementedCrawlerNodeServer) Call(context.Context, *CallRequest) (*CallResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Call not implemented")
}
func (UnimplementedCrawlerNodeServer) mustEmbedUnimplementedCrawlerNodeServer() {}

// UnsafeCrawlerNodeServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to CrawlerNodeServer will
// result in compilation errors.
type UnsafeCrawlerNodeServer interface {
	mustEmbedUnimplementedCrawlerNodeServer()
}

func RegisterCrawlerNodeServer(s *grpc.Server, srv CrawlerNodeServer) {
	s.RegisterService(&_CrawlerNode_serviceDesc, srv)
}

func _CrawlerNode_Version_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CrawlerNodeServer).Version(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/chameleon.smelter.v1.crawl.CrawlerNode/Version",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CrawlerNodeServer).Version(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _CrawlerNode_CrawlerOptions_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CrawlerOptionsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CrawlerNodeServer).CrawlerOptions(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/chameleon.smelter.v1.crawl.CrawlerNode/CrawlerOptions",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CrawlerNodeServer).CrawlerOptions(ctx, req.(*CrawlerOptionsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CrawlerNode_AllowedDomains_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CrawlerNodeServer).AllowedDomains(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/chameleon.smelter.v1.crawl.CrawlerNode/AllowedDomains",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CrawlerNodeServer).AllowedDomains(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _CrawlerNode_CanonicalUrl_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CanonicalUrlRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CrawlerNodeServer).CanonicalUrl(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/chameleon.smelter.v1.crawl.CrawlerNode/CanonicalUrl",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CrawlerNodeServer).CanonicalUrl(ctx, req.(*CanonicalUrlRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CrawlerNode_Parse_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(Request)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(CrawlerNodeServer).Parse(m, &crawlerNodeParseServer{stream})
}

type CrawlerNode_ParseServer interface {
	Send(*anypb.Any) error
	grpc.ServerStream
}

type crawlerNodeParseServer struct {
	grpc.ServerStream
}

func (x *crawlerNodeParseServer) Send(m *anypb.Any) error {
	return x.ServerStream.SendMsg(m)
}

func _CrawlerNode_Call_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CallRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CrawlerNodeServer).Call(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/chameleon.smelter.v1.crawl.CrawlerNode/Call",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CrawlerNodeServer).Call(ctx, req.(*CallRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _CrawlerNode_serviceDesc = grpc.ServiceDesc{
	ServiceName: "chameleon.smelter.v1.crawl.CrawlerNode",
	HandlerType: (*CrawlerNodeServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Version",
			Handler:    _CrawlerNode_Version_Handler,
		},
		{
			MethodName: "CrawlerOptions",
			Handler:    _CrawlerNode_CrawlerOptions_Handler,
		},
		{
			MethodName: "AllowedDomains",
			Handler:    _CrawlerNode_AllowedDomains_Handler,
		},
		{
			MethodName: "CanonicalUrl",
			Handler:    _CrawlerNode_CanonicalUrl_Handler,
		},
		{
			MethodName: "Call",
			Handler:    _CrawlerNode_Call_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Parse",
			Handler:       _CrawlerNode_Parse_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "chameleon/smelter/v1/crawl/service.proto",
}

// CrawlerRegisterClient is the client API for CrawlerRegister service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type CrawlerRegisterClient interface {
	// Connect
	Connect(ctx context.Context, opts ...grpc.CallOption) (CrawlerRegister_ConnectClient, error)
}

type crawlerRegisterClient struct {
	cc grpc.ClientConnInterface
}

func NewCrawlerRegisterClient(cc grpc.ClientConnInterface) CrawlerRegisterClient {
	return &crawlerRegisterClient{cc}
}

func (c *crawlerRegisterClient) Connect(ctx context.Context, opts ...grpc.CallOption) (CrawlerRegister_ConnectClient, error) {
	stream, err := c.cc.NewStream(ctx, &_CrawlerRegister_serviceDesc.Streams[0], "/chameleon.smelter.v1.crawl.CrawlerRegister/Connect", opts...)
	if err != nil {
		return nil, err
	}
	x := &crawlerRegisterConnectClient{stream}
	return x, nil
}

type CrawlerRegister_ConnectClient interface {
	Send(*anypb.Any) error
	Recv() (*anypb.Any, error)
	grpc.ClientStream
}

type crawlerRegisterConnectClient struct {
	grpc.ClientStream
}

func (x *crawlerRegisterConnectClient) Send(m *anypb.Any) error {
	return x.ClientStream.SendMsg(m)
}

func (x *crawlerRegisterConnectClient) Recv() (*anypb.Any, error) {
	m := new(anypb.Any)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// CrawlerRegisterServer is the server API for CrawlerRegister service.
// All implementations must embed UnimplementedCrawlerRegisterServer
// for forward compatibility
type CrawlerRegisterServer interface {
	// Connect
	Connect(CrawlerRegister_ConnectServer) error
	mustEmbedUnimplementedCrawlerRegisterServer()
}

// UnimplementedCrawlerRegisterServer must be embedded to have forward compatible implementations.
type UnimplementedCrawlerRegisterServer struct {
}

func (UnimplementedCrawlerRegisterServer) Connect(CrawlerRegister_ConnectServer) error {
	return status.Errorf(codes.Unimplemented, "method Connect not implemented")
}
func (UnimplementedCrawlerRegisterServer) mustEmbedUnimplementedCrawlerRegisterServer() {}

// UnsafeCrawlerRegisterServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to CrawlerRegisterServer will
// result in compilation errors.
type UnsafeCrawlerRegisterServer interface {
	mustEmbedUnimplementedCrawlerRegisterServer()
}

func RegisterCrawlerRegisterServer(s *grpc.Server, srv CrawlerRegisterServer) {
	s.RegisterService(&_CrawlerRegister_serviceDesc, srv)
}

func _CrawlerRegister_Connect_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(CrawlerRegisterServer).Connect(&crawlerRegisterConnectServer{stream})
}

type CrawlerRegister_ConnectServer interface {
	Send(*anypb.Any) error
	Recv() (*anypb.Any, error)
	grpc.ServerStream
}

type crawlerRegisterConnectServer struct {
	grpc.ServerStream
}

func (x *crawlerRegisterConnectServer) Send(m *anypb.Any) error {
	return x.ServerStream.SendMsg(m)
}

func (x *crawlerRegisterConnectServer) Recv() (*anypb.Any, error) {
	m := new(anypb.Any)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

var _CrawlerRegister_serviceDesc = grpc.ServiceDesc{
	ServiceName: "chameleon.smelter.v1.crawl.CrawlerRegister",
	HandlerType: (*CrawlerRegisterServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Connect",
			Handler:       _CrawlerRegister_Connect_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "chameleon/smelter/v1/crawl/service.proto",
}

// CrawlerManagerClient is the client API for CrawlerManager service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type CrawlerManagerClient interface {
	// Crawlers
	GetCrawlers(ctx context.Context, in *GetCrawlersRequest, opts ...grpc.CallOption) (*GetCrawlersResponse, error)
	// GetCrawler
	GetCrawler(ctx context.Context, in *GetCrawlerRequest, opts ...grpc.CallOption) (*GetCrawlerResponse, error)
	// GetCrawlerOptions
	GetCrawlerOptions(ctx context.Context, in *GetCrawlerOptionsRequest, opts ...grpc.CallOption) (*GetCrawlerOptionsResponse, error)
	// GetCanonicalUrl
	GetCanonicalUrl(ctx context.Context, in *GetCanonicalUrlRequest, opts ...grpc.CallOption) (*GetCanonicalUrlResponse, error)
	// 抓取 @desc 提交URL地址
	// 对于不同情况下，抓取的数据响应处理方式不同;
	// 对于定时抓取任务，或者全库抓取任务，抓取数据通过MQ提交给处理逻辑
	//
	// 任何一个实现了该接口的爬虫服务，都需要将在服务启动后将自身的爬虫信息
	// 提交给爬虫管理中心；具体的数据格式见`CrawlerController`
	DoParse(ctx context.Context, in *DoParseRequest, opts ...grpc.CallOption) (*DoParseResponse, error)
	// RemoteCall
	RemoteCall(ctx context.Context, in *RemoteCallRequest, opts ...grpc.CallOption) (*RemoteCallResponse, error)
}

type crawlerManagerClient struct {
	cc grpc.ClientConnInterface
}

func NewCrawlerManagerClient(cc grpc.ClientConnInterface) CrawlerManagerClient {
	return &crawlerManagerClient{cc}
}

func (c *crawlerManagerClient) GetCrawlers(ctx context.Context, in *GetCrawlersRequest, opts ...grpc.CallOption) (*GetCrawlersResponse, error) {
	out := new(GetCrawlersResponse)
	err := c.cc.Invoke(ctx, "/chameleon.smelter.v1.crawl.CrawlerManager/GetCrawlers", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *crawlerManagerClient) GetCrawler(ctx context.Context, in *GetCrawlerRequest, opts ...grpc.CallOption) (*GetCrawlerResponse, error) {
	out := new(GetCrawlerResponse)
	err := c.cc.Invoke(ctx, "/chameleon.smelter.v1.crawl.CrawlerManager/GetCrawler", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *crawlerManagerClient) GetCrawlerOptions(ctx context.Context, in *GetCrawlerOptionsRequest, opts ...grpc.CallOption) (*GetCrawlerOptionsResponse, error) {
	out := new(GetCrawlerOptionsResponse)
	err := c.cc.Invoke(ctx, "/chameleon.smelter.v1.crawl.CrawlerManager/GetCrawlerOptions", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *crawlerManagerClient) GetCanonicalUrl(ctx context.Context, in *GetCanonicalUrlRequest, opts ...grpc.CallOption) (*GetCanonicalUrlResponse, error) {
	out := new(GetCanonicalUrlResponse)
	err := c.cc.Invoke(ctx, "/chameleon.smelter.v1.crawl.CrawlerManager/GetCanonicalUrl", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *crawlerManagerClient) DoParse(ctx context.Context, in *DoParseRequest, opts ...grpc.CallOption) (*DoParseResponse, error) {
	out := new(DoParseResponse)
	err := c.cc.Invoke(ctx, "/chameleon.smelter.v1.crawl.CrawlerManager/DoParse", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *crawlerManagerClient) RemoteCall(ctx context.Context, in *RemoteCallRequest, opts ...grpc.CallOption) (*RemoteCallResponse, error) {
	out := new(RemoteCallResponse)
	err := c.cc.Invoke(ctx, "/chameleon.smelter.v1.crawl.CrawlerManager/RemoteCall", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// CrawlerManagerServer is the server API for CrawlerManager service.
// All implementations must embed UnimplementedCrawlerManagerServer
// for forward compatibility
type CrawlerManagerServer interface {
	// Crawlers
	GetCrawlers(context.Context, *GetCrawlersRequest) (*GetCrawlersResponse, error)
	// GetCrawler
	GetCrawler(context.Context, *GetCrawlerRequest) (*GetCrawlerResponse, error)
	// GetCrawlerOptions
	GetCrawlerOptions(context.Context, *GetCrawlerOptionsRequest) (*GetCrawlerOptionsResponse, error)
	// GetCanonicalUrl
	GetCanonicalUrl(context.Context, *GetCanonicalUrlRequest) (*GetCanonicalUrlResponse, error)
	// 抓取 @desc 提交URL地址
	// 对于不同情况下，抓取的数据响应处理方式不同;
	// 对于定时抓取任务，或者全库抓取任务，抓取数据通过MQ提交给处理逻辑
	//
	// 任何一个实现了该接口的爬虫服务，都需要将在服务启动后将自身的爬虫信息
	// 提交给爬虫管理中心；具体的数据格式见`CrawlerController`
	DoParse(context.Context, *DoParseRequest) (*DoParseResponse, error)
	// RemoteCall
	RemoteCall(context.Context, *RemoteCallRequest) (*RemoteCallResponse, error)
	mustEmbedUnimplementedCrawlerManagerServer()
}

// UnimplementedCrawlerManagerServer must be embedded to have forward compatible implementations.
type UnimplementedCrawlerManagerServer struct {
}

func (UnimplementedCrawlerManagerServer) GetCrawlers(context.Context, *GetCrawlersRequest) (*GetCrawlersResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetCrawlers not implemented")
}
func (UnimplementedCrawlerManagerServer) GetCrawler(context.Context, *GetCrawlerRequest) (*GetCrawlerResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetCrawler not implemented")
}
func (UnimplementedCrawlerManagerServer) GetCrawlerOptions(context.Context, *GetCrawlerOptionsRequest) (*GetCrawlerOptionsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetCrawlerOptions not implemented")
}
func (UnimplementedCrawlerManagerServer) GetCanonicalUrl(context.Context, *GetCanonicalUrlRequest) (*GetCanonicalUrlResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetCanonicalUrl not implemented")
}
func (UnimplementedCrawlerManagerServer) DoParse(context.Context, *DoParseRequest) (*DoParseResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DoParse not implemented")
}
func (UnimplementedCrawlerManagerServer) RemoteCall(context.Context, *RemoteCallRequest) (*RemoteCallResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RemoteCall not implemented")
}
func (UnimplementedCrawlerManagerServer) mustEmbedUnimplementedCrawlerManagerServer() {}

// UnsafeCrawlerManagerServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to CrawlerManagerServer will
// result in compilation errors.
type UnsafeCrawlerManagerServer interface {
	mustEmbedUnimplementedCrawlerManagerServer()
}

func RegisterCrawlerManagerServer(s *grpc.Server, srv CrawlerManagerServer) {
	s.RegisterService(&_CrawlerManager_serviceDesc, srv)
}

func _CrawlerManager_GetCrawlers_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetCrawlersRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CrawlerManagerServer).GetCrawlers(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/chameleon.smelter.v1.crawl.CrawlerManager/GetCrawlers",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CrawlerManagerServer).GetCrawlers(ctx, req.(*GetCrawlersRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CrawlerManager_GetCrawler_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetCrawlerRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CrawlerManagerServer).GetCrawler(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/chameleon.smelter.v1.crawl.CrawlerManager/GetCrawler",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CrawlerManagerServer).GetCrawler(ctx, req.(*GetCrawlerRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CrawlerManager_GetCrawlerOptions_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetCrawlerOptionsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CrawlerManagerServer).GetCrawlerOptions(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/chameleon.smelter.v1.crawl.CrawlerManager/GetCrawlerOptions",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CrawlerManagerServer).GetCrawlerOptions(ctx, req.(*GetCrawlerOptionsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CrawlerManager_GetCanonicalUrl_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetCanonicalUrlRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CrawlerManagerServer).GetCanonicalUrl(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/chameleon.smelter.v1.crawl.CrawlerManager/GetCanonicalUrl",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CrawlerManagerServer).GetCanonicalUrl(ctx, req.(*GetCanonicalUrlRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CrawlerManager_DoParse_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DoParseRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CrawlerManagerServer).DoParse(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/chameleon.smelter.v1.crawl.CrawlerManager/DoParse",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CrawlerManagerServer).DoParse(ctx, req.(*DoParseRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CrawlerManager_RemoteCall_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RemoteCallRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CrawlerManagerServer).RemoteCall(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/chameleon.smelter.v1.crawl.CrawlerManager/RemoteCall",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CrawlerManagerServer).RemoteCall(ctx, req.(*RemoteCallRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _CrawlerManager_serviceDesc = grpc.ServiceDesc{
	ServiceName: "chameleon.smelter.v1.crawl.CrawlerManager",
	HandlerType: (*CrawlerManagerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetCrawlers",
			Handler:    _CrawlerManager_GetCrawlers_Handler,
		},
		{
			MethodName: "GetCrawler",
			Handler:    _CrawlerManager_GetCrawler_Handler,
		},
		{
			MethodName: "GetCrawlerOptions",
			Handler:    _CrawlerManager_GetCrawlerOptions_Handler,
		},
		{
			MethodName: "GetCanonicalUrl",
			Handler:    _CrawlerManager_GetCanonicalUrl_Handler,
		},
		{
			MethodName: "DoParse",
			Handler:    _CrawlerManager_DoParse_Handler,
		},
		{
			MethodName: "RemoteCall",
			Handler:    _CrawlerManager_RemoteCall_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "chameleon/smelter/v1/crawl/service.proto",
}

// GatewayClient is the client API for Gateway service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type GatewayClient interface {
	// Crawlers
	GetCrawlers(ctx context.Context, in *GetCrawlersRequest, opts ...grpc.CallOption) (*GetCrawlersResponse, error)
	// GetCrawler
	GetCrawler(ctx context.Context, in *GetCrawlerRequest, opts ...grpc.CallOption) (*GetCrawlerResponse, error)
	// GetCanonicalUrl
	GetCanonicalUrl(ctx context.Context, in *GetCanonicalUrlRequest, opts ...grpc.CallOption) (*GetCanonicalUrlResponse, error)
	// RemoteCall
	RemoteCall(ctx context.Context, in *RemoteCallRequest, opts ...grpc.CallOption) (*RemoteCallResponse, error)
	// 抓取 @desc 提交URL地址
	// 对于不同情况下，抓取的数据响应处理方式不同;
	// 对于定时抓取任务，或者全库抓取任务，抓取数据通过MQ提交给处理逻辑
	// 对于及时抓取，比如获取商品的价格，直接返回结果
	//
	// 任何一个实现了该接口的爬虫服务，都需要将在服务启动后将自身的爬虫信息
	// 提交给爬虫管理中心；具体的数据格式见`CrawlerController`
	Fetch(ctx context.Context, in *FetchRequest, opts ...grpc.CallOption) (*FetchResponse, error)
	// GetRequest
	GetRequest(ctx context.Context, in *GetRequestRequest, opts ...grpc.CallOption) (*GetRequestResponse, error)
	// GetCrawlerLogs
	GetCrawlerLogs(ctx context.Context, in *GetCrawlerLogsRequest, opts ...grpc.CallOption) (*GetCrawlerLogsResponse, error)
}

type gatewayClient struct {
	cc grpc.ClientConnInterface
}

func NewGatewayClient(cc grpc.ClientConnInterface) GatewayClient {
	return &gatewayClient{cc}
}

func (c *gatewayClient) GetCrawlers(ctx context.Context, in *GetCrawlersRequest, opts ...grpc.CallOption) (*GetCrawlersResponse, error) {
	out := new(GetCrawlersResponse)
	err := c.cc.Invoke(ctx, "/chameleon.smelter.v1.crawl.Gateway/GetCrawlers", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *gatewayClient) GetCrawler(ctx context.Context, in *GetCrawlerRequest, opts ...grpc.CallOption) (*GetCrawlerResponse, error) {
	out := new(GetCrawlerResponse)
	err := c.cc.Invoke(ctx, "/chameleon.smelter.v1.crawl.Gateway/GetCrawler", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *gatewayClient) GetCanonicalUrl(ctx context.Context, in *GetCanonicalUrlRequest, opts ...grpc.CallOption) (*GetCanonicalUrlResponse, error) {
	out := new(GetCanonicalUrlResponse)
	err := c.cc.Invoke(ctx, "/chameleon.smelter.v1.crawl.Gateway/GetCanonicalUrl", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *gatewayClient) RemoteCall(ctx context.Context, in *RemoteCallRequest, opts ...grpc.CallOption) (*RemoteCallResponse, error) {
	out := new(RemoteCallResponse)
	err := c.cc.Invoke(ctx, "/chameleon.smelter.v1.crawl.Gateway/RemoteCall", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *gatewayClient) Fetch(ctx context.Context, in *FetchRequest, opts ...grpc.CallOption) (*FetchResponse, error) {
	out := new(FetchResponse)
	err := c.cc.Invoke(ctx, "/chameleon.smelter.v1.crawl.Gateway/Fetch", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *gatewayClient) GetRequest(ctx context.Context, in *GetRequestRequest, opts ...grpc.CallOption) (*GetRequestResponse, error) {
	out := new(GetRequestResponse)
	err := c.cc.Invoke(ctx, "/chameleon.smelter.v1.crawl.Gateway/GetRequest", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *gatewayClient) GetCrawlerLogs(ctx context.Context, in *GetCrawlerLogsRequest, opts ...grpc.CallOption) (*GetCrawlerLogsResponse, error) {
	out := new(GetCrawlerLogsResponse)
	err := c.cc.Invoke(ctx, "/chameleon.smelter.v1.crawl.Gateway/GetCrawlerLogs", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// GatewayServer is the server API for Gateway service.
// All implementations must embed UnimplementedGatewayServer
// for forward compatibility
type GatewayServer interface {
	// Crawlers
	GetCrawlers(context.Context, *GetCrawlersRequest) (*GetCrawlersResponse, error)
	// GetCrawler
	GetCrawler(context.Context, *GetCrawlerRequest) (*GetCrawlerResponse, error)
	// GetCanonicalUrl
	GetCanonicalUrl(context.Context, *GetCanonicalUrlRequest) (*GetCanonicalUrlResponse, error)
	// RemoteCall
	RemoteCall(context.Context, *RemoteCallRequest) (*RemoteCallResponse, error)
	// 抓取 @desc 提交URL地址
	// 对于不同情况下，抓取的数据响应处理方式不同;
	// 对于定时抓取任务，或者全库抓取任务，抓取数据通过MQ提交给处理逻辑
	// 对于及时抓取，比如获取商品的价格，直接返回结果
	//
	// 任何一个实现了该接口的爬虫服务，都需要将在服务启动后将自身的爬虫信息
	// 提交给爬虫管理中心；具体的数据格式见`CrawlerController`
	Fetch(context.Context, *FetchRequest) (*FetchResponse, error)
	// GetRequest
	GetRequest(context.Context, *GetRequestRequest) (*GetRequestResponse, error)
	// GetCrawlerLogs
	GetCrawlerLogs(context.Context, *GetCrawlerLogsRequest) (*GetCrawlerLogsResponse, error)
	mustEmbedUnimplementedGatewayServer()
}

// UnimplementedGatewayServer must be embedded to have forward compatible implementations.
type UnimplementedGatewayServer struct {
}

func (UnimplementedGatewayServer) GetCrawlers(context.Context, *GetCrawlersRequest) (*GetCrawlersResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetCrawlers not implemented")
}
func (UnimplementedGatewayServer) GetCrawler(context.Context, *GetCrawlerRequest) (*GetCrawlerResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetCrawler not implemented")
}
func (UnimplementedGatewayServer) GetCanonicalUrl(context.Context, *GetCanonicalUrlRequest) (*GetCanonicalUrlResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetCanonicalUrl not implemented")
}
func (UnimplementedGatewayServer) RemoteCall(context.Context, *RemoteCallRequest) (*RemoteCallResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RemoteCall not implemented")
}
func (UnimplementedGatewayServer) Fetch(context.Context, *FetchRequest) (*FetchResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Fetch not implemented")
}
func (UnimplementedGatewayServer) GetRequest(context.Context, *GetRequestRequest) (*GetRequestResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetRequest not implemented")
}
func (UnimplementedGatewayServer) GetCrawlerLogs(context.Context, *GetCrawlerLogsRequest) (*GetCrawlerLogsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetCrawlerLogs not implemented")
}
func (UnimplementedGatewayServer) mustEmbedUnimplementedGatewayServer() {}

// UnsafeGatewayServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to GatewayServer will
// result in compilation errors.
type UnsafeGatewayServer interface {
	mustEmbedUnimplementedGatewayServer()
}

func RegisterGatewayServer(s *grpc.Server, srv GatewayServer) {
	s.RegisterService(&_Gateway_serviceDesc, srv)
}

func _Gateway_GetCrawlers_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetCrawlersRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GatewayServer).GetCrawlers(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/chameleon.smelter.v1.crawl.Gateway/GetCrawlers",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GatewayServer).GetCrawlers(ctx, req.(*GetCrawlersRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Gateway_GetCrawler_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetCrawlerRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GatewayServer).GetCrawler(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/chameleon.smelter.v1.crawl.Gateway/GetCrawler",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GatewayServer).GetCrawler(ctx, req.(*GetCrawlerRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Gateway_GetCanonicalUrl_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetCanonicalUrlRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GatewayServer).GetCanonicalUrl(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/chameleon.smelter.v1.crawl.Gateway/GetCanonicalUrl",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GatewayServer).GetCanonicalUrl(ctx, req.(*GetCanonicalUrlRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Gateway_RemoteCall_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RemoteCallRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GatewayServer).RemoteCall(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/chameleon.smelter.v1.crawl.Gateway/RemoteCall",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GatewayServer).RemoteCall(ctx, req.(*RemoteCallRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Gateway_Fetch_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(FetchRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GatewayServer).Fetch(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/chameleon.smelter.v1.crawl.Gateway/Fetch",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GatewayServer).Fetch(ctx, req.(*FetchRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Gateway_GetRequest_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetRequestRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GatewayServer).GetRequest(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/chameleon.smelter.v1.crawl.Gateway/GetRequest",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GatewayServer).GetRequest(ctx, req.(*GetRequestRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Gateway_GetCrawlerLogs_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetCrawlerLogsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GatewayServer).GetCrawlerLogs(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/chameleon.smelter.v1.crawl.Gateway/GetCrawlerLogs",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GatewayServer).GetCrawlerLogs(ctx, req.(*GetCrawlerLogsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Gateway_serviceDesc = grpc.ServiceDesc{
	ServiceName: "chameleon.smelter.v1.crawl.Gateway",
	HandlerType: (*GatewayServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetCrawlers",
			Handler:    _Gateway_GetCrawlers_Handler,
		},
		{
			MethodName: "GetCrawler",
			Handler:    _Gateway_GetCrawler_Handler,
		},
		{
			MethodName: "GetCanonicalUrl",
			Handler:    _Gateway_GetCanonicalUrl_Handler,
		},
		{
			MethodName: "RemoteCall",
			Handler:    _Gateway_RemoteCall_Handler,
		},
		{
			MethodName: "Fetch",
			Handler:    _Gateway_Fetch_Handler,
		},
		{
			MethodName: "GetRequest",
			Handler:    _Gateway_GetRequest_Handler,
		},
		{
			MethodName: "GetCrawlerLogs",
			Handler:    _Gateway_GetCrawlerLogs_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "chameleon/smelter/v1/crawl/service.proto",
}