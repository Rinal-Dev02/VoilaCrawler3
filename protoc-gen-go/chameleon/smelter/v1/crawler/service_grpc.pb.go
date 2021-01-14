// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package crawler

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

// CrawlerManagerClient is the client API for CrawlerManager service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type CrawlerManagerClient interface {
	// 获取爬虫列表
	GetCrawlers(ctx context.Context, in *GetCrawlersRequest, opts ...grpc.CallOption) (*GetCrawlersResponse, error)
	// 获取爬虫详情 @desc 获取爬虫详情，包括状态数据
	GetCrawler(ctx context.Context, in *GetCrawlerRequest, opts ...grpc.CallOption) (*GetCrawlerResponse, error)
	// 禁用爬虫
	CordonCrawler(ctx context.Context, in *CordonCrawlerRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	// 启用爬虫
	UncordonCrawler(ctx context.Context, in *UncordonCrawlerRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
}

type crawlerManagerClient struct {
	cc grpc.ClientConnInterface
}

func NewCrawlerManagerClient(cc grpc.ClientConnInterface) CrawlerManagerClient {
	return &crawlerManagerClient{cc}
}

func (c *crawlerManagerClient) GetCrawlers(ctx context.Context, in *GetCrawlersRequest, opts ...grpc.CallOption) (*GetCrawlersResponse, error) {
	out := new(GetCrawlersResponse)
	err := c.cc.Invoke(ctx, "/chameleon.smelter.v1.crawler.CrawlerManager/GetCrawlers", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *crawlerManagerClient) GetCrawler(ctx context.Context, in *GetCrawlerRequest, opts ...grpc.CallOption) (*GetCrawlerResponse, error) {
	out := new(GetCrawlerResponse)
	err := c.cc.Invoke(ctx, "/chameleon.smelter.v1.crawler.CrawlerManager/GetCrawler", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *crawlerManagerClient) CordonCrawler(ctx context.Context, in *CordonCrawlerRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/chameleon.smelter.v1.crawler.CrawlerManager/CordonCrawler", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *crawlerManagerClient) UncordonCrawler(ctx context.Context, in *UncordonCrawlerRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/chameleon.smelter.v1.crawler.CrawlerManager/UncordonCrawler", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// CrawlerManagerServer is the server API for CrawlerManager service.
// All implementations must embed UnimplementedCrawlerManagerServer
// for forward compatibility
type CrawlerManagerServer interface {
	// 获取爬虫列表
	GetCrawlers(context.Context, *GetCrawlersRequest) (*GetCrawlersResponse, error)
	// 获取爬虫详情 @desc 获取爬虫详情，包括状态数据
	GetCrawler(context.Context, *GetCrawlerRequest) (*GetCrawlerResponse, error)
	// 禁用爬虫
	CordonCrawler(context.Context, *CordonCrawlerRequest) (*emptypb.Empty, error)
	// 启用爬虫
	UncordonCrawler(context.Context, *UncordonCrawlerRequest) (*emptypb.Empty, error)
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
func (UnimplementedCrawlerManagerServer) CordonCrawler(context.Context, *CordonCrawlerRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CordonCrawler not implemented")
}
func (UnimplementedCrawlerManagerServer) UncordonCrawler(context.Context, *UncordonCrawlerRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UncordonCrawler not implemented")
}
func (UnimplementedCrawlerManagerServer) mustEmbedUnimplementedCrawlerManagerServer() {}

// UnsafeCrawlerManagerServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to CrawlerManagerServer will
// result in compilation errors.
type UnsafeCrawlerManagerServer interface {
	mustEmbedUnimplementedCrawlerManagerServer()
}

func RegisterCrawlerManagerServer(s grpc.ServiceRegistrar, srv CrawlerManagerServer) {
	s.RegisterService(&CrawlerManager_ServiceDesc, srv)
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
		FullMethod: "/chameleon.smelter.v1.crawler.CrawlerManager/GetCrawlers",
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
		FullMethod: "/chameleon.smelter.v1.crawler.CrawlerManager/GetCrawler",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CrawlerManagerServer).GetCrawler(ctx, req.(*GetCrawlerRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CrawlerManager_CordonCrawler_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CordonCrawlerRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CrawlerManagerServer).CordonCrawler(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/chameleon.smelter.v1.crawler.CrawlerManager/CordonCrawler",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CrawlerManagerServer).CordonCrawler(ctx, req.(*CordonCrawlerRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CrawlerManager_UncordonCrawler_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UncordonCrawlerRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CrawlerManagerServer).UncordonCrawler(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/chameleon.smelter.v1.crawler.CrawlerManager/UncordonCrawler",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CrawlerManagerServer).UncordonCrawler(ctx, req.(*UncordonCrawlerRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// CrawlerManager_ServiceDesc is the grpc.ServiceDesc for CrawlerManager service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var CrawlerManager_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "chameleon.smelter.v1.crawler.CrawlerManager",
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
			MethodName: "CordonCrawler",
			Handler:    _CrawlerManager_CordonCrawler_Handler,
		},
		{
			MethodName: "UncordonCrawler",
			Handler:    _CrawlerManager_UncordonCrawler_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "chameleon/smelter/v1/crawler/service.proto",
}

// NodeManagerClient is the client API for NodeManager service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type NodeManagerClient interface {
	// GetNodes
	GetNodes(ctx context.Context, in *GetNodesRequest, opts ...grpc.CallOption) (*GetNodesResponse, error)
	// GetNode
	GetNode(ctx context.Context, in *GetNodeRequest, opts ...grpc.CallOption) (*GetNodeResponse, error)
}

type nodeManagerClient struct {
	cc grpc.ClientConnInterface
}

func NewNodeManagerClient(cc grpc.ClientConnInterface) NodeManagerClient {
	return &nodeManagerClient{cc}
}

func (c *nodeManagerClient) GetNodes(ctx context.Context, in *GetNodesRequest, opts ...grpc.CallOption) (*GetNodesResponse, error) {
	out := new(GetNodesResponse)
	err := c.cc.Invoke(ctx, "/chameleon.smelter.v1.crawler.NodeManager/GetNodes", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *nodeManagerClient) GetNode(ctx context.Context, in *GetNodeRequest, opts ...grpc.CallOption) (*GetNodeResponse, error) {
	out := new(GetNodeResponse)
	err := c.cc.Invoke(ctx, "/chameleon.smelter.v1.crawler.NodeManager/GetNode", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// NodeManagerServer is the server API for NodeManager service.
// All implementations must embed UnimplementedNodeManagerServer
// for forward compatibility
type NodeManagerServer interface {
	// GetNodes
	GetNodes(context.Context, *GetNodesRequest) (*GetNodesResponse, error)
	// GetNode
	GetNode(context.Context, *GetNodeRequest) (*GetNodeResponse, error)
	mustEmbedUnimplementedNodeManagerServer()
}

// UnimplementedNodeManagerServer must be embedded to have forward compatible implementations.
type UnimplementedNodeManagerServer struct {
}

func (UnimplementedNodeManagerServer) GetNodes(context.Context, *GetNodesRequest) (*GetNodesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetNodes not implemented")
}
func (UnimplementedNodeManagerServer) GetNode(context.Context, *GetNodeRequest) (*GetNodeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetNode not implemented")
}
func (UnimplementedNodeManagerServer) mustEmbedUnimplementedNodeManagerServer() {}

// UnsafeNodeManagerServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to NodeManagerServer will
// result in compilation errors.
type UnsafeNodeManagerServer interface {
	mustEmbedUnimplementedNodeManagerServer()
}

func RegisterNodeManagerServer(s grpc.ServiceRegistrar, srv NodeManagerServer) {
	s.RegisterService(&NodeManager_ServiceDesc, srv)
}

func _NodeManager_GetNodes_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetNodesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NodeManagerServer).GetNodes(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/chameleon.smelter.v1.crawler.NodeManager/GetNodes",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NodeManagerServer).GetNodes(ctx, req.(*GetNodesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _NodeManager_GetNode_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetNodeRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NodeManagerServer).GetNode(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/chameleon.smelter.v1.crawler.NodeManager/GetNode",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NodeManagerServer).GetNode(ctx, req.(*GetNodeRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// NodeManager_ServiceDesc is the grpc.ServiceDesc for NodeManager service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var NodeManager_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "chameleon.smelter.v1.crawler.NodeManager",
	HandlerType: (*NodeManagerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetNodes",
			Handler:    _NodeManager_GetNodes_Handler,
		},
		{
			MethodName: "GetNode",
			Handler:    _NodeManager_GetNode_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "chameleon/smelter/v1/crawler/service.proto",
}

// CrawlerControllerClient is the client API for CrawlerController service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type CrawlerControllerClient interface {
	// 通道 @desc 用于双方发送指令
	// 从节点向控制节点发送心跳包，包括：节点信息，爬虫信息，可处理的任务信息，当前正在处理的任务信息
	// 主节点通过该节点下发任务数据
	Channel(ctx context.Context, opts ...grpc.CallOption) (CrawlerController_ChannelClient, error)
	// 抓取 @desc 提交URL地址
	// 对于不同情况下，抓取的数据响应处理方式不同;
	// 对于定时抓取任务，或者全库抓取任务，抓取数据通过MQ提交给处理逻辑
	// 对于及时抓取，比如获取商品的价格，直接返回结果
	//
	// 任何一个实现了该接口的爬虫服务，都需要将在服务启动后将自身的爬虫信息
	// 提交给爬虫管理中心；具体的数据格式见`CrawlerController`
	Fetch(ctx context.Context, in *FetchRequest, opts ...grpc.CallOption) (*FetchResponse, error)
}

type crawlerControllerClient struct {
	cc grpc.ClientConnInterface
}

func NewCrawlerControllerClient(cc grpc.ClientConnInterface) CrawlerControllerClient {
	return &crawlerControllerClient{cc}
}

func (c *crawlerControllerClient) Channel(ctx context.Context, opts ...grpc.CallOption) (CrawlerController_ChannelClient, error) {
	stream, err := c.cc.NewStream(ctx, &CrawlerController_ServiceDesc.Streams[0], "/chameleon.smelter.v1.crawler.CrawlerController/Channel", opts...)
	if err != nil {
		return nil, err
	}
	x := &crawlerControllerChannelClient{stream}
	return x, nil
}

type CrawlerController_ChannelClient interface {
	Send(*anypb.Any) error
	Recv() (*anypb.Any, error)
	grpc.ClientStream
}

type crawlerControllerChannelClient struct {
	grpc.ClientStream
}

func (x *crawlerControllerChannelClient) Send(m *anypb.Any) error {
	return x.ClientStream.SendMsg(m)
}

func (x *crawlerControllerChannelClient) Recv() (*anypb.Any, error) {
	m := new(anypb.Any)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *crawlerControllerClient) Fetch(ctx context.Context, in *FetchRequest, opts ...grpc.CallOption) (*FetchResponse, error) {
	out := new(FetchResponse)
	err := c.cc.Invoke(ctx, "/chameleon.smelter.v1.crawler.CrawlerController/Fetch", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// CrawlerControllerServer is the server API for CrawlerController service.
// All implementations must embed UnimplementedCrawlerControllerServer
// for forward compatibility
type CrawlerControllerServer interface {
	// 通道 @desc 用于双方发送指令
	// 从节点向控制节点发送心跳包，包括：节点信息，爬虫信息，可处理的任务信息，当前正在处理的任务信息
	// 主节点通过该节点下发任务数据
	Channel(CrawlerController_ChannelServer) error
	// 抓取 @desc 提交URL地址
	// 对于不同情况下，抓取的数据响应处理方式不同;
	// 对于定时抓取任务，或者全库抓取任务，抓取数据通过MQ提交给处理逻辑
	// 对于及时抓取，比如获取商品的价格，直接返回结果
	//
	// 任何一个实现了该接口的爬虫服务，都需要将在服务启动后将自身的爬虫信息
	// 提交给爬虫管理中心；具体的数据格式见`CrawlerController`
	Fetch(context.Context, *FetchRequest) (*FetchResponse, error)
	mustEmbedUnimplementedCrawlerControllerServer()
}

// UnimplementedCrawlerControllerServer must be embedded to have forward compatible implementations.
type UnimplementedCrawlerControllerServer struct {
}

func (UnimplementedCrawlerControllerServer) Channel(CrawlerController_ChannelServer) error {
	return status.Errorf(codes.Unimplemented, "method Channel not implemented")
}
func (UnimplementedCrawlerControllerServer) Fetch(context.Context, *FetchRequest) (*FetchResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Fetch not implemented")
}
func (UnimplementedCrawlerControllerServer) mustEmbedUnimplementedCrawlerControllerServer() {}

// UnsafeCrawlerControllerServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to CrawlerControllerServer will
// result in compilation errors.
type UnsafeCrawlerControllerServer interface {
	mustEmbedUnimplementedCrawlerControllerServer()
}

func RegisterCrawlerControllerServer(s grpc.ServiceRegistrar, srv CrawlerControllerServer) {
	s.RegisterService(&CrawlerController_ServiceDesc, srv)
}

func _CrawlerController_Channel_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(CrawlerControllerServer).Channel(&crawlerControllerChannelServer{stream})
}

type CrawlerController_ChannelServer interface {
	Send(*anypb.Any) error
	Recv() (*anypb.Any, error)
	grpc.ServerStream
}

type crawlerControllerChannelServer struct {
	grpc.ServerStream
}

func (x *crawlerControllerChannelServer) Send(m *anypb.Any) error {
	return x.ServerStream.SendMsg(m)
}

func (x *crawlerControllerChannelServer) Recv() (*anypb.Any, error) {
	m := new(anypb.Any)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _CrawlerController_Fetch_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(FetchRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CrawlerControllerServer).Fetch(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/chameleon.smelter.v1.crawler.CrawlerController/Fetch",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CrawlerControllerServer).Fetch(ctx, req.(*FetchRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// CrawlerController_ServiceDesc is the grpc.ServiceDesc for CrawlerController service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var CrawlerController_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "chameleon.smelter.v1.crawler.CrawlerController",
	HandlerType: (*CrawlerControllerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Fetch",
			Handler:    _CrawlerController_Fetch_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Channel",
			Handler:       _CrawlerController_Channel_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "chameleon/smelter/v1/crawler/service.proto",
}
