// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0-devel
// 	protoc        v3.14.0
// source: chameleon/smelter/v1/crawl/proxy/data.proto

package proxy

import (
	http "github.com/voiladev/go-crawler/protoc-gen-go/chameleon/api/http"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// ProxyReliability
type ProxyReliability int32

const (
	// ReliabilityDefault will use ReliabilityLow proxies first, if failed, use ReliabilityHigh proxy later
	ProxyReliability_ReliabilityDefault ProxyReliability = 0
	// ReliabilityLow use backconnect proxy which use the shared proxies
	ProxyReliability_ReliabilityLow ProxyReliability = 1
	// ReliabilityMedium
	ProxyReliability_ReliabilityMedium ProxyReliability = 2
	// ReliabilityHigh use high reliable proxy
	ProxyReliability_ReliabilityHigh ProxyReliability = 3
	// (TODO)ReliabilityIntelligent judge the policy according to history status
	ProxyReliability_ReliabilityIntelligent ProxyReliability = 10
)

// Enum value maps for ProxyReliability.
var (
	ProxyReliability_name = map[int32]string{
		0:  "ReliabilityDefault",
		1:  "ReliabilityLow",
		2:  "ReliabilityMedium",
		3:  "ReliabilityHigh",
		10: "ReliabilityIntelligent",
	}
	ProxyReliability_value = map[string]int32{
		"ReliabilityDefault":     0,
		"ReliabilityLow":         1,
		"ReliabilityMedium":      2,
		"ReliabilityHigh":        3,
		"ReliabilityIntelligent": 10,
	}
)

func (x ProxyReliability) Enum() *ProxyReliability {
	p := new(ProxyReliability)
	*p = x
	return p
}

func (x ProxyReliability) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (ProxyReliability) Descriptor() protoreflect.EnumDescriptor {
	return file_chameleon_smelter_v1_crawl_proxy_data_proto_enumTypes[0].Descriptor()
}

func (ProxyReliability) Type() protoreflect.EnumType {
	return &file_chameleon_smelter_v1_crawl_proxy_data_proto_enumTypes[0]
}

func (x ProxyReliability) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use ProxyReliability.Descriptor instead.
func (ProxyReliability) EnumDescriptor() ([]byte, []int) {
	return file_chameleon_smelter_v1_crawl_proxy_data_proto_rawDescGZIP(), []int{0}
}

// Request
type Request struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// TracingId
	TracingId string `protobuf:"bytes,1,opt,name=tracingId,proto3" json:"tracingId,omitempty"`
	// JobId
	JobId string `protobuf:"bytes,2,opt,name=jobId,proto3" json:"jobId,omitempty"`
	// ReqId
	ReqId string `protobuf:"bytes,3,opt,name=reqId,proto3" json:"reqId,omitempty"`
	// Method
	Method string `protobuf:"bytes,6,opt,name=method,proto3" json:"method,omitempty"`
	// URL
	Url string `protobuf:"bytes,7,opt,name=url,proto3" json:"url,omitempty"`
	// Headers
	Headers map[string]*http.ListValue `protobuf:"bytes,8,rep,name=headers,proto3" json:"headers,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	// Body
	Body []byte `protobuf:"bytes,9,opt,name=body,proto3" json:"body,omitempty"`
	// Options
	Options *Request_Options `protobuf:"bytes,11,opt,name=options,proto3" json:"options,omitempty"`
	// Response
	Response *Response `protobuf:"bytes,15,opt,name=response,proto3" json:"response,omitempty"`
}

func (x *Request) Reset() {
	*x = Request{}
	if protoimpl.UnsafeEnabled {
		mi := &file_chameleon_smelter_v1_crawl_proxy_data_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Request) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Request) ProtoMessage() {}

func (x *Request) ProtoReflect() protoreflect.Message {
	mi := &file_chameleon_smelter_v1_crawl_proxy_data_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Request.ProtoReflect.Descriptor instead.
func (*Request) Descriptor() ([]byte, []int) {
	return file_chameleon_smelter_v1_crawl_proxy_data_proto_rawDescGZIP(), []int{0}
}

func (x *Request) GetTracingId() string {
	if x != nil {
		return x.TracingId
	}
	return ""
}

func (x *Request) GetJobId() string {
	if x != nil {
		return x.JobId
	}
	return ""
}

func (x *Request) GetReqId() string {
	if x != nil {
		return x.ReqId
	}
	return ""
}

func (x *Request) GetMethod() string {
	if x != nil {
		return x.Method
	}
	return ""
}

func (x *Request) GetUrl() string {
	if x != nil {
		return x.Url
	}
	return ""
}

func (x *Request) GetHeaders() map[string]*http.ListValue {
	if x != nil {
		return x.Headers
	}
	return nil
}

func (x *Request) GetBody() []byte {
	if x != nil {
		return x.Body
	}
	return nil
}

func (x *Request) GetOptions() *Request_Options {
	if x != nil {
		return x.Options
	}
	return nil
}

func (x *Request) GetResponse() *Response {
	if x != nil {
		return x.Response
	}
	return nil
}

// Response
type Response struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// StatusCode
	StatusCode int32 `protobuf:"varint,1,opt,name=statusCode,proto3" json:"statusCode,omitempty"`
	// Status
	Status string `protobuf:"bytes,2,opt,name=status,proto3" json:"status,omitempty"`
	// Proto
	Proto string `protobuf:"bytes,3,opt,name=proto,proto3" json:"proto,omitempty"`
	// ProtoMajor
	ProtoMajor int32 `protobuf:"varint,4,opt,name=protoMajor,proto3" json:"protoMajor,omitempty"`
	// ProtoMinor
	ProtoMinor int32 `protobuf:"varint,5,opt,name=protoMinor,proto3" json:"protoMinor,omitempty"`
	// Headers
	Headers map[string]*http.ListValue `protobuf:"bytes,6,rep,name=headers,proto3" json:"headers,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	// Body
	Body []byte `protobuf:"bytes,9,opt,name=body,proto3" json:"body,omitempty"`
	// (TODO)BodyCacheLink if the size of body to large, use the cached link to fetch
	BodyCacheLink string `protobuf:"bytes,11,opt,name=bodyCacheLink,proto3" json:"bodyCacheLink,omitempty"`
	// Duration
	Duration int64 `protobuf:"varint,12,opt,name=duration,proto3" json:"duration,omitempty"`
	// AverageDuration
	AverageDuration int64 `protobuf:"varint,13,opt,name=averageDuration,proto3" json:"averageDuration,omitempty"`
	// Request
	Request *Request `protobuf:"bytes,15,opt,name=request,proto3" json:"request,omitempty"`
}

