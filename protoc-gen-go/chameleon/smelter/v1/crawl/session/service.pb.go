// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0-devel
// 	protoc        v3.14.0
// source: chameleon/smelter/v1/crawl/session/service.proto

package session

import (

	_ "github.com/voiladev/protobuf/protoc-gen-go/openapiv2/options"
	_ "github.com/voiladev/protobuf/protoc-gen-go/protobuf/google/api"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	_ "google.golang.org/protobuf/types/known/anypb"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	reflect "reflect"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

var File_chameleon_smelter_v1_crawl_session_service_proto protoreflect.FileDescriptor

var file_chameleon_smelter_v1_crawl_session_service_proto_rawDesc = []byte{
	0x0a, 0x30, 0x63, 0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x2f, 0x73, 0x6d, 0x65, 0x6c,
	0x74, 0x65, 0x72, 0x2f, 0x76, 0x31, 0x2f, 0x63, 0x72, 0x61, 0x77, 0x6c, 0x2f, 0x73, 0x65, 0x73,
	0x73, 0x69, 0x6f, 0x6e, 0x2f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x12, 0x22, 0x63, 0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x2e, 0x73, 0x6d,
	0x65, 0x6c, 0x74, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x63, 0x72, 0x61, 0x77, 0x6c, 0x2e, 0x73,
	0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x1a, 0x25, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
	0x2f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x61, 0x6e, 0x6e, 0x6f,
	0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1b, 0x67,
	0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x65,
	0x6d, 0x70, 0x74, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x19, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x61, 0x6e, 0x79, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x23, 0x6f, 0x70, 0x65, 0x6e, 0x61, 0x70, 0x69, 0x76, 0x32,
	0x2f, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x38, 0x63, 0x68, 0x61, 0x6d,
	0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x2f, 0x73, 0x6d, 0x65, 0x6c, 0x74, 0x65, 0x72, 0x2f, 0x76, 0x31,
	0x2f, 0x63, 0x72, 0x61, 0x77, 0x6c, 0x2f, 0x73, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x2f, 0x73,
	0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x5f, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x32, 0x9a, 0x03, 0x0a, 0x0e, 0x53, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e,
	0x4d, 0x61, 0x6e, 0x61, 0x67, 0x65, 0x72, 0x12, 0x9f, 0x01, 0x0a, 0x0a, 0x47, 0x65, 0x74, 0x43,
	0x6f, 0x6f, 0x6b, 0x69, 0x65, 0x73, 0x12, 0x35, 0x2e, 0x63, 0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65,
	0x6f, 0x6e, 0x2e, 0x73, 0x6d, 0x65, 0x6c, 0x74, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x63, 0x72,
	0x61, 0x77, 0x6c, 0x2e, 0x73, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x2e, 0x47, 0x65, 0x74, 0x43,
	0x6f, 0x6f, 0x6b, 0x69, 0x65, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x36, 0x2e,
	0x63, 0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x2e, 0x73, 0x6d, 0x65, 0x6c, 0x74, 0x65,
	0x72, 0x2e, 0x76, 0x31, 0x2e, 0x63, 0x72, 0x61, 0x77, 0x6c, 0x2e, 0x73, 0x65, 0x73, 0x73, 0x69,
	0x6f, 0x6e, 0x2e, 0x47, 0x65, 0x74, 0x43, 0x6f, 0x6f, 0x6b, 0x69, 0x65, 0x73, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x22, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x1c, 0x12, 0x1a, 0x2f,
	0x73, 0x6d, 0x65, 0x6c, 0x74, 0x65, 0x72, 0x2f, 0x76, 0x31, 0x2f, 0x63, 0x72, 0x61, 0x77, 0x6c,
	0x2f, 0x73, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x73, 0x12, 0x82, 0x01, 0x0a, 0x0a, 0x53, 0x65,
	0x74, 0x43, 0x6f, 0x6f, 0x6b, 0x69, 0x65, 0x73, 0x12, 0x35, 0x2e, 0x63, 0x68, 0x61, 0x6d, 0x65,
	0x6c, 0x65, 0x6f, 0x6e, 0x2e, 0x73, 0x6d, 0x65, 0x6c, 0x74, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e,
	0x63, 0x72, 0x61, 0x77, 0x6c, 0x2e, 0x73, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x2e, 0x53, 0x65,
	0x74, 0x43, 0x6f, 0x6f, 0x6b, 0x69, 0x65, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a,
	0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75,
	0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x22, 0x25, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x1f, 0x22,
	0x1a, 0x2f, 0x73, 0x6d, 0x65, 0x6c, 0x74, 0x65, 0x72, 0x2f, 0x76, 0x31, 0x2f, 0x63, 0x72, 0x61,
	0x77, 0x6c, 0x2f, 0x73, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x73, 0x3a, 0x01, 0x2a, 0x12, 0x61,
	0x0a, 0x0c, 0x43, 0x6c, 0x65, 0x61, 0x72, 0x43, 0x6f, 0x6f, 0x6b, 0x69, 0x65, 0x73, 0x12, 0x37,
	0x2e, 0x63, 0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x2e, 0x73, 0x6d, 0x65, 0x6c, 0x74,
	0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x63, 0x72, 0x61, 0x77, 0x6c, 0x2e, 0x73, 0x65, 0x73, 0x73,
	0x69, 0x6f, 0x6e, 0x2e, 0x43, 0x6c, 0x65, 0x61, 0x72, 0x43, 0x6f, 0x6f, 0x6b, 0x69, 0x65, 0x73,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x22,
	0x00, 0x42, 0x93, 0x02, 0x5a, 0x2a, 0x63, 0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x2f,
	0x73, 0x6d, 0x65, 0x6c, 0x74, 0x65, 0x72, 0x2f, 0x76, 0x31, 0x2f, 0x63, 0x72, 0x61, 0x77, 0x6c,
	0x2f, 0x73, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x3b, 0x73, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e,
	0x92, 0x41, 0xe3, 0x01, 0x12, 0x6b, 0x0a, 0x18, 0x73, 0x6d, 0x65, 0x6c, 0x74, 0x65, 0x72, 0x2e,
	0x76, 0x31, 0x2e, 0x63, 0x72, 0x61, 0x77, 0x6c, 0x2e, 0x73, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e,
	0x22, 0x42, 0x0a, 0x04, 0x53, 0x65, 0x65, 0x72, 0x12, 0x28, 0x68, 0x74, 0x74, 0x70, 0x73, 0x3a,
	0x2f, 0x2f, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x63, 0x68, 0x61,
	0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x64, 0x65, 0x76, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62,
	0x75, 0x66, 0x1a, 0x10, 0x6b, 0x76, 0x63, 0x6e, 0x6f, 0x77, 0x40, 0x67, 0x6d, 0x61, 0x69, 0x6c,
	0x2e, 0x63, 0x6f, 0x6d, 0x32, 0x0b, 0x5f, 0x5f, 0x56, 0x45, 0x52, 0x53, 0x49, 0x4f, 0x4e, 0x5f,
	0x5f, 0x2a, 0x02, 0x01, 0x02, 0x32, 0x10, 0x61, 0x70, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x69,
	0x6f, 0x6e, 0x2f, 0x6a, 0x73, 0x6f, 0x6e, 0x3a, 0x10, 0x61, 0x70, 0x70, 0x6c, 0x69, 0x63, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x2f, 0x6a, 0x73, 0x6f, 0x6e, 0x52, 0x12, 0x0a, 0x03, 0x34, 0x30, 0x33,
	0x12, 0x0b, 0x0a, 0x09, 0xe6, 0x9c, 0xaa, 0xe6, 0x8e, 0x88, 0xe6, 0x9d, 0x83, 0x52, 0x21, 0x0a,
	0x03, 0x34, 0x30, 0x34, 0x12, 0x1a, 0x0a, 0x18, 0xe6, 0x9c, 0xaa, 0xe6, 0x89, 0xbe, 0xe5, 0x88,
	0xb0, 0xe6, 0x88, 0x96, 0xe8, 0x80, 0x85, 0xe4, 0xb8, 0x8d, 0xe5, 0xad, 0x98, 0xe5, 0x9c, 0xa8,
	0x52, 0x15, 0x0a, 0x03, 0x35, 0x30, 0x30, 0x12, 0x0e, 0x0a, 0x0c, 0xe7, 0xb3, 0xbb, 0xe7, 0xbb,
	0x9f, 0xe9, 0x94, 0x99, 0xe8, 0xaf, 0xaf, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,

}

var file_chameleon_smelter_v1_crawl_session_service_proto_goTypes = []interface{}{
	(*GetCookiesRequest)(nil),   // 0: chameleon.smelter.v1.crawl.session.GetCookiesRequest
	(*SetCookiesRequest)(nil),   // 1: chameleon.smelter.v1.crawl.session.SetCookiesRequest
	(*ClearCookiesRequest)(nil), // 2: chameleon.smelter.v1.crawl.session.ClearCookiesRequest
	(*GetCookiesResponse)(nil),  // 3: chameleon.smelter.v1.crawl.session.GetCookiesResponse
	(*emptypb.Empty)(nil),       // 4: google.protobuf.Empty
}
var file_chameleon_smelter_v1_crawl_session_service_proto_depIdxs = []int32{
	0, // 0: chameleon.smelter.v1.crawl.session.SessionManager.GetCookies:input_type -> chameleon.smelter.v1.crawl.session.GetCookiesRequest
	1, // 1: chameleon.smelter.v1.crawl.session.SessionManager.SetCookies:input_type -> chameleon.smelter.v1.crawl.session.SetCookiesRequest
	2, // 2: chameleon.smelter.v1.crawl.session.SessionManager.ClearCookies:input_type -> chameleon.smelter.v1.crawl.session.ClearCookiesRequest
	3, // 3: chameleon.smelter.v1.crawl.session.SessionManager.GetCookies:output_type -> chameleon.smelter.v1.crawl.session.GetCookiesResponse
	4, // 4: chameleon.smelter.v1.crawl.session.SessionManager.SetCookies:output_type -> google.protobuf.Empty
	4, // 5: chameleon.smelter.v1.crawl.session.SessionManager.ClearCookies:output_type -> google.protobuf.Empty
	3, // [3:6] is the sub-list for method output_type
	0, // [0:3] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_chameleon_smelter_v1_crawl_session_service_proto_init() }
func file_chameleon_smelter_v1_crawl_session_service_proto_init() {
	if File_chameleon_smelter_v1_crawl_session_service_proto != nil {
		return
	}
	file_chameleon_smelter_v1_crawl_session_service_message_proto_init()
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_chameleon_smelter_v1_crawl_session_service_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   0,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_chameleon_smelter_v1_crawl_session_service_proto_goTypes,
		DependencyIndexes: file_chameleon_smelter_v1_crawl_session_service_proto_depIdxs,
	}.Build()
	File_chameleon_smelter_v1_crawl_session_service_proto = out.File
	file_chameleon_smelter_v1_crawl_session_service_proto_rawDesc = nil
	file_chameleon_smelter_v1_crawl_session_service_proto_goTypes = nil
	file_chameleon_smelter_v1_crawl_session_service_proto_depIdxs = nil
}
