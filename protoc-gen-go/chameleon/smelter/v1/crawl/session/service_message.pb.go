// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0-devel
// 	protoc        v3.14.0
// source: chameleon/smelter/v1/crawl/session/service_message.proto

package session

import (
	http "github.com/voiladev/go-crawler/protoc-gen-go/chameleon/api/http"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	_ "google.golang.org/protobuf/types/known/anypb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// GetCookiesRequest
type GetCookiesRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// TracingId
	TracingId string `protobuf:"bytes,1,opt,name=tracingId,proto3" json:"tracingId,omitempty"`
	// Url @required
	Url string `protobuf:"bytes,2,opt,name=url,proto3" json:"url,omitempty"`
}

func (x *GetCookiesRequest) Reset() {
	*x = GetCookiesRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_chameleon_smelter_v1_crawl_session_service_message_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetCookiesRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetCookiesRequest) ProtoMessage() {}

func (x *GetCookiesRequest) ProtoReflect() protoreflect.Message {
	mi := &file_chameleon_smelter_v1_crawl_session_service_message_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetCookiesRequest.ProtoReflect.Descriptor instead.
func (*GetCookiesRequest) Descriptor() ([]byte, []int) {
	return file_chameleon_smelter_v1_crawl_session_service_message_proto_rawDescGZIP(), []int{0}
}

func (x *GetCookiesRequest) GetTracingId() string {
	if x != nil {
		return x.TracingId
	}
	return ""
}

func (x *GetCookiesRequest) GetUrl() string {
	if x != nil {
		return x.Url
	}
	return ""
}

// GetCookiesResponse
type GetCookiesResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Data
	Data []*http.Cookie `protobuf:"bytes,6,rep,name=data,proto3" json:"data,omitempty"`
}

func (x *GetCookiesResponse) Reset() {
	*x = GetCookiesResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_chameleon_smelter_v1_crawl_session_service_message_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetCookiesResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetCookiesResponse) ProtoMessage() {}

func (x *GetCookiesResponse) ProtoReflect() protoreflect.Message {
	mi := &file_chameleon_smelter_v1_crawl_session_service_message_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetCookiesResponse.ProtoReflect.Descriptor instead.
func (*GetCookiesResponse) Descriptor() ([]byte, []int) {
	return file_chameleon_smelter_v1_crawl_session_service_message_proto_rawDescGZIP(), []int{1}
}

func (x *GetCookiesResponse) GetData() []*http.Cookie {
	if x != nil {
		return x.Data
	}
	return nil
}

// SetCookiesRequest
type SetCookiesRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// TracingId
	TracingId string `protobuf:"bytes,1,opt,name=tracingId,proto3" json:"tracingId,omitempty"`
	// Url @required
	Url string `protobuf:"bytes,2,opt,name=url,proto3" json:"url,omitempty"`
	// Cookies @required
	Cookies []*http.Cookie `protobuf:"bytes,6,rep,name=cookies,proto3" json:"cookies,omitempty"`
}

func (x *SetCookiesRequest) Reset() {
	*x = SetCookiesRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_chameleon_smelter_v1_crawl_session_service_message_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SetCookiesRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SetCookiesRequest) ProtoMessage() {}

func (x *SetCookiesRequest) ProtoReflect() protoreflect.Message {
	mi := &file_chameleon_smelter_v1_crawl_session_service_message_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SetCookiesRequest.ProtoReflect.Descriptor instead.
func (*SetCookiesRequest) Descriptor() ([]byte, []int) {
	return file_chameleon_smelter_v1_crawl_session_service_message_proto_rawDescGZIP(), []int{2}
}

func (x *SetCookiesRequest) GetTracingId() string {
	if x != nil {
		return x.TracingId
	}
	return ""
}

func (x *SetCookiesRequest) GetUrl() string {
	if x != nil {
		return x.Url
	}
	return ""
}

func (x *SetCookiesRequest) GetCookies() []*http.Cookie {
	if x != nil {
		return x.Cookies
	}
	return nil
}

// ClearCookiesRequest
type ClearCookiesRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// TracingId
	TracingId string `protobuf:"bytes,1,opt,name=tracingId,proto3" json:"tracingId,omitempty"`
	// Url @required
	Url string `protobuf:"bytes,2,opt,name=url,proto3" json:"url,omitempty"`
}

func (x *ClearCookiesRequest) Reset() {
	*x = ClearCookiesRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_chameleon_smelter_v1_crawl_session_service_message_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ClearCookiesRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ClearCookiesRequest) ProtoMessage() {}