func (x *Response) Reset() {
	*x = Response{}
	if protoimpl.UnsafeEnabled {
		mi := &file_chameleon_smelter_v1_crawl_proxy_data_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Response) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Response) ProtoMessage() {}

func (x *Response) ProtoReflect() protoreflect.Message {
	mi := &file_chameleon_smelter_v1_crawl_proxy_data_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Response.ProtoReflect.Descriptor instead.
func (*Response) Descriptor() ([]byte, []int) {
	return file_chameleon_smelter_v1_crawl_proxy_data_proto_rawDescGZIP(), []int{1}
}

func (x *Response) GetStatusCode() int32 {
	if x != nil {
		return x.StatusCode
	}
	return 0
}

func (x *Response) GetStatus() string {
	if x != nil {
		return x.Status
	}
	return ""
}

func (x *Response) GetProto() string {
	if x != nil {
		return x.Proto
	}
	return ""
}

func (x *Response) GetProtoMajor() int32 {
	if x != nil {
		return x.ProtoMajor
	}
	return 0
}

func (x *Response) GetProtoMinor() int32 {
	if x != nil {
		return x.ProtoMinor
	}
	return 0
}

func (x *Response) GetHeaders() map[string]*http.ListValue {
	if x != nil {
		return x.Headers
	}
	return nil
}

func (x *Response) GetBody() []byte {
	if x != nil {
		return x.Body
	}
	return nil
}

func (x *Response) GetBodyCacheLink() string {
	if x != nil {
		return x.BodyCacheLink
	}
	return ""
}

func (x *Response) GetDuration() int64 {
	if x != nil {
		return x.Duration
	}
	return 0
}

func (x *Response) GetAverageDuration() int64 {
	if x != nil {
		return x.AverageDuration
	}
	return 0
}

func (x *Response) GetRequest() *Request {
	if x != nil {
		return x.Request
	}
	return nil
}

// RequestWrap
type RequestWrap struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// ReqId
	ReqId string `protobuf:"bytes,1,opt,name=reqId,proto3" json:"reqId,omitempty"`
	// Request
	Request *Request `protobuf:"bytes,2,opt,name=request,proto3" json:"request,omitempty"`
	// Deadline
	Deadline int64 `protobuf:"varint,6,opt,name=deadline,proto3" json:"deadline,omitempty"`
	// ExecCount
	ExecCount int32 `protobuf:"varint,7,opt,name=execCount,proto3" json:"execCount,omitempty"`
	// Duration
	Duration int64 `protobuf:"varint,11,opt,name=duration,proto3" json:"duration,omitempty"`
	// AverageDuration
	AverageDuration int64 `protobuf:"varint,12,opt,name=averageDuration,proto3" json:"averageDuration,omitempty"`
	// Options
	Options *Request_Options `protobuf:"bytes,15,opt,name=options,proto3" json:"options,omitempty"`
}

