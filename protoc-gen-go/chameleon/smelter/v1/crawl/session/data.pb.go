// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0-devel
// 	protoc        v3.14.0
// source: chameleon/smelter/v1/crawl/session/data.proto

package session

import (
	http "github.com/voiladev/VoilaCrawler/protoc-gen-go/chameleon/api/http"
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

// Session
type Session struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Id
	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	// Url
	Url string `protobuf:"bytes,3,opt,name=url,proto3" json:"url,omitempty"`
	// TracingId
	TracingId string `protobuf:"bytes,6,opt,name=tracingId,proto3" json:"tracingId,omitempty"`
	// Cookies
	Cookies map[string]*http.Cookie `protobuf:"bytes,11,rep,name=cookies,proto3" json:"cookies,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *Session) Reset() {
	*x = Session{}
	if protoimpl.UnsafeEnabled {
		mi := &file_chameleon_smelter_v1_crawl_session_data_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Session) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Session) ProtoMessage() {}

func (x *Session) ProtoReflect() protoreflect.Message {
	mi := &file_chameleon_smelter_v1_crawl_session_data_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Session.ProtoReflect.Descriptor instead.
func (*Session) Descriptor() ([]byte, []int) {
	return file_chameleon_smelter_v1_crawl_session_data_proto_rawDescGZIP(), []int{0}
}

func (x *Session) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *Session) GetUrl() string {
	if x != nil {
		return x.Url
	}
	return ""
}

func (x *Session) GetTracingId() string {
	if x != nil {
		return x.TracingId
	}
	return ""
}

func (x *Session) GetCookies() map[string]*http.Cookie {
	if x != nil {
		return x.Cookies
	}
	return nil
}

var File_chameleon_smelter_v1_crawl_session_data_proto protoreflect.FileDescriptor

var file_chameleon_smelter_v1_crawl_session_data_proto_rawDesc = []byte{
	0x0a, 0x2d, 0x63, 0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x2f, 0x73, 0x6d, 0x65, 0x6c,
	0x74, 0x65, 0x72, 0x2f, 0x76, 0x31, 0x2f, 0x63, 0x72, 0x61, 0x77, 0x6c, 0x2f, 0x73, 0x65, 0x73,
	0x73, 0x69, 0x6f, 0x6e, 0x2f, 0x64, 0x61, 0x74, 0x61, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x22, 0x63, 0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x2e, 0x73, 0x6d, 0x65, 0x6c, 0x74,
	0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x63, 0x72, 0x61, 0x77, 0x6c, 0x2e, 0x73, 0x65, 0x73, 0x73,
	0x69, 0x6f, 0x6e, 0x1a, 0x1d, 0x63, 0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x2f, 0x61,
	0x70, 0x69, 0x2f, 0x68, 0x74, 0x74, 0x70, 0x2f, 0x64, 0x61, 0x74, 0x61, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x22, 0xf5, 0x01, 0x0a, 0x07, 0x53, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x0e,
	0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x10,
	0x0a, 0x03, 0x75, 0x72, 0x6c, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x75, 0x72, 0x6c,
	0x12, 0x1c, 0x0a, 0x09, 0x74, 0x72, 0x61, 0x63, 0x69, 0x6e, 0x67, 0x49, 0x64, 0x18, 0x06, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x09, 0x74, 0x72, 0x61, 0x63, 0x69, 0x6e, 0x67, 0x49, 0x64, 0x12, 0x52,
	0x0a, 0x07, 0x63, 0x6f, 0x6f, 0x6b, 0x69, 0x65, 0x73, 0x18, 0x0b, 0x20, 0x03, 0x28, 0x0b, 0x32,
	0x38, 0x2e, 0x63, 0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x2e, 0x73, 0x6d, 0x65, 0x6c,
	0x74, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x63, 0x72, 0x61, 0x77, 0x6c, 0x2e, 0x73, 0x65, 0x73,
	0x73, 0x69, 0x6f, 0x6e, 0x2e, 0x53, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x2e, 0x43, 0x6f, 0x6f,
	0x6b, 0x69, 0x65, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x07, 0x63, 0x6f, 0x6f, 0x6b, 0x69,
	0x65, 0x73, 0x1a, 0x56, 0x0a, 0x0c, 0x43, 0x6f, 0x6f, 0x6b, 0x69, 0x65, 0x73, 0x45, 0x6e, 0x74,
	0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x03, 0x6b, 0x65, 0x79, 0x12, 0x30, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x63, 0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x2e,
	0x61, 0x70, 0x69, 0x2e, 0x68, 0x74, 0x74, 0x70, 0x2e, 0x43, 0x6f, 0x6f, 0x6b, 0x69, 0x65, 0x52,
	0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x42, 0x2c, 0x5a, 0x2a, 0x63, 0x68,
	0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x2f, 0x73, 0x6d, 0x65, 0x6c, 0x74, 0x65, 0x72, 0x2f,
	0x76, 0x31, 0x2f, 0x63, 0x72, 0x61, 0x77, 0x6c, 0x2f, 0x73, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e,
	0x3b, 0x73, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_chameleon_smelter_v1_crawl_session_data_proto_rawDescOnce sync.Once
	file_chameleon_smelter_v1_crawl_session_data_proto_rawDescData = file_chameleon_smelter_v1_crawl_session_data_proto_rawDesc
)

func file_chameleon_smelter_v1_crawl_session_data_proto_rawDescGZIP() []byte {
	file_chameleon_smelter_v1_crawl_session_data_proto_rawDescOnce.Do(func() {
		file_chameleon_smelter_v1_crawl_session_data_proto_rawDescData = protoimpl.X.CompressGZIP(file_chameleon_smelter_v1_crawl_session_data_proto_rawDescData)
	})
	return file_chameleon_smelter_v1_crawl_session_data_proto_rawDescData
}

var file_chameleon_smelter_v1_crawl_session_data_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_chameleon_smelter_v1_crawl_session_data_proto_goTypes = []interface{}{
	(*Session)(nil),     // 0: chameleon.smelter.v1.crawl.session.Session
	nil,                 // 1: chameleon.smelter.v1.crawl.session.Session.CookiesEntry
	(*http.Cookie)(nil), // 2: chameleon.api.http.Cookie
}
var file_chameleon_smelter_v1_crawl_session_data_proto_depIdxs = []int32{
	1, // 0: chameleon.smelter.v1.crawl.session.Session.cookies:type_name -> chameleon.smelter.v1.crawl.session.Session.CookiesEntry
	2, // 1: chameleon.smelter.v1.crawl.session.Session.CookiesEntry.value:type_name -> chameleon.api.http.Cookie
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_chameleon_smelter_v1_crawl_session_data_proto_init() }
func file_chameleon_smelter_v1_crawl_session_data_proto_init() {
	if File_chameleon_smelter_v1_crawl_session_data_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_chameleon_smelter_v1_crawl_session_data_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Session); i {
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
			RawDescriptor: file_chameleon_smelter_v1_crawl_session_data_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_chameleon_smelter_v1_crawl_session_data_proto_goTypes,
		DependencyIndexes: file_chameleon_smelter_v1_crawl_session_data_proto_depIdxs,
		MessageInfos:      file_chameleon_smelter_v1_crawl_session_data_proto_msgTypes,
	}.Build()
	File_chameleon_smelter_v1_crawl_session_data_proto = out.File
	file_chameleon_smelter_v1_crawl_session_data_proto_rawDesc = nil
	file_chameleon_smelter_v1_crawl_session_data_proto_goTypes = nil
	file_chameleon_smelter_v1_crawl_session_data_proto_depIdxs = nil
}