func (x *ClearCookiesRequest) ProtoReflect() protoreflect.Message {
	mi := &file_chameleon_smelter_v1_crawl_session_service_message_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ClearCookiesRequest.ProtoReflect.Descriptor instead.
func (*ClearCookiesRequest) Descriptor() ([]byte, []int) {
	return file_chameleon_smelter_v1_crawl_session_service_message_proto_rawDescGZIP(), []int{3}
}

func (x *ClearCookiesRequest) GetTracingId() string {
	if x != nil {
		return x.TracingId
	}
	return ""
}

func (x *ClearCookiesRequest) GetUrl() string {
	if x != nil {
		return x.Url
	}
	return ""
}

var File_chameleon_smelter_v1_crawl_session_service_message_proto protoreflect.FileDescriptor

var file_chameleon_smelter_v1_crawl_session_service_message_proto_rawDesc = []byte{
	0x0a, 0x38, 0x63, 0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x2f, 0x73, 0x6d, 0x65, 0x6c,
	0x74, 0x65, 0x72, 0x2f, 0x76, 0x31, 0x2f, 0x63, 0x72, 0x61, 0x77, 0x6c, 0x2f, 0x73, 0x65, 0x73,
	0x73, 0x69, 0x6f, 0x6e, 0x2f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x5f, 0x6d, 0x65, 0x73,
	0x73, 0x61, 0x67, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x22, 0x63, 0x68, 0x61, 0x6d,
	0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x2e, 0x73, 0x6d, 0x65, 0x6c, 0x74, 0x65, 0x72, 0x2e, 0x76, 0x31,
	0x2e, 0x63, 0x72, 0x61, 0x77, 0x6c, 0x2e, 0x73, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x1a, 0x19,
	0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f,
	0x61, 0x6e, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1d, 0x63, 0x68, 0x61, 0x6d, 0x65,
	0x6c, 0x65, 0x6f, 0x6e, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x68, 0x74, 0x74, 0x70, 0x2f, 0x64, 0x61,
	0x74, 0x61, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x2d, 0x63, 0x68, 0x61, 0x6d, 0x65, 0x6c,
	0x65, 0x6f, 0x6e, 0x2f, 0x73, 0x6d, 0x65, 0x6c, 0x74, 0x65, 0x72, 0x2f, 0x76, 0x31, 0x2f, 0x63,
	0x72, 0x61, 0x77, 0x6c, 0x2f, 0x73, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x2f, 0x64, 0x61, 0x74,
	0x61, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x43, 0x0a, 0x11, 0x47, 0x65, 0x74, 0x43, 0x6f,
	0x6f, 0x6b, 0x69, 0x65, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1c, 0x0a, 0x09,
	0x74, 0x72, 0x61, 0x63, 0x69, 0x6e, 0x67, 0x49, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x09, 0x74, 0x72, 0x61, 0x63, 0x69, 0x6e, 0x67, 0x49, 0x64, 0x12, 0x10, 0x0a, 0x03, 0x75, 0x72,
	0x6c, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x75, 0x72, 0x6c, 0x22, 0x44, 0x0a, 0x12,
	0x47, 0x65, 0x74, 0x43, 0x6f, 0x6f, 0x6b, 0x69, 0x65, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x12, 0x2e, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x06, 0x20, 0x03, 0x28, 0x0b,
	0x32, 0x1a, 0x2e, 0x63, 0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x2e, 0x61, 0x70, 0x69,
	0x2e, 0x68, 0x74, 0x74, 0x70, 0x2e, 0x43, 0x6f, 0x6f, 0x6b, 0x69, 0x65, 0x52, 0x04, 0x64, 0x61,
	0x74, 0x61, 0x22, 0x79, 0x0a, 0x11, 0x53, 0x65, 0x74, 0x43, 0x6f, 0x6f, 0x6b, 0x69, 0x65, 0x73,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1c, 0x0a, 0x09, 0x74, 0x72, 0x61, 0x63, 0x69,
	0x6e, 0x67, 0x49, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x74, 0x72, 0x61, 0x63,
	0x69, 0x6e, 0x67, 0x49, 0x64, 0x12, 0x10, 0x0a, 0x03, 0x75, 0x72, 0x6c, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x03, 0x75, 0x72, 0x6c, 0x12, 0x34, 0x0a, 0x07, 0x63, 0x6f, 0x6f, 0x6b, 0x69,
	0x65, 0x73, 0x18, 0x06, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x63, 0x68, 0x61, 0x6d, 0x65,
	0x6c, 0x65, 0x6f, 0x6e, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x68, 0x74, 0x74, 0x70, 0x2e, 0x43, 0x6f,
	0x6f, 0x6b, 0x69, 0x65, 0x52, 0x07, 0x63, 0x6f, 0x6f, 0x6b, 0x69, 0x65, 0x73, 0x22, 0x45, 0x0a,
	0x13, 0x43, 0x6c, 0x65, 0x61, 0x72, 0x43, 0x6f, 0x6f, 0x6b, 0x69, 0x65, 0x73, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x12, 0x1c, 0x0a, 0x09, 0x74, 0x72, 0x61, 0x63, 0x69, 0x6e, 0x67, 0x49,
	0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x74, 0x72, 0x61, 0x63, 0x69, 0x6e, 0x67,
	0x49, 0x64, 0x12, 0x10, 0x0a, 0x03, 0x75, 0x72, 0x6c, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x03, 0x75, 0x72, 0x6c, 0x42, 0x2c, 0x5a, 0x2a, 0x63, 0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f,
	0x6e, 0x2f, 0x73, 0x6d, 0x65, 0x6c, 0x74, 0x65, 0x72, 0x2f, 0x76, 0x31, 0x2f, 0x63, 0x72, 0x61,
	0x77, 0x6c, 0x2f, 0x73, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x3b, 0x73, 0x65, 0x73, 0x73, 0x69,
	0x6f, 0x6e, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_chameleon_smelter_v1_crawl_session_service_message_proto_rawDescOnce sync.Once
	file_chameleon_smelter_v1_crawl_session_service_message_proto_rawDescData = file_chameleon_smelter_v1_crawl_session_service_message_proto_rawDesc
)

func file_chameleon_smelter_v1_crawl_session_service_message_proto_rawDescGZIP() []byte {
	file_chameleon_smelter_v1_crawl_session_service_message_proto_rawDescOnce.Do(func() {
		file_chameleon_smelter_v1_crawl_session_service_message_proto_rawDescData = protoimpl.X.CompressGZIP(file_chameleon_smelter_v1_crawl_session_service_message_proto_rawDescData)
	})
	return file_chameleon_smelter_v1_crawl_session_service_message_proto_rawDescData
}

var file_chameleon_smelter_v1_crawl_session_service_message_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_chameleon_smelter_v1_crawl_session_service_message_proto_goTypes = []interface{}{
	(*GetCookiesRequest)(nil),   // 0: chameleon.smelter.v1.crawl.session.GetCookiesRequest
	(*GetCookiesResponse)(nil),  // 1: chameleon.smelter.v1.crawl.session.GetCookiesResponse
	(*SetCookiesRequest)(nil),   // 2: chameleon.smelter.v1.crawl.session.SetCookiesRequest
	(*ClearCookiesRequest)(nil), // 3: chameleon.smelter.v1.crawl.session.ClearCookiesRequest
	(*http.Cookie)(nil),         // 4: chameleon.api.http.Cookie
}
var file_chameleon_smelter_v1_crawl_session_service_message_proto_depIdxs = []int32{
	4, // 0: chameleon.smelter.v1.crawl.session.GetCookiesResponse.data:type_name -> chameleon.api.http.Cookie
	4, // 1: chameleon.smelter.v1.crawl.session.SetCookiesRequest.cookies:type_name -> chameleon.api.http.Cookie
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_chameleon_smelter_v1_crawl_session_service_message_proto_init() }
func file_chameleon_smelter_v1_crawl_session_service_message_proto_init() {
	if File_chameleon_smelter_v1_crawl_session_service_message_proto != nil {
		return
	}
	file_chameleon_smelter_v1_crawl_session_data_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_chameleon_smelter_v1_crawl_session_service_message_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetCookiesRequest); i {
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
		file_chameleon_smelter_v1_crawl_session_service_message_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetCookiesResponse); i {
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
		file_chameleon_smelter_v1_crawl_session_service_message_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SetCookiesRequest); i {
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
		file_chameleon_smelter_v1_crawl_session_service_message_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ClearCookiesRequest); i {
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
			RawDescriptor: file_chameleon_smelter_v1_crawl_session_service_message_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_chameleon_smelter_v1_crawl_session_service_message_proto_goTypes,
		DependencyIndexes: file_chameleon_smelter_v1_crawl_session_service_message_proto_depIdxs,
		MessageInfos:      file_chameleon_smelter_v1_crawl_session_service_message_proto_msgTypes,
	}.Build()
	File_chameleon_smelter_v1_crawl_session_service_message_proto = out.File
	file_chameleon_smelter_v1_crawl_session_service_message_proto_rawDesc = nil
	file_chameleon_smelter_v1_crawl_session_service_message_proto_goTypes = nil
	file_chameleon_smelter_v1_crawl_session_service_message_proto_depIdxs = nil
}