func (x *RequestWrap) Reset() {
	*x = RequestWrap{}
	if protoimpl.UnsafeEnabled {
		mi := &file_chameleon_smelter_v1_crawl_proxy_data_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RequestWrap) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RequestWrap) ProtoMessage() {}

func (x *RequestWrap) ProtoReflect() protoreflect.Message {
	mi := &file_chameleon_smelter_v1_crawl_proxy_data_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RequestWrap.ProtoReflect.Descriptor instead.
func (*RequestWrap) Descriptor() ([]byte, []int) {
	return file_chameleon_smelter_v1_crawl_proxy_data_proto_rawDescGZIP(), []int{2}
}

func (x *RequestWrap) GetReqId() string {
	if x != nil {
		return x.ReqId
	}
	return ""
}

func (x *RequestWrap) GetRequest() *Request {
	if x != nil {
		return x.Request
	}
	return nil
}

func (x *RequestWrap) GetDeadline() int64 {
	if x != nil {
		return x.Deadline
	}
	return 0
}

func (x *RequestWrap) GetExecCount() int32 {
	if x != nil {
		return x.ExecCount
	}
	return 0
}

func (x *RequestWrap) GetDuration() int64 {
	if x != nil {
		return x.Duration
	}
	return 0
}

func (x *RequestWrap) GetAverageDuration() int64 {
	if x != nil {
		return x.AverageDuration
	}
	return 0
}

func (x *RequestWrap) GetOptions() *Request_Options {
	if x != nil {
		return x.Options
	}
	return nil
}

// Options
type Request_Options struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// EnableProxy
	EnableProxy bool `protobuf:"varint,1,opt,name=enableProxy,proto3" json:"enableProxy,omitempty"`
	// Reliability
	Reliability ProxyReliability `protobuf:"varint,2,opt,name=reliability,proto3,enum=chameleon.smelter.v1.crawl.proxy.ProxyReliability" json:"reliability,omitempty"`
	// EnableHeadless
	EnableHeadless bool `protobuf:"varint,3,opt,name=enableHeadless,proto3" json:"enableHeadless,omitempty"`
	// JSWaitDuration default 6 seconds
	JsWaitDuration int64 `protobuf:"varint,4,opt,name=jsWaitDuration,proto3" json:"jsWaitDuration,omitempty"`
	// EnableSessionInit
	EnableSessionInit bool `protobuf:"varint,5,opt,name=enableSessionInit,proto3" json:"enableSessionInit,omitempty"`
	// KeepSession
	KeepSession bool `protobuf:"varint,6,opt,name=keepSession,proto3" json:"keepSession,omitempty"`
	// DisableCookieJar disables save cookie to jar
	DisableCookieJar bool `protobuf:"varint,7,opt,name=disableCookieJar,proto3" json:"disableCookieJar,omitempty"`
	// MaxTtlPerRequest
	MaxTtlPerRequest int64 `protobuf:"varint,8,opt,name=maxTtlPerRequest,proto3" json:"maxTtlPerRequest,omitempty"`
	// DisableRedirect
	DisableRedirect bool `protobuf:"varint,11,opt,name=disableRedirect,proto3" json:"disableRedirect,omitempty"`
	// RequestFilterKeys use to filter the response from multi request of the same url
	RequestFilterKeys []string `protobuf:"bytes,15,rep,name=requestFilterKeys,proto3" json:"requestFilterKeys,omitempty"`
}

func (x *Request_Options) Reset() {
	*x = Request_Options{}
	if protoimpl.UnsafeEnabled {
		mi := &file_chameleon_smelter_v1_crawl_proxy_data_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Request_Options) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Request_Options) ProtoMessage() {}

