// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0-devel
// 	protoc        v3.14.0
// source: chameleon/smelter/v1/crawl/service.proto

package crawl

import (
	_ "github.com/voiladev/protobuf/protoc-gen-go/openapiv2/options"
	_ "google.golang.org/genproto/googleapis/api/annotations"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	anypb "google.golang.org/protobuf/types/known/anypb"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	reflect "reflect"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

var File_chameleon_smelter_v1_crawl_service_proto protoreflect.FileDescriptor

var file_chameleon_smelter_v1_crawl_service_proto_rawDesc = []byte{
	0x0a, 0x28, 0x63, 0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x2f, 0x73, 0x6d, 0x65, 0x6c,
	0x74, 0x65, 0x72, 0x2f, 0x76, 0x31, 0x2f, 0x63, 0x72, 0x61, 0x77, 0x6c, 0x2f, 0x73, 0x65, 0x72,
	0x76, 0x69, 0x63, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x1a, 0x63, 0x68, 0x61, 0x6d,
	0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x2e, 0x73, 0x6d, 0x65, 0x6c, 0x74, 0x65, 0x72, 0x2e, 0x76, 0x31,
	0x2e, 0x63, 0x72, 0x61, 0x77, 0x6c, 0x1a, 0x25, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
	0x2f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x61, 0x6e, 0x6e, 0x6f,
	0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1b, 0x67,
	0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x65,
	0x6d, 0x70, 0x74, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x19, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x61, 0x6e, 0x79, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x23, 0x6f, 0x70, 0x65, 0x6e, 0x61, 0x70, 0x69, 0x76, 0x32,
	0x2f, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x25, 0x63, 0x68, 0x61, 0x6d,
	0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x2f, 0x73, 0x6d, 0x65, 0x6c, 0x74, 0x65, 0x72, 0x2f, 0x76, 0x31,
	0x2f, 0x63, 0x72, 0x61, 0x77, 0x6c, 0x2f, 0x64, 0x61, 0x74, 0x61, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x1a, 0x30, 0x63, 0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x2f, 0x73, 0x6d, 0x65,
	0x6c, 0x74, 0x65, 0x72, 0x2f, 0x76, 0x31, 0x2f, 0x63, 0x72, 0x61, 0x77, 0x6c, 0x2f, 0x73, 0x65,
	0x72, 0x76, 0x69, 0x63, 0x65, 0x5f, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x32, 0xd4, 0x04, 0x0a, 0x0b, 0x43, 0x72, 0x61, 0x77, 0x6c, 0x65, 0x72, 0x4e,
	0x6f, 0x64, 0x65, 0x12, 0x50, 0x0a, 0x07, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x16,
	0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
	0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x1a, 0x2b, 0x2e, 0x63, 0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65,
	0x6f, 0x6e, 0x2e, 0x73, 0x6d, 0x65, 0x6c, 0x74, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x63, 0x72,
	0x61, 0x77, 0x6c, 0x2e, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x79, 0x0a, 0x0e, 0x43, 0x72, 0x61, 0x77, 0x6c, 0x65, 0x72,
	0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x12, 0x31, 0x2e, 0x63, 0x68, 0x61, 0x6d, 0x65, 0x6c,
	0x65, 0x6f, 0x6e, 0x2e, 0x73, 0x6d, 0x65, 0x6c, 0x74, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x63,
	0x72, 0x61, 0x77, 0x6c, 0x2e, 0x43, 0x72, 0x61, 0x77, 0x6c, 0x65, 0x72, 0x4f, 0x70, 0x74, 0x69,
	0x6f, 0x6e, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x32, 0x2e, 0x63, 0x68, 0x61,
	0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x2e, 0x73, 0x6d, 0x65, 0x6c, 0x74, 0x65, 0x72, 0x2e, 0x76,
	0x31, 0x2e, 0x63, 0x72, 0x61, 0x77, 0x6c, 0x2e, 0x43, 0x72, 0x61, 0x77, 0x6c, 0x65, 0x72, 0x4f,
	0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00,
	0x12, 0x5e, 0x0a, 0x0e, 0x41, 0x6c, 0x6c, 0x6f, 0x77, 0x65, 0x64, 0x44, 0x6f, 0x6d, 0x61, 0x69,
	0x6e, 0x73, 0x12, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x1a, 0x32, 0x2e, 0x63, 0x68, 0x61,
	0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x2e, 0x73, 0x6d, 0x65, 0x6c, 0x74, 0x65, 0x72, 0x2e, 0x76,
	0x31, 0x2e, 0x63, 0x72, 0x61, 0x77, 0x6c, 0x2e, 0x41, 0x6c, 0x6c, 0x6f, 0x77, 0x65, 0x64, 0x44,
	0x6f, 0x6d, 0x61, 0x69, 0x6e, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00,
	0x12, 0x73, 0x0a, 0x0c, 0x43, 0x61, 0x6e, 0x6f, 0x6e, 0x69, 0x63, 0x61, 0x6c, 0x55, 0x72, 0x6c,
	0x12, 0x2f, 0x2e, 0x63, 0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x2e, 0x73, 0x6d, 0x65,
	0x6c, 0x74, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x63, 0x72, 0x61, 0x77, 0x6c, 0x2e, 0x43, 0x61,
	0x6e, 0x6f, 0x6e, 0x69, 0x63, 0x61, 0x6c, 0x55, 0x72, 0x6c, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x1a, 0x30, 0x2e, 0x63, 0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x2e, 0x73, 0x6d,
	0x65, 0x6c, 0x74, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x63, 0x72, 0x61, 0x77, 0x6c, 0x2e, 0x43,
	0x61, 0x6e, 0x6f, 0x6e, 0x69, 0x63, 0x61, 0x6c, 0x55, 0x72, 0x6c, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x46, 0x0a, 0x05, 0x50, 0x61, 0x72, 0x73, 0x65, 0x12, 0x23,
	0x2e, 0x63, 0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x2e, 0x73, 0x6d, 0x65, 0x6c, 0x74,
	0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x63, 0x72, 0x61, 0x77, 0x6c, 0x2e, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x1a, 0x14, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x41, 0x6e, 0x79, 0x22, 0x00, 0x30, 0x01, 0x12, 0x5b, 0x0a,
	0x04, 0x43, 0x61, 0x6c, 0x6c, 0x12, 0x27, 0x2e, 0x63, 0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f,
	0x6e, 0x2e, 0x73, 0x6d, 0x65, 0x6c, 0x74, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x63, 0x72, 0x61,
	0x77, 0x6c, 0x2e, 0x43, 0x61, 0x6c, 0x6c, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x28,
	0x2e, 0x63, 0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x2e, 0x73, 0x6d, 0x65, 0x6c, 0x74,
	0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x63, 0x72, 0x61, 0x77, 0x6c, 0x2e, 0x43, 0x61, 0x6c, 0x6c,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x32, 0x4e, 0x0a, 0x0f, 0x43, 0x72,
	0x61, 0x77, 0x6c, 0x65, 0x72, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x65, 0x72, 0x12, 0x3b, 0x0a,
	0x07, 0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x12, 0x14, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x41, 0x6e, 0x79, 0x1a, 0x14,
	0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
	0x2e, 0x41, 0x6e, 0x79, 0x22, 0x00, 0x28, 0x01, 0x30, 0x01, 0x32, 0xd2, 0x07, 0x0a, 0x0e, 0x43,
	0x72, 0x61, 0x77, 0x6c, 0x65, 0x72, 0x4d, 0x61, 0x6e, 0x61, 0x67, 0x65, 0x72, 0x12, 0x92, 0x01,
	0x0a, 0x0b, 0x47, 0x65, 0x74, 0x43, 0x72, 0x61, 0x77, 0x6c, 0x65, 0x72, 0x73, 0x12, 0x2e, 0x2e,
	0x63, 0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x2e, 0x73, 0x6d, 0x65, 0x6c, 0x74, 0x65,
	0x72, 0x2e, 0x76, 0x31, 0x2e, 0x63, 0x72, 0x61, 0x77, 0x6c, 0x2e, 0x47, 0x65, 0x74, 0x43, 0x72,
	0x61, 0x77, 0x6c, 0x65, 0x72, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x2f, 0x2e,
	0x63, 0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x2e, 0x73, 0x6d, 0x65, 0x6c, 0x74, 0x65,
	0x72, 0x2e, 0x76, 0x31, 0x2e, 0x63, 0x72, 0x61, 0x77, 0x6c, 0x2e, 0x47, 0x65, 0x74, 0x43, 0x72,
	0x61, 0x77, 0x6c, 0x65, 0x72, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x22,
	0x82, 0xd3, 0xe4, 0x93, 0x02, 0x1c, 0x12, 0x1a, 0x2f, 0x73, 0x6d, 0x65, 0x6c, 0x74, 0x65, 0x72,
	0x2f, 0x76, 0x31, 0x2f, 0x63, 0x72, 0x61, 0x77, 0x6c, 0x2f, 0x63, 0x72, 0x61, 0x77, 0x6c, 0x65,
	0x72, 0x73, 0x12, 0x98, 0x01, 0x0a, 0x0a, 0x47, 0x65, 0x74, 0x43, 0x72, 0x61, 0x77, 0x6c, 0x65,
	0x72, 0x12, 0x2d, 0x2e, 0x63, 0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x2e, 0x73, 0x6d,
	0x65, 0x6c, 0x74, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x63, 0x72, 0x61, 0x77, 0x6c, 0x2e, 0x47,
	0x65, 0x74, 0x43, 0x72, 0x61, 0x77, 0x6c, 0x65, 0x72, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x1a, 0x2e, 0x2e, 0x63, 0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x2e, 0x73, 0x6d, 0x65,
	0x6c, 0x74, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x63, 0x72, 0x61, 0x77, 0x6c, 0x2e, 0x47, 0x65,
	0x74, 0x43, 0x72, 0x61, 0x77, 0x6c, 0x65, 0x72, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x22, 0x2b, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x25, 0x12, 0x23, 0x2f, 0x73, 0x6d, 0x65, 0x6c, 0x74,
	0x65, 0x72, 0x2f, 0x76, 0x31, 0x2f, 0x63, 0x72, 0x61, 0x77, 0x6c, 0x2f, 0x63, 0x72, 0x61, 0x77,
	0x6c, 0x65, 0x72, 0x73, 0x2f, 0x7b, 0x73, 0x69, 0x74, 0x65, 0x49, 0x64, 0x7d, 0x12, 0xbd, 0x01,
	0x0a, 0x11, 0x47, 0x65, 0x74, 0x43, 0x72, 0x61, 0x77, 0x6c, 0x65, 0x72, 0x4f, 0x70, 0x74, 0x69,
	0x6f, 0x6e, 0x73, 0x12, 0x34, 0x2e, 0x63, 0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x2e,
	0x73, 0x6d, 0x65, 0x6c, 0x74, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x63, 0x72, 0x61, 0x77, 0x6c,
	0x2e, 0x47, 0x65, 0x74, 0x43, 0x72, 0x61, 0x77, 0x6c, 0x65, 0x72, 0x4f, 0x70, 0x74, 0x69, 0x6f,
	0x6e, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x35, 0x2e, 0x63, 0x68, 0x61, 0x6d,
	0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x2e, 0x73, 0x6d, 0x65, 0x6c, 0x74, 0x65, 0x72, 0x2e, 0x76, 0x31,
	0x2e, 0x63, 0x72, 0x61, 0x77, 0x6c, 0x2e, 0x47, 0x65, 0x74, 0x43, 0x72, 0x61, 0x77, 0x6c, 0x65,
	0x72, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x22, 0x3b, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x35, 0x12, 0x33, 0x2f, 0x73, 0x6d, 0x65, 0x6c, 0x74,
	0x65, 0x72, 0x2f, 0x76, 0x31, 0x2f, 0x63, 0x72, 0x61, 0x77, 0x6c, 0x2f, 0x63, 0x72, 0x61, 0x77,
	0x6c, 0x65, 0x72, 0x73, 0x2f, 0x7b, 0x73, 0x69, 0x74, 0x65, 0x49, 0x64, 0x7d, 0x2f, 0x63, 0x72,
	0x61, 0x77, 0x6c, 0x65, 0x72, 0x5f, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x12, 0xb5, 0x01,
	0x0a, 0x0f, 0x47, 0x65, 0x74, 0x43, 0x61, 0x6e, 0x6f, 0x6e, 0x69, 0x63, 0x61, 0x6c, 0x55, 0x72,
	0x6c, 0x12, 0x32, 0x2e, 0x63, 0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x2e, 0x73, 0x6d,
	0x65, 0x6c, 0x74, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x63, 0x72, 0x61, 0x77, 0x6c, 0x2e, 0x47,
	0x65, 0x74, 0x43, 0x61, 0x6e, 0x6f, 0x6e, 0x69, 0x63, 0x61, 0x6c, 0x55, 0x72, 0x6c, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x33, 0x2e, 0x63, 0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f,
	0x6e, 0x2e, 0x73, 0x6d, 0x65, 0x6c, 0x74, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x63, 0x72, 0x61,
	0x77, 0x6c, 0x2e, 0x47, 0x65, 0x74, 0x43, 0x61, 0x6e, 0x6f, 0x6e, 0x69, 0x63, 0x61, 0x6c, 0x55,
	0x72, 0x6c, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x39, 0x82, 0xd3, 0xe4, 0x93,
	0x02, 0x33, 0x12, 0x31, 0x2f, 0x73, 0x6d, 0x65, 0x6c, 0x74, 0x65, 0x72, 0x2f, 0x76, 0x31, 0x2f,
	0x63, 0x72, 0x61, 0x77, 0x6c, 0x2f, 0x63, 0x72, 0x61, 0x77, 0x6c, 0x65, 0x72, 0x73, 0x2f, 0x7b,
	0x73, 0x69, 0x74, 0x65, 0x49, 0x64, 0x7d, 0x2f, 0x63, 0x61, 0x6e, 0x6f, 0x6e, 0x69, 0x63, 0x61,
	0x6c, 0x5f, 0x75, 0x72, 0x6c, 0x12, 0x86, 0x01, 0x0a, 0x07, 0x44, 0x6f, 0x50, 0x61, 0x72, 0x73,
	0x65, 0x12, 0x2a, 0x2e, 0x63, 0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x2e, 0x73, 0x6d,
	0x65, 0x6c, 0x74, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x63, 0x72, 0x61, 0x77, 0x6c, 0x2e, 0x44,
	0x6f, 0x50, 0x61, 0x72, 0x73, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x2b, 0x2e,
	0x63, 0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x2e, 0x73, 0x6d, 0x65, 0x6c, 0x74, 0x65,
	0x72, 0x2e, 0x76, 0x31, 0x2e, 0x63, 0x72, 0x61, 0x77, 0x6c, 0x2e, 0x44, 0x6f, 0x50, 0x61, 0x72,
	0x73, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x22, 0x82, 0xd3, 0xe4, 0x93,
	0x02, 0x1c, 0x22, 0x17, 0x2f, 0x73, 0x6d, 0x65, 0x6c, 0x74, 0x65, 0x72, 0x2f, 0x76, 0x31, 0x2f,
	0x63, 0x72, 0x61, 0x77, 0x6c, 0x2f, 0x70, 0x61, 0x72, 0x73, 0x65, 0x3a, 0x01, 0x2a, 0x12, 0x8e,
	0x01, 0x0a, 0x0a, 0x52, 0x65, 0x6d, 0x6f, 0x74, 0x65, 0x43, 0x61, 0x6c, 0x6c, 0x12, 0x2d, 0x2e,
	0x63, 0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x2e, 0x73, 0x6d, 0x65, 0x6c, 0x74, 0x65,
	0x72, 0x2e, 0x76, 0x31, 0x2e, 0x63, 0x72, 0x61, 0x77, 0x6c, 0x2e, 0x52, 0x65, 0x6d, 0x6f, 0x74,
	0x65, 0x43, 0x61, 0x6c, 0x6c, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x2e, 0x2e, 0x63,
	0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x2e, 0x73, 0x6d, 0x65, 0x6c, 0x74, 0x65, 0x72,
	0x2e, 0x76, 0x31, 0x2e, 0x63, 0x72, 0x61, 0x77, 0x6c, 0x2e, 0x52, 0x65, 0x6d, 0x6f, 0x74, 0x65,
	0x43, 0x61, 0x6c, 0x6c, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x21, 0x82, 0xd3,
	0xe4, 0x93, 0x02, 0x1b, 0x22, 0x16, 0x2f, 0x73, 0x6d, 0x65, 0x6c, 0x74, 0x65, 0x72, 0x2f, 0x76,
	0x31, 0x2f, 0x63, 0x72, 0x61, 0x77, 0x6c, 0x2f, 0x63, 0x61, 0x6c, 0x6c, 0x3a, 0x01, 0x2a, 0x32,
	0xa5, 0x07, 0x0a, 0x07, 0x47, 0x61, 0x74, 0x65, 0x77, 0x61, 0x79, 0x12, 0x92, 0x01, 0x0a, 0x0b,
	0x47, 0x65, 0x74, 0x43, 0x72, 0x61, 0x77, 0x6c, 0x65, 0x72, 0x73, 0x12, 0x2e, 0x2e, 0x63, 0x68,
	0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x2e, 0x73, 0x6d, 0x65, 0x6c, 0x74, 0x65, 0x72, 0x2e,
	0x76, 0x31, 0x2e, 0x63, 0x72, 0x61, 0x77, 0x6c, 0x2e, 0x47, 0x65, 0x74, 0x43, 0x72, 0x61, 0x77,
	0x6c, 0x65, 0x72, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x2f, 0x2e, 0x63, 0x68,
	0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x2e, 0x73, 0x6d, 0x65, 0x6c, 0x74, 0x65, 0x72, 0x2e,
	0x76, 0x31, 0x2e, 0x63, 0x72, 0x61, 0x77, 0x6c, 0x2e, 0x47, 0x65, 0x74, 0x43, 0x72, 0x61, 0x77,
	0x6c, 0x65, 0x72, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x22, 0x82, 0xd3,
	0xe4, 0x93, 0x02, 0x1c, 0x12, 0x1a, 0x2f, 0x73, 0x6d, 0x65, 0x6c, 0x74, 0x65, 0x72, 0x2f, 0x76,
	0x31, 0x2f, 0x63, 0x72, 0x61, 0x77, 0x6c, 0x2f, 0x63, 0x72, 0x61, 0x77, 0x6c, 0x65, 0x72, 0x73,
	0x12, 0x98, 0x01, 0x0a, 0x0a, 0x47, 0x65, 0x74, 0x43, 0x72, 0x61, 0x77, 0x6c, 0x65, 0x72, 0x12,
	0x2d, 0x2e, 0x63, 0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x2e, 0x73, 0x6d, 0x65, 0x6c,
	0x74, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x63, 0x72, 0x61, 0x77, 0x6c, 0x2e, 0x47, 0x65, 0x74,
	0x43, 0x72, 0x61, 0x77, 0x6c, 0x65, 0x72, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x2e,
	0x2e, 0x63, 0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x2e, 0x73, 0x6d, 0x65, 0x6c, 0x74,
	0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x63, 0x72, 0x61, 0x77, 0x6c, 0x2e, 0x47, 0x65, 0x74, 0x43,
	0x72, 0x61, 0x77, 0x6c, 0x65, 0x72, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x2b,
	0x82, 0xd3, 0xe4, 0x93, 0x02, 0x25, 0x12, 0x23, 0x2f, 0x73, 0x6d, 0x65, 0x6c, 0x74, 0x65, 0x72,
	0x2f, 0x76, 0x31, 0x2f, 0x63, 0x72, 0x61, 0x77, 0x6c, 0x2f, 0x63, 0x72, 0x61, 0x77, 0x6c, 0x65,
	0x72, 0x73, 0x2f, 0x7b, 0x73, 0x69, 0x74, 0x65, 0x49, 0x64, 0x7d, 0x12, 0xb5, 0x01, 0x0a, 0x0f,
	0x47, 0x65, 0x74, 0x43, 0x61, 0x6e, 0x6f, 0x6e, 0x69, 0x63, 0x61, 0x6c, 0x55, 0x72, 0x6c, 0x12,
	0x32, 0x2e, 0x63, 0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x2e, 0x73, 0x6d, 0x65, 0x6c,
	0x74, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x63, 0x72, 0x61, 0x77, 0x6c, 0x2e, 0x47, 0x65, 0x74,
	0x43, 0x61, 0x6e, 0x6f, 0x6e, 0x69, 0x63, 0x61, 0x6c, 0x55, 0x72, 0x6c, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x1a, 0x33, 0x2e, 0x63, 0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x2e,
	0x73, 0x6d, 0x65, 0x6c, 0x74, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x63, 0x72, 0x61, 0x77, 0x6c,
	0x2e, 0x47, 0x65, 0x74, 0x43, 0x61, 0x6e, 0x6f, 0x6e, 0x69, 0x63, 0x61, 0x6c, 0x55, 0x72, 0x6c,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x39, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x33,
	0x12, 0x31, 0x2f, 0x73, 0x6d, 0x65, 0x6c, 0x74, 0x65, 0x72, 0x2f, 0x76, 0x31, 0x2f, 0x63, 0x72,
	0x61, 0x77, 0x6c, 0x2f, 0x63, 0x72, 0x61, 0x77, 0x6c, 0x65, 0x72, 0x73, 0x2f, 0x7b, 0x73, 0x69,
	0x74, 0x65, 0x49, 0x64, 0x7d, 0x2f, 0x63, 0x61, 0x6e, 0x6f, 0x6e, 0x69, 0x63, 0x61, 0x6c, 0x5f,
	0x75, 0x72, 0x6c, 0x12, 0x80, 0x01, 0x0a, 0x05, 0x46, 0x65, 0x74, 0x63, 0x68, 0x12, 0x28, 0x2e,
	0x63, 0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x2e, 0x73, 0x6d, 0x65, 0x6c, 0x74, 0x65,
	0x72, 0x2e, 0x76, 0x31, 0x2e, 0x63, 0x72, 0x61, 0x77, 0x6c, 0x2e, 0x46, 0x65, 0x74, 0x63, 0x68,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x29, 0x2e, 0x63, 0x68, 0x61, 0x6d, 0x65, 0x6c,
	0x65, 0x6f, 0x6e, 0x2e, 0x73, 0x6d, 0x65, 0x6c, 0x74, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x63,
	0x72, 0x61, 0x77, 0x6c, 0x2e, 0x46, 0x65, 0x74, 0x63, 0x68, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x22, 0x22, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x1c, 0x22, 0x17, 0x2f, 0x73, 0x6d, 0x65,
	0x6c, 0x74, 0x65, 0x72, 0x2f, 0x76, 0x31, 0x2f, 0x63, 0x72, 0x61, 0x77, 0x6c, 0x2f, 0x66, 0x65,
	0x74, 0x63, 0x68, 0x3a, 0x01, 0x2a, 0x12, 0x94, 0x01, 0x0a, 0x0a, 0x47, 0x65, 0x74, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x2d, 0x2e, 0x63, 0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f,
	0x6e, 0x2e, 0x73, 0x6d, 0x65, 0x6c, 0x74, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x63, 0x72, 0x61,
	0x77, 0x6c, 0x2e, 0x47, 0x65, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x1a, 0x2e, 0x2e, 0x63, 0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e,
	0x2e, 0x73, 0x6d, 0x65, 0x6c, 0x74, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x63, 0x72, 0x61, 0x77,
	0x6c, 0x2e, 0x47, 0x65, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x22, 0x27, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x21, 0x12, 0x1f, 0x2f, 0x73,
	0x6d, 0x65, 0x6c, 0x74, 0x65, 0x72, 0x2f, 0x76, 0x31, 0x2f, 0x63, 0x72, 0x61, 0x77, 0x6c, 0x2f,
	0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x73, 0x2f, 0x7b, 0x69, 0x64, 0x7d, 0x12, 0x97, 0x01,
	0x0a, 0x0e, 0x47, 0x65, 0x74, 0x43, 0x72, 0x61, 0x77, 0x6c, 0x65, 0x72, 0x4c, 0x6f, 0x67, 0x73,
	0x12, 0x31, 0x2e, 0x63, 0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x2e, 0x73, 0x6d, 0x65,
	0x6c, 0x74, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x63, 0x72, 0x61, 0x77, 0x6c, 0x2e, 0x47, 0x65,
	0x74, 0x43, 0x72, 0x61, 0x77, 0x6c, 0x65, 0x72, 0x4c, 0x6f, 0x67, 0x73, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x1a, 0x32, 0x2e, 0x63, 0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x2e,
	0x73, 0x6d, 0x65, 0x6c, 0x74, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x63, 0x72, 0x61, 0x77, 0x6c,
	0x2e, 0x47, 0x65, 0x74, 0x43, 0x72, 0x61, 0x77, 0x6c, 0x65, 0x72, 0x4c, 0x6f, 0x67, 0x73, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x1e, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x18, 0x12,
	0x16, 0x2f, 0x73, 0x6d, 0x65, 0x6c, 0x74, 0x65, 0x72, 0x2f, 0x76, 0x31, 0x2f, 0x63, 0x72, 0x61,
	0x77, 0x6c, 0x2f, 0x6c, 0x6f, 0x67, 0x73, 0x42, 0x81, 0x02, 0x5a, 0x20, 0x63, 0x68, 0x61, 0x6d,
	0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x2f, 0x73, 0x6d, 0x65, 0x6c, 0x74, 0x65, 0x72, 0x2f, 0x76, 0x31,
	0x2f, 0x63, 0x72, 0x61, 0x77, 0x6c, 0x3b, 0x63, 0x72, 0x61, 0x77, 0x6c, 0x92, 0x41, 0xdb, 0x01,
	0x12, 0x63, 0x0a, 0x10, 0x73, 0x6d, 0x65, 0x6c, 0x74, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x63,
	0x72, 0x61, 0x77, 0x6c, 0x22, 0x42, 0x0a, 0x04, 0x53, 0x65, 0x65, 0x72, 0x12, 0x28, 0x68, 0x74,
	0x74, 0x70, 0x73, 0x3a, 0x2f, 0x2f, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d,
	0x2f, 0x63, 0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x64, 0x65, 0x76, 0x2f, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x1a, 0x10, 0x6b, 0x76, 0x63, 0x6e, 0x6f, 0x77, 0x40, 0x67,
	0x6d, 0x61, 0x69, 0x6c, 0x2e, 0x63, 0x6f, 0x6d, 0x32, 0x0b, 0x5f, 0x5f, 0x56, 0x45, 0x52, 0x53,
	0x49, 0x4f, 0x4e, 0x5f, 0x5f, 0x2a, 0x02, 0x01, 0x02, 0x32, 0x10, 0x61, 0x70, 0x70, 0x6c, 0x69,
	0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2f, 0x6a, 0x73, 0x6f, 0x6e, 0x3a, 0x10, 0x61, 0x70, 0x70,
	0x6c, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2f, 0x6a, 0x73, 0x6f, 0x6e, 0x52, 0x12, 0x0a,
	0x03, 0x34, 0x30, 0x33, 0x12, 0x0b, 0x0a, 0x09, 0xe6, 0x9c, 0xaa, 0xe6, 0x8e, 0x88, 0xe6, 0x9d,
	0x83, 0x52, 0x21, 0x0a, 0x03, 0x34, 0x30, 0x34, 0x12, 0x1a, 0x0a, 0x18, 0xe6, 0x9c, 0xaa, 0xe6,
	0x89, 0xbe, 0xe5, 0x88, 0xb0, 0xe6, 0x88, 0x96, 0xe8, 0x80, 0x85, 0xe4, 0xb8, 0x8d, 0xe5, 0xad,
	0x98, 0xe5, 0x9c, 0xa8, 0x52, 0x15, 0x0a, 0x03, 0x35, 0x30, 0x30, 0x12, 0x0e, 0x0a, 0x0c, 0xe7,
	0xb3, 0xbb, 0xe7, 0xbb, 0x9f, 0xe9, 0x94, 0x99, 0xe8, 0xaf, 0xaf, 0x62, 0x06, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x33,
}

var file_chameleon_smelter_v1_crawl_service_proto_goTypes = []interface{}{
	(*emptypb.Empty)(nil),             // 0: google.protobuf.Empty
	(*CrawlerOptionsRequest)(nil),     // 1: chameleon.smelter.v1.crawl.CrawlerOptionsRequest
	(*CanonicalUrlRequest)(nil),       // 2: chameleon.smelter.v1.crawl.CanonicalUrlRequest
	(*Request)(nil),                   // 3: chameleon.smelter.v1.crawl.Request
	(*CallRequest)(nil),               // 4: chameleon.smelter.v1.crawl.CallRequest
	(*anypb.Any)(nil),                 // 5: google.protobuf.Any
	(*GetCrawlersRequest)(nil),        // 6: chameleon.smelter.v1.crawl.GetCrawlersRequest
	(*GetCrawlerRequest)(nil),         // 7: chameleon.smelter.v1.crawl.GetCrawlerRequest
	(*GetCrawlerOptionsRequest)(nil),  // 8: chameleon.smelter.v1.crawl.GetCrawlerOptionsRequest
	(*GetCanonicalUrlRequest)(nil),    // 9: chameleon.smelter.v1.crawl.GetCanonicalUrlRequest
	(*DoParseRequest)(nil),            // 10: chameleon.smelter.v1.crawl.DoParseRequest
	(*RemoteCallRequest)(nil),         // 11: chameleon.smelter.v1.crawl.RemoteCallRequest
	(*FetchRequest)(nil),              // 12: chameleon.smelter.v1.crawl.FetchRequest
	(*GetRequestRequest)(nil),         // 13: chameleon.smelter.v1.crawl.GetRequestRequest
	(*GetCrawlerLogsRequest)(nil),     // 14: chameleon.smelter.v1.crawl.GetCrawlerLogsRequest
	(*VersionResponse)(nil),           // 15: chameleon.smelter.v1.crawl.VersionResponse
	(*CrawlerOptionsResponse)(nil),    // 16: chameleon.smelter.v1.crawl.CrawlerOptionsResponse
	(*AllowedDomainsResponse)(nil),    // 17: chameleon.smelter.v1.crawl.AllowedDomainsResponse
	(*CanonicalUrlResponse)(nil),      // 18: chameleon.smelter.v1.crawl.CanonicalUrlResponse
	(*CallResponse)(nil),              // 19: chameleon.smelter.v1.crawl.CallResponse
	(*GetCrawlersResponse)(nil),       // 20: chameleon.smelter.v1.crawl.GetCrawlersResponse
	(*GetCrawlerResponse)(nil),        // 21: chameleon.smelter.v1.crawl.GetCrawlerResponse
	(*GetCrawlerOptionsResponse)(nil), // 22: chameleon.smelter.v1.crawl.GetCrawlerOptionsResponse
	(*GetCanonicalUrlResponse)(nil),   // 23: chameleon.smelter.v1.crawl.GetCanonicalUrlResponse
	(*DoParseResponse)(nil),           // 24: chameleon.smelter.v1.crawl.DoParseResponse
	(*RemoteCallResponse)(nil),        // 25: chameleon.smelter.v1.crawl.RemoteCallResponse
	(*FetchResponse)(nil),             // 26: chameleon.smelter.v1.crawl.FetchResponse
	(*GetRequestResponse)(nil),        // 27: chameleon.smelter.v1.crawl.GetRequestResponse
	(*GetCrawlerLogsResponse)(nil),    // 28: chameleon.smelter.v1.crawl.GetCrawlerLogsResponse
}
var file_chameleon_smelter_v1_crawl_service_proto_depIdxs = []int32{
	0,  // 0: chameleon.smelter.v1.crawl.CrawlerNode.Version:input_type -> google.protobuf.Empty
	1,  // 1: chameleon.smelter.v1.crawl.CrawlerNode.CrawlerOptions:input_type -> chameleon.smelter.v1.crawl.CrawlerOptionsRequest
	0,  // 2: chameleon.smelter.v1.crawl.CrawlerNode.AllowedDomains:input_type -> google.protobuf.Empty
	2,  // 3: chameleon.smelter.v1.crawl.CrawlerNode.CanonicalUrl:input_type -> chameleon.smelter.v1.crawl.CanonicalUrlRequest
	3,  // 4: chameleon.smelter.v1.crawl.CrawlerNode.Parse:input_type -> chameleon.smelter.v1.crawl.Request
	4,  // 5: chameleon.smelter.v1.crawl.CrawlerNode.Call:input_type -> chameleon.smelter.v1.crawl.CallRequest
	5,  // 6: chameleon.smelter.v1.crawl.CrawlerRegister.Connect:input_type -> google.protobuf.Any
	6,  // 7: chameleon.smelter.v1.crawl.CrawlerManager.GetCrawlers:input_type -> chameleon.smelter.v1.crawl.GetCrawlersRequest
	7,  // 8: chameleon.smelter.v1.crawl.CrawlerManager.GetCrawler:input_type -> chameleon.smelter.v1.crawl.GetCrawlerRequest
	8,  // 9: chameleon.smelter.v1.crawl.CrawlerManager.GetCrawlerOptions:input_type -> chameleon.smelter.v1.crawl.GetCrawlerOptionsRequest
	9,  // 10: chameleon.smelter.v1.crawl.CrawlerManager.GetCanonicalUrl:input_type -> chameleon.smelter.v1.crawl.GetCanonicalUrlRequest
	10, // 11: chameleon.smelter.v1.crawl.CrawlerManager.DoParse:input_type -> chameleon.smelter.v1.crawl.DoParseRequest
	11, // 12: chameleon.smelter.v1.crawl.CrawlerManager.RemoteCall:input_type -> chameleon.smelter.v1.crawl.RemoteCallRequest
	6,  // 13: chameleon.smelter.v1.crawl.Gateway.GetCrawlers:input_type -> chameleon.smelter.v1.crawl.GetCrawlersRequest
	7,  // 14: chameleon.smelter.v1.crawl.Gateway.GetCrawler:input_type -> chameleon.smelter.v1.crawl.GetCrawlerRequest
	9,  // 15: chameleon.smelter.v1.crawl.Gateway.GetCanonicalUrl:input_type -> chameleon.smelter.v1.crawl.GetCanonicalUrlRequest
	12, // 16: chameleon.smelter.v1.crawl.Gateway.Fetch:input_type -> chameleon.smelter.v1.crawl.FetchRequest
	13, // 17: chameleon.smelter.v1.crawl.Gateway.GetRequest:input_type -> chameleon.smelter.v1.crawl.GetRequestRequest
	14, // 18: chameleon.smelter.v1.crawl.Gateway.GetCrawlerLogs:input_type -> chameleon.smelter.v1.crawl.GetCrawlerLogsRequest
	15, // 19: chameleon.smelter.v1.crawl.CrawlerNode.Version:output_type -> chameleon.smelter.v1.crawl.VersionResponse
	16, // 20: chameleon.smelter.v1.crawl.CrawlerNode.CrawlerOptions:output_type -> chameleon.smelter.v1.crawl.CrawlerOptionsResponse
	17, // 21: chameleon.smelter.v1.crawl.CrawlerNode.AllowedDomains:output_type -> chameleon.smelter.v1.crawl.AllowedDomainsResponse
	18, // 22: chameleon.smelter.v1.crawl.CrawlerNode.CanonicalUrl:output_type -> chameleon.smelter.v1.crawl.CanonicalUrlResponse
	5,  // 23: chameleon.smelter.v1.crawl.CrawlerNode.Parse:output_type -> google.protobuf.Any
	19, // 24: chameleon.smelter.v1.crawl.CrawlerNode.Call:output_type -> chameleon.smelter.v1.crawl.CallResponse
	5,  // 25: chameleon.smelter.v1.crawl.CrawlerRegister.Connect:output_type -> google.protobuf.Any
	20, // 26: chameleon.smelter.v1.crawl.CrawlerManager.GetCrawlers:output_type -> chameleon.smelter.v1.crawl.GetCrawlersResponse
	21, // 27: chameleon.smelter.v1.crawl.CrawlerManager.GetCrawler:output_type -> chameleon.smelter.v1.crawl.GetCrawlerResponse
	22, // 28: chameleon.smelter.v1.crawl.CrawlerManager.GetCrawlerOptions:output_type -> chameleon.smelter.v1.crawl.GetCrawlerOptionsResponse
	23, // 29: chameleon.smelter.v1.crawl.CrawlerManager.GetCanonicalUrl:output_type -> chameleon.smelter.v1.crawl.GetCanonicalUrlResponse
	24, // 30: chameleon.smelter.v1.crawl.CrawlerManager.DoParse:output_type -> chameleon.smelter.v1.crawl.DoParseResponse
	25, // 31: chameleon.smelter.v1.crawl.CrawlerManager.RemoteCall:output_type -> chameleon.smelter.v1.crawl.RemoteCallResponse
	20, // 32: chameleon.smelter.v1.crawl.Gateway.GetCrawlers:output_type -> chameleon.smelter.v1.crawl.GetCrawlersResponse
	21, // 33: chameleon.smelter.v1.crawl.Gateway.GetCrawler:output_type -> chameleon.smelter.v1.crawl.GetCrawlerResponse
	23, // 34: chameleon.smelter.v1.crawl.Gateway.GetCanonicalUrl:output_type -> chameleon.smelter.v1.crawl.GetCanonicalUrlResponse
	26, // 35: chameleon.smelter.v1.crawl.Gateway.Fetch:output_type -> chameleon.smelter.v1.crawl.FetchResponse
	27, // 36: chameleon.smelter.v1.crawl.Gateway.GetRequest:output_type -> chameleon.smelter.v1.crawl.GetRequestResponse
	28, // 37: chameleon.smelter.v1.crawl.Gateway.GetCrawlerLogs:output_type -> chameleon.smelter.v1.crawl.GetCrawlerLogsResponse
	19, // [19:38] is the sub-list for method output_type
	0,  // [0:19] is the sub-list for method input_type
	0,  // [0:0] is the sub-list for extension type_name
	0,  // [0:0] is the sub-list for extension extendee
	0,  // [0:0] is the sub-list for field type_name
}

func init() { file_chameleon_smelter_v1_crawl_service_proto_init() }
func file_chameleon_smelter_v1_crawl_service_proto_init() {
	if File_chameleon_smelter_v1_crawl_service_proto != nil {
		return
	}
	file_chameleon_smelter_v1_crawl_data_proto_init()
	file_chameleon_smelter_v1_crawl_service_message_proto_init()
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_chameleon_smelter_v1_crawl_service_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   0,
			NumExtensions: 0,
			NumServices:   4,
		},
		GoTypes:           file_chameleon_smelter_v1_crawl_service_proto_goTypes,
		DependencyIndexes: file_chameleon_smelter_v1_crawl_service_proto_depIdxs,
	}.Build()
	File_chameleon_smelter_v1_crawl_service_proto = out.File
	file_chameleon_smelter_v1_crawl_service_proto_rawDesc = nil
	file_chameleon_smelter_v1_crawl_service_proto_goTypes = nil
	file_chameleon_smelter_v1_crawl_service_proto_depIdxs = nil
}