func (x *Request_Options) ProtoReflect() protoreflect.Message {
	mi := &file_chameleon_smelter_v1_crawl_proxy_data_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Request_Options.ProtoReflect.Descriptor instead.
func (*Request_Options) Descriptor() ([]byte, []int) {
	return file_chameleon_smelter_v1_crawl_proxy_data_proto_rawDescGZIP(), []int{0, 0}
}

func (x *Request_Options) GetEnableProxy() bool {
	if x != nil {
		return x.EnableProxy
	}
	return false
}

func (x *Request_Options) GetReliability() ProxyReliability {
	if x != nil {
		return x.Reliability
	}
	return ProxyReliability_ReliabilityDefault
}

func (x *Request_Options) GetEnableHeadless() bool {
	if x != nil {
		return x.EnableHeadless
	}
	return false
}

func (x *Request_Options) GetJsWaitDuration() int64 {
	if x != nil {
		return x.JsWaitDuration
	}
	return 0
}

func (x *Request_Options) GetEnableSessionInit() bool {
	if x != nil {
		return x.EnableSessionInit
	}
	return false
}

func (x *Request_Options) GetKeepSession() bool {
	if x != nil {
		return x.KeepSession
	}
	return false
}

func (x *Request_Options) GetDisableCookieJar() bool {
	if x != nil {
		return x.DisableCookieJar
	}
	return false
}

func (x *Request_Options) GetMaxTtlPerRequest() int64 {
	if x != nil {
		return x.MaxTtlPerRequest
	}
	return 0
}

func (x *Request_Options) GetDisableRedirect() bool {
	if x != nil {
		return x.DisableRedirect
	}
	return false
}

func (x *Request_Options) GetRequestFilterKeys() []string {
	if x != nil {
		return x.RequestFilterKeys
	}
	return nil
}

var File_chameleon_smelter_v1_crawl_proxy_data_proto protoreflect.FileDescriptor

var file_chameleon_smelter_v1_crawl_proxy_data_proto_rawDesc = []byte{
	0x0a, 0x2b, 0x63, 0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x2f, 0x73, 0x6d, 0x65, 0x6c,
	0x74, 0x65, 0x72, 0x2f, 0x76, 0x31, 0x2f, 0x63, 0x72, 0x61, 0x77, 0x6c, 0x2f, 0x70, 0x72, 0x6f,
	0x78, 0x79, 0x2f, 0x64, 0x61, 0x74, 0x61, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x20, 0x63,
	0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x2e, 0x73, 0x6d, 0x65, 0x6c, 0x74, 0x65, 0x72,
	0x2e, 0x76, 0x31, 0x2e, 0x63, 0x72, 0x61, 0x77, 0x6c, 0x2e, 0x70, 0x72, 0x6f, 0x78, 0x79, 0x1a,
	0x1d, 0x63, 0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x68,
	0x74, 0x74, 0x70, 0x2f, 0x64, 0x61, 0x74, 0x61, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xa7,
	0x07, 0x0a, 0x07, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1c, 0x0a, 0x09, 0x74, 0x72,
	0x61, 0x63, 0x69, 0x6e, 0x67, 0x49, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x74,
	0x72, 0x61, 0x63, 0x69, 0x6e, 0x67, 0x49, 0x64, 0x12, 0x14, 0x0a, 0x05, 0x6a, 0x6f, 0x62, 0x49,
	0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x6a, 0x6f, 0x62, 0x49, 0x64, 0x12, 0x14,
	0x0a, 0x05, 0x72, 0x65, 0x71, 0x49, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x72,
	0x65, 0x71, 0x49, 0x64, 0x12, 0x16, 0x0a, 0x06, 0x6d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x18, 0x06,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x6d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x12, 0x10, 0x0a, 0x03,
	0x75, 0x72, 0x6c, 0x18, 0x07, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x75, 0x72, 0x6c, 0x12, 0x50,
	0x0a, 0x07, 0x68, 0x65, 0x61, 0x64, 0x65, 0x72, 0x73, 0x18, 0x08, 0x20, 0x03, 0x28, 0x0b, 0x32,
	0x36, 0x2e, 0x63, 0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x2e, 0x73, 0x6d, 0x65, 0x6c,
	0x74, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x63, 0x72, 0x61, 0x77, 0x6c, 0x2e, 0x70, 0x72, 0x6f,
	0x78, 0x79, 0x2e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x2e, 0x48, 0x65, 0x61, 0x64, 0x65,
	0x72, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x07, 0x68, 0x65, 0x61, 0x64, 0x65, 0x72, 0x73,
	0x12, 0x12, 0x0a, 0x04, 0x62, 0x6f, 0x64, 0x79, 0x18, 0x09, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x04,
	0x62, 0x6f, 0x64, 0x79, 0x12, 0x4b, 0x0a, 0x07, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18,
	0x0b, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x31, 0x2e, 0x63, 0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f,
	0x6e, 0x2e, 0x73, 0x6d, 0x65, 0x6c, 0x74, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x63, 0x72, 0x61,
	0x77, 0x6c, 0x2e, 0x70, 0x72, 0x6f, 0x78, 0x79, 0x2e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x2e, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x52, 0x07, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e,
	0x73, 0x12, 0x46, 0x0a, 0x08, 0x72, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x18, 0x0f, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x2a, 0x2e, 0x63, 0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x2e,
	0x73, 0x6d, 0x65, 0x6c, 0x74, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x63, 0x72, 0x61, 0x77, 0x6c,
	0x2e, 0x70, 0x72, 0x6f, 0x78, 0x79, 0x2e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x52,
	0x08, 0x72, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x1a, 0xd1, 0x03, 0x0a, 0x07, 0x4f, 0x70,
	0x74, 0x69, 0x6f, 0x6e, 0x73, 0x12, 0x20, 0x0a, 0x0b, 0x65, 0x6e, 0x61, 0x62, 0x6c, 0x65, 0x50,
	0x72, 0x6f, 0x78, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0b, 0x65, 0x6e, 0x61, 0x62,
	0x6c, 0x65, 0x50, 0x72, 0x6f, 0x78, 0x79, 0x12, 0x54, 0x0a, 0x0b, 0x72, 0x65, 0x6c, 0x69, 0x61,
	0x62, 0x69, 0x6c, 0x69, 0x74, 0x79, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x32, 0x2e, 0x63,
	0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x2e, 0x73, 0x6d, 0x65, 0x6c, 0x74, 0x65, 0x72,
	0x2e, 0x76, 0x31, 0x2e, 0x63, 0x72, 0x61, 0x77, 0x6c, 0x2e, 0x70, 0x72, 0x6f, 0x78, 0x79, 0x2e,
	0x50, 0x72, 0x6f, 0x78, 0x79, 0x52, 0x65, 0x6c, 0x69, 0x61, 0x62, 0x69, 0x6c, 0x69, 0x74, 0x79,
	0x52, 0x0b, 0x72, 0x65, 0x6c, 0x69, 0x61, 0x62, 0x69, 0x6c, 0x69, 0x74, 0x79, 0x12, 0x26, 0x0a,
	0x0e, 0x65, 0x6e, 0x61, 0x62, 0x6c, 0x65, 0x48, 0x65, 0x61, 0x64, 0x6c, 0x65, 0x73, 0x73, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0e, 0x65, 0x6e, 0x61, 0x62, 0x6c, 0x65, 0x48, 0x65, 0x61,
	0x64, 0x6c, 0x65, 0x73, 0x73, 0x12, 0x26, 0x0a, 0x0e, 0x6a, 0x73, 0x57, 0x61, 0x69, 0x74, 0x44,
	0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x04, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0e, 0x6a,
	0x73, 0x57, 0x61, 0x69, 0x74, 0x44, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x2c, 0x0a,
	0x11, 0x65, 0x6e, 0x61, 0x62, 0x6c, 0x65, 0x53, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x49, 0x6e,
	0x69, 0x74, 0x18, 0x05, 0x20, 0x01, 0x28, 0x08, 0x52, 0x11, 0x65, 0x6e, 0x61, 0x62, 0x6c, 0x65,
	0x53, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x49, 0x6e, 0x69, 0x74, 0x12, 0x20, 0x0a, 0x0b, 0x6b,
	0x65, 0x65, 0x70, 0x53, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x06, 0x20, 0x01, 0x28, 0x08,
	0x52, 0x0b, 0x6b, 0x65, 0x65, 0x70, 0x53, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x2a, 0x0a,
	0x10, 0x64, 0x69, 0x73, 0x61, 0x62, 0x6c, 0x65, 0x43, 0x6f, 0x6f, 0x6b, 0x69, 0x65, 0x4a, 0x61,
	0x72, 0x18, 0x07, 0x20, 0x01, 0x28, 0x08, 0x52, 0x10, 0x64, 0x69, 0x73, 0x61, 0x62, 0x6c, 0x65,
	0x43, 0x6f, 0x6f, 0x6b, 0x69, 0x65, 0x4a, 0x61, 0x72, 0x12, 0x2a, 0x0a, 0x10, 0x6d, 0x61, 0x78,
	0x54, 0x74, 0x6c, 0x50, 0x65, 0x72, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x18, 0x08, 0x20,
	0x01, 0x28, 0x03, 0x52, 0x10, 0x6d, 0x61, 0x78, 0x54, 0x74, 0x6c, 0x50, 0x65, 0x72, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x28, 0x0a, 0x0f, 0x64, 0x69, 0x73, 0x61, 0x62, 0x6c, 0x65,
	0x52, 0x65, 0x64, 0x69, 0x72, 0x65, 0x63, 0x74, 0x18, 0x0b, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0f,
	0x64, 0x69, 0x73, 0x61, 0x62, 0x6c, 0x65, 0x52, 0x65, 0x64, 0x69, 0x72, 0x65, 0x63, 0x74, 0x12,
	0x2c, 0x0a, 0x11, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x46, 0x69, 0x6c, 0x74, 0x65, 0x72,
	0x4b, 0x65, 0x79, 0x73, 0x18, 0x0f, 0x20, 0x03, 0x28, 0x09, 0x52, 0x11, 0x72, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x46, 0x69, 0x6c, 0x74, 0x65, 0x72, 0x4b, 0x65, 0x79, 0x73, 0x1a, 0x59, 0x0a,
	0x0c, 0x48, 0x65, 0x61, 0x64, 0x65, 0x72, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a,
	0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12,
	0x33, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1d,
	0x2e, 0x63, 0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x68,
	0x74, 0x74, 0x70, 0x2e, 0x4c, 0x69, 0x73, 0x74, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x52, 0x05, 0x76,
	0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x22, 0x8b, 0x04, 0x0a, 0x08, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x1e, 0x0a, 0x0a, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x43,
	0x6f, 0x64, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0a, 0x73, 0x74, 0x61, 0x74, 0x75,
	0x73, 0x43, 0x6f, 0x64, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x14, 0x0a,
	0x05, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x12, 0x1e, 0x0a, 0x0a, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x4d, 0x61, 0x6a, 0x6f,
	0x72, 0x18, 0x04, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0a, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x4d, 0x61,
	0x6a, 0x6f, 0x72, 0x12, 0x1e, 0x0a, 0x0a, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x4d, 0x69, 0x6e, 0x6f,
	0x72, 0x18, 0x05, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0a, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x4d, 0x69,
	0x6e, 0x6f, 0x72, 0x12, 0x51, 0x0a, 0x07, 0x68, 0x65, 0x61, 0x64, 0x65, 0x72, 0x73, 0x18, 0x06,
	0x20, 0x03, 0x28, 0x0b, 0x32, 0x37, 0x2e, 0x63, 0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e,
	0x2e, 0x73, 0x6d, 0x65, 0x6c, 0x74, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x63, 0x72, 0x61, 0x77,
	0x6c, 0x2e, 0x70, 0x72, 0x6f, 0x78, 0x79, 0x2e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x2e, 0x48, 0x65, 0x61, 0x64, 0x65, 0x72, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x07, 0x68,
	0x65, 0x61, 0x64, 0x65, 0x72, 0x73, 0x12, 0x12, 0x0a, 0x04, 0x62, 0x6f, 0x64, 0x79, 0x18, 0x09,
	0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x62, 0x6f, 0x64, 0x79, 0x12, 0x24, 0x0a, 0x0d, 0x62, 0x6f,
	0x64, 0x79, 0x43, 0x61, 0x63, 0x68, 0x65, 0x4c, 0x69, 0x6e, 0x6b, 0x18, 0x0b, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x0d, 0x62, 0x6f, 0x64, 0x79, 0x43, 0x61, 0x63, 0x68, 0x65, 0x4c, 0x69, 0x6e, 0x6b,
	0x12, 0x1a, 0x0a, 0x08, 0x64, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x0c, 0x20, 0x01,
	0x28, 0x03, 0x52, 0x08, 0x64, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x28, 0x0a, 0x0f,
	0x61, 0x76, 0x65, 0x72, 0x61, 0x67, 0x65, 0x44, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x18,
	0x0d, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0f, 0x61, 0x76, 0x65, 0x72, 0x61, 0x67, 0x65, 0x44, 0x75,
	0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x43, 0x0a, 0x07, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x18, 0x0f, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x29, 0x2e, 0x63, 0x68, 0x61, 0x6d, 0x65, 0x6c,
	0x65, 0x6f, 0x6e, 0x2e, 0x73, 0x6d, 0x65, 0x6c, 0x74, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x63,
	0x72, 0x61, 0x77, 0x6c, 0x2e, 0x70, 0x72, 0x6f, 0x78, 0x79, 0x2e, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x52, 0x07, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x59, 0x0a, 0x0c, 0x48,
	0x65, 0x61, 0x64, 0x65, 0x72, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b,
	0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x33, 0x0a,
	0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1d, 0x2e, 0x63,
	0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x68, 0x74, 0x74,
	0x70, 0x2e, 0x4c, 0x69, 0x73, 0x74, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x52, 0x05, 0x76, 0x61, 0x6c,
	0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x22, 0xb5, 0x02, 0x0a, 0x0b, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x57, 0x72, 0x61, 0x70, 0x12, 0x14, 0x0a, 0x05, 0x72, 0x65, 0x71, 0x49, 0x64, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x72, 0x65, 0x71, 0x49, 0x64, 0x12, 0x43, 0x0a, 0x07,
	0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x29, 0x2e,
	0x63, 0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x2e, 0x73, 0x6d, 0x65, 0x6c, 0x74, 0x65,
	0x72, 0x2e, 0x76, 0x31, 0x2e, 0x63, 0x72, 0x61, 0x77, 0x6c, 0x2e, 0x70, 0x72, 0x6f, 0x78, 0x79,
	0x2e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x52, 0x07, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x12, 0x1a, 0x0a, 0x08, 0x64, 0x65, 0x61, 0x64, 0x6c, 0x69, 0x6e, 0x65, 0x18, 0x06, 0x20,
	0x01, 0x28, 0x03, 0x52, 0x08, 0x64, 0x65, 0x61, 0x64, 0x6c, 0x69, 0x6e, 0x65, 0x12, 0x1c, 0x0a,
	0x09, 0x65, 0x78, 0x65, 0x63, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x07, 0x20, 0x01, 0x28, 0x05,
	0x52, 0x09, 0x65, 0x78, 0x65, 0x63, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x1a, 0x0a, 0x08, 0x64,
	0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x0b, 0x20, 0x01, 0x28, 0x03, 0x52, 0x08, 0x64,
	0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x28, 0x0a, 0x0f, 0x61, 0x76, 0x65, 0x72, 0x61,
	0x67, 0x65, 0x44, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x0c, 0x20, 0x01, 0x28, 0x03,
	0x52, 0x0f, 0x61, 0x76, 0x65, 0x72, 0x61, 0x67, 0x65, 0x44, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x12, 0x4b, 0x0a, 0x07, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x0f, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x31, 0x2e, 0x63, 0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x2e, 0x73,
	0x6d, 0x65, 0x6c, 0x74, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x63, 0x72, 0x61, 0x77, 0x6c, 0x2e,
	0x70, 0x72, 0x6f, 0x78, 0x79, 0x2e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x2e, 0x4f, 0x70,
	0x74, 0x69, 0x6f, 0x6e, 0x73, 0x52, 0x07, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2a, 0x86,
	0x01, 0x0a, 0x10, 0x50, 0x72, 0x6f, 0x78, 0x79, 0x52, 0x65, 0x6c, 0x69, 0x61, 0x62, 0x69, 0x6c,
	0x69, 0x74, 0x79, 0x12, 0x16, 0x0a, 0x12, 0x52, 0x65, 0x6c, 0x69, 0x61, 0x62, 0x69, 0x6c, 0x69,
	0x74, 0x79, 0x44, 0x65, 0x66, 0x61, 0x75, 0x6c, 0x74, 0x10, 0x00, 0x12, 0x12, 0x0a, 0x0e, 0x52,
	0x65, 0x6c, 0x69, 0x61, 0x62, 0x69, 0x6c, 0x69, 0x74, 0x79, 0x4c, 0x6f, 0x77, 0x10, 0x01, 0x12,
	0x15, 0x0a, 0x11, 0x52, 0x65, 0x6c, 0x69, 0x61, 0x62, 0x69, 0x6c, 0x69, 0x74, 0x79, 0x4d, 0x65,
	0x64, 0x69, 0x75, 0x6d, 0x10, 0x02, 0x12, 0x13, 0x0a, 0x0f, 0x52, 0x65, 0x6c, 0x69, 0x61, 0x62,
	0x69, 0x6c, 0x69, 0x74, 0x79, 0x48, 0x69, 0x67, 0x68, 0x10, 0x03, 0x12, 0x1a, 0x0a, 0x16, 0x52,
	0x65, 0x6c, 0x69, 0x61, 0x62, 0x69, 0x6c, 0x69, 0x74, 0x79, 0x49, 0x6e, 0x74, 0x65, 0x6c, 0x6c,
	0x69, 0x67, 0x65, 0x6e, 0x74, 0x10, 0x0a, 0x42, 0x28, 0x5a, 0x26, 0x63, 0x68, 0x61, 0x6d, 0x65,
	0x6c, 0x65, 0x6f, 0x6e, 0x2f, 0x73, 0x6d, 0x65, 0x6c, 0x74, 0x65, 0x72, 0x2f, 0x76, 0x31, 0x2f,
	0x63, 0x72, 0x61, 0x77, 0x6c, 0x2f, 0x70, 0x72, 0x6f, 0x78, 0x79, 0x3b, 0x70, 0x72, 0x6f, 0x78,
	0x79, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_chameleon_smelter_v1_crawl_proxy_data_proto_rawDescOnce sync.Once
	file_chameleon_smelter_v1_crawl_proxy_data_proto_rawDescData = file_chameleon_smelter_v1_crawl_proxy_data_proto_rawDesc
)

func file_chameleon_smelter_v1_crawl_proxy_data_proto_rawDescGZIP() []byte {
	file_chameleon_smelter_v1_crawl_proxy_data_proto_rawDescOnce.Do(func() {
		file_chameleon_smelter_v1_crawl_proxy_data_proto_rawDescData = protoimpl.X.CompressGZIP(file_chameleon_smelter_v1_crawl_proxy_data_proto_rawDescData)
	})
	return file_chameleon_smelter_v1_crawl_proxy_data_proto_rawDescData
}

var file_chameleon_smelter_v1_crawl_proxy_data_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_chameleon_smelter_v1_crawl_proxy_data_proto_msgTypes = make([]protoimpl.MessageInfo, 6)
var file_chameleon_smelter_v1_crawl_proxy_data_proto_goTypes = []interface{}{
	(ProxyReliability)(0),   // 0: chameleon.smelter.v1.crawl.proxy.ProxyReliability
	(*Request)(nil),         // 1: chameleon.smelter.v1.crawl.proxy.Request
	(*Response)(nil),        // 2: chameleon.smelter.v1.crawl.proxy.Response
	(*RequestWrap)(nil),     // 3: chameleon.smelter.v1.crawl.proxy.RequestWrap
	(*Request_Options)(nil), // 4: chameleon.smelter.v1.crawl.proxy.Request.Options
	nil,                     // 5: chameleon.smelter.v1.crawl.proxy.Request.HeadersEntry
	nil,                     // 6: chameleon.smelter.v1.crawl.proxy.Response.HeadersEntry
	(*http.ListValue)(nil),  // 7: chameleon.api.http.ListValue
}
var file_chameleon_smelter_v1_crawl_proxy_data_proto_depIdxs = []int32{
	5,  // 0: chameleon.smelter.v1.crawl.proxy.Request.headers:type_name -> chameleon.smelter.v1.crawl.proxy.Request.HeadersEntry
	4,  // 1: chameleon.smelter.v1.crawl.proxy.Request.options:type_name -> chameleon.smelter.v1.crawl.proxy.Request.Options
	2,  // 2: chameleon.smelter.v1.crawl.proxy.Request.response:type_name -> chameleon.smelter.v1.crawl.proxy.Response
	6,  // 3: chameleon.smelter.v1.crawl.proxy.Response.headers:type_name -> chameleon.smelter.v1.crawl.proxy.Response.HeadersEntry
	1,  // 4: chameleon.smelter.v1.crawl.proxy.Response.request:type_name -> chameleon.smelter.v1.crawl.proxy.Request
	1,  // 5: chameleon.smelter.v1.crawl.proxy.RequestWrap.request:type_name -> chameleon.smelter.v1.crawl.proxy.Request
	4,  // 6: chameleon.smelter.v1.crawl.proxy.RequestWrap.options:type_name -> chameleon.smelter.v1.crawl.proxy.Request.Options
	0,  // 7: chameleon.smelter.v1.crawl.proxy.Request.Options.reliability:type_name -> chameleon.smelter.v1.crawl.proxy.ProxyReliability
	7,  // 8: chameleon.smelter.v1.crawl.proxy.Request.HeadersEntry.value:type_name -> chameleon.api.http.ListValue
	7,  // 9: chameleon.smelter.v1.crawl.proxy.Response.HeadersEntry.value:type_name -> chameleon.api.http.ListValue
	10, // [10:10] is the sub-list for method output_type
	10, // [10:10] is the sub-list for method input_type
	10, // [10:10] is the sub-list for extension type_name
	10, // [10:10] is the sub-list for extension extendee
	0,  // [0:10] is the sub-list for field type_name
}

func init() { file_chameleon_smelter_v1_crawl_proxy_data_proto_init() }
func file_chameleon_smelter_v1_crawl_proxy_data_proto_init() {
	if File_chameleon_smelter_v1_crawl_proxy_data_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_chameleon_smelter_v1_crawl_proxy_data_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Request); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_chameleon_smelter_v1_crawl_proxy_data_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Response); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_chameleon_smelter_v1_crawl_proxy_data_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RequestWrap); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_chameleon_smelter_v1_crawl_proxy_data_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Request_Options); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_chameleon_smelter_v1_crawl_proxy_data_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   6,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_chameleon_smelter_v1_crawl_proxy_data_proto_goTypes,
		DependencyIndexes: file_chameleon_smelter_v1_crawl_proxy_data_proto_depIdxs,
		EnumInfos:         file_chameleon_smelter_v1_crawl_proxy_data_proto_enumTypes,
		MessageInfos:      file_chameleon_smelter_v1_crawl_proxy_data_proto_msgTypes,
	}.Build()
	File_chameleon_smelter_v1_crawl_proxy_data_proto = out.File
	file_chameleon_smelter_v1_crawl_proxy_data_proto_rawDesc = nil
	file_chameleon_smelter_v1_crawl_proxy_data_proto_goTypes = nil
	file_chameleon_smelter_v1_crawl_proxy_data_proto_depIdxs = nil
}
