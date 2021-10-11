// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        v3.18.0
// source: chameleon/api/media/data.proto

package media

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	anypb "google.golang.org/protobuf/types/known/anypb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// Media
type Media struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Detail
	Detail *anypb.Any `protobuf:"bytes,1,opt,name=detail,proto3" json:"detail,omitempty"`
	// Text
	Text string `protobuf:"bytes,2,opt,name=text,proto3" json:"text,omitempty"`
	// IsDefault
	IsDefault bool `protobuf:"varint,3,opt,name=isDefault,proto3" json:"isDefault,omitempty"`
}

func (x *Media) Reset() {
	*x = Media{}
	if protoimpl.UnsafeEnabled {
		mi := &file_chameleon_api_media_data_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Media) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Media) ProtoMessage() {}

func (x *Media) ProtoReflect() protoreflect.Message {
	mi := &file_chameleon_api_media_data_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Media.ProtoReflect.Descriptor instead.
func (*Media) Descriptor() ([]byte, []int) {
	return file_chameleon_api_media_data_proto_rawDescGZIP(), []int{0}
}

func (x *Media) GetDetail() *anypb.Any {
	if x != nil {
		return x.Detail
	}
	return nil
}

func (x *Media) GetText() string {
	if x != nil {
		return x.Text
	}
	return ""
}

func (x *Media) GetIsDefault() bool {
	if x != nil {
		return x.IsDefault
	}
	return false
}

// Image
type Media_Image struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// ID
	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	// OriginalUrl the source url of the image
	OriginalUrl string `protobuf:"bytes,3,opt,name=originalUrl,proto3" json:"originalUrl,omitempty"`
	// largeUrl  width>=800px
	LargeUrl string `protobuf:"bytes,4,opt,name=largeUrl,proto3" json:"largeUrl,omitempty"`
	// MediumUrl width>=600px
	MediumUrl string `protobuf:"bytes,5,opt,name=mediumUrl,proto3" json:"mediumUrl,omitempty"`
	// SmallUrl  width>=500px
	SmallUrl string `protobuf:"bytes,6,opt,name=smallUrl,proto3" json:"smallUrl,omitempty"`
}

func (x *Media_Image) Reset() {
	*x = Media_Image{}
	if protoimpl.UnsafeEnabled {
		mi := &file_chameleon_api_media_data_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Media_Image) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Media_Image) ProtoMessage() {}

func (x *Media_Image) ProtoReflect() protoreflect.Message {
	mi := &file_chameleon_api_media_data_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Media_Image.ProtoReflect.Descriptor instead.
func (*Media_Image) Descriptor() ([]byte, []int) {
	return file_chameleon_api_media_data_proto_rawDescGZIP(), []int{0, 0}
}

func (x *Media_Image) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *Media_Image) GetOriginalUrl() string {
	if x != nil {
		return x.OriginalUrl
	}
	return ""
}

func (x *Media_Image) GetLargeUrl() string {
	if x != nil {
		return x.LargeUrl
	}
	return ""
}

func (x *Media_Image) GetMediumUrl() string {
	if x != nil {
		return x.MediumUrl
	}
	return ""
}

func (x *Media_Image) GetSmallUrl() string {
	if x != nil {
		return x.SmallUrl
	}
	return ""
}

// Video
type Media_Video struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Id
	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	// Type
	Type string `protobuf:"bytes,2,opt,name=type,proto3" json:"type,omitempty"`
	// OriginalUrl
	OriginalUrl string `protobuf:"bytes,3,opt,name=originalUrl,proto3" json:"originalUrl,omitempty"`
	// Width
	Width int32 `protobuf:"varint,6,opt,name=width,proto3" json:"width,omitempty"`
	// Height
	Height int32 `protobuf:"varint,7,opt,name=height,proto3" json:"height,omitempty"`
	// Duration
	Duration int32 `protobuf:"varint,8,opt,name=duration,proto3" json:"duration,omitempty"`
	// Cover video cover image
	Cover *Media_Image `protobuf:"bytes,11,opt,name=cover,proto3" json:"cover,omitempty"`
}

func (x *Media_Video) Reset() {
	*x = Media_Video{}
	if protoimpl.UnsafeEnabled {
		mi := &file_chameleon_api_media_data_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Media_Video) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Media_Video) ProtoMessage() {}

func (x *Media_Video) ProtoReflect() protoreflect.Message {
	mi := &file_chameleon_api_media_data_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Media_Video.ProtoReflect.Descriptor instead.
func (*Media_Video) Descriptor() ([]byte, []int) {
	return file_chameleon_api_media_data_proto_rawDescGZIP(), []int{0, 1}
}

func (x *Media_Video) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *Media_Video) GetType() string {
	if x != nil {
		return x.Type
	}
	return ""
}

func (x *Media_Video) GetOriginalUrl() string {
	if x != nil {
		return x.OriginalUrl
	}
	return ""
}

func (x *Media_Video) GetWidth() int32 {
	if x != nil {
		return x.Width
	}
	return 0
}

func (x *Media_Video) GetHeight() int32 {
	if x != nil {
		return x.Height
	}
	return 0
}

func (x *Media_Video) GetDuration() int32 {
	if x != nil {
		return x.Duration
	}
	return 0
}

func (x *Media_Video) GetCover() *Media_Image {
	if x != nil {
		return x.Cover
	}
	return nil
}

var File_chameleon_api_media_data_proto protoreflect.FileDescriptor

var file_chameleon_api_media_data_proto_rawDesc = []byte{
	0x0a, 0x1e, 0x63, 0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x2f, 0x61, 0x70, 0x69, 0x2f,
	0x6d, 0x65, 0x64, 0x69, 0x61, 0x2f, 0x64, 0x61, 0x74, 0x61, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x12, 0x13, 0x63, 0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x2e, 0x61, 0x70, 0x69, 0x2e,
	0x6d, 0x65, 0x64, 0x69, 0x61, 0x1a, 0x19, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x61, 0x6e, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x22, 0xcb, 0x03, 0x0a, 0x05, 0x4d, 0x65, 0x64, 0x69, 0x61, 0x12, 0x2c, 0x0a, 0x06, 0x64, 0x65,
	0x74, 0x61, 0x69, 0x6c, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x67, 0x6f, 0x6f,
	0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x41, 0x6e, 0x79,
	0x52, 0x06, 0x64, 0x65, 0x74, 0x61, 0x69, 0x6c, 0x12, 0x12, 0x0a, 0x04, 0x74, 0x65, 0x78, 0x74,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x74, 0x65, 0x78, 0x74, 0x12, 0x1c, 0x0a, 0x09,
	0x69, 0x73, 0x44, 0x65, 0x66, 0x61, 0x75, 0x6c, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x08, 0x52,
	0x09, 0x69, 0x73, 0x44, 0x65, 0x66, 0x61, 0x75, 0x6c, 0x74, 0x1a, 0x8f, 0x01, 0x0a, 0x05, 0x49,
	0x6d, 0x61, 0x67, 0x65, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x02, 0x69, 0x64, 0x12, 0x20, 0x0a, 0x0b, 0x6f, 0x72, 0x69, 0x67, 0x69, 0x6e, 0x61, 0x6c,
	0x55, 0x72, 0x6c, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x6f, 0x72, 0x69, 0x67, 0x69,
	0x6e, 0x61, 0x6c, 0x55, 0x72, 0x6c, 0x12, 0x1a, 0x0a, 0x08, 0x6c, 0x61, 0x72, 0x67, 0x65, 0x55,
	0x72, 0x6c, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x6c, 0x61, 0x72, 0x67, 0x65, 0x55,
	0x72, 0x6c, 0x12, 0x1c, 0x0a, 0x09, 0x6d, 0x65, 0x64, 0x69, 0x75, 0x6d, 0x55, 0x72, 0x6c, 0x18,
	0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x6d, 0x65, 0x64, 0x69, 0x75, 0x6d, 0x55, 0x72, 0x6c,
	0x12, 0x1a, 0x0a, 0x08, 0x73, 0x6d, 0x61, 0x6c, 0x6c, 0x55, 0x72, 0x6c, 0x18, 0x06, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x08, 0x73, 0x6d, 0x61, 0x6c, 0x6c, 0x55, 0x72, 0x6c, 0x1a, 0xcf, 0x01, 0x0a,
	0x05, 0x56, 0x69, 0x64, 0x65, 0x6f, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x74, 0x79, 0x70, 0x65, 0x12, 0x20, 0x0a, 0x0b, 0x6f, 0x72,
	0x69, 0x67, 0x69, 0x6e, 0x61, 0x6c, 0x55, 0x72, 0x6c, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x0b, 0x6f, 0x72, 0x69, 0x67, 0x69, 0x6e, 0x61, 0x6c, 0x55, 0x72, 0x6c, 0x12, 0x14, 0x0a, 0x05,
	0x77, 0x69, 0x64, 0x74, 0x68, 0x18, 0x06, 0x20, 0x01, 0x28, 0x05, 0x52, 0x05, 0x77, 0x69, 0x64,
	0x74, 0x68, 0x12, 0x16, 0x0a, 0x06, 0x68, 0x65, 0x69, 0x67, 0x68, 0x74, 0x18, 0x07, 0x20, 0x01,
	0x28, 0x05, 0x52, 0x06, 0x68, 0x65, 0x69, 0x67, 0x68, 0x74, 0x12, 0x1a, 0x0a, 0x08, 0x64, 0x75,
	0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x08, 0x20, 0x01, 0x28, 0x05, 0x52, 0x08, 0x64, 0x75,
	0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x36, 0x0a, 0x05, 0x63, 0x6f, 0x76, 0x65, 0x72, 0x18,
	0x0b, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x20, 0x2e, 0x63, 0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f,
	0x6e, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x6d, 0x65, 0x64, 0x69, 0x61, 0x2e, 0x4d, 0x65, 0x64, 0x69,
	0x61, 0x2e, 0x49, 0x6d, 0x61, 0x67, 0x65, 0x52, 0x05, 0x63, 0x6f, 0x76, 0x65, 0x72, 0x42, 0x1b,
	0x5a, 0x19, 0x63, 0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x2f, 0x61, 0x70, 0x69, 0x2f,
	0x6d, 0x65, 0x64, 0x69, 0x61, 0x3b, 0x6d, 0x65, 0x64, 0x69, 0x61, 0x62, 0x06, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x33,
}

var (
	file_chameleon_api_media_data_proto_rawDescOnce sync.Once
	file_chameleon_api_media_data_proto_rawDescData = file_chameleon_api_media_data_proto_rawDesc
)

func file_chameleon_api_media_data_proto_rawDescGZIP() []byte {
	file_chameleon_api_media_data_proto_rawDescOnce.Do(func() {
		file_chameleon_api_media_data_proto_rawDescData = protoimpl.X.CompressGZIP(file_chameleon_api_media_data_proto_rawDescData)
	})
	return file_chameleon_api_media_data_proto_rawDescData
}

var file_chameleon_api_media_data_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_chameleon_api_media_data_proto_goTypes = []interface{}{
	(*Media)(nil),       // 0: chameleon.api.media.Media
	(*Media_Image)(nil), // 1: chameleon.api.media.Media.Image
	(*Media_Video)(nil), // 2: chameleon.api.media.Media.Video
	(*anypb.Any)(nil),   // 3: google.protobuf.Any
}
var file_chameleon_api_media_data_proto_depIdxs = []int32{
	3, // 0: chameleon.api.media.Media.detail:type_name -> google.protobuf.Any
	1, // 1: chameleon.api.media.Media.Video.cover:type_name -> chameleon.api.media.Media.Image
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_chameleon_api_media_data_proto_init() }
func file_chameleon_api_media_data_proto_init() {
	if File_chameleon_api_media_data_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_chameleon_api_media_data_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Media); i {
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
		file_chameleon_api_media_data_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Media_Image); i {
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
		file_chameleon_api_media_data_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Media_Video); i {
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
			RawDescriptor: file_chameleon_api_media_data_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_chameleon_api_media_data_proto_goTypes,
		DependencyIndexes: file_chameleon_api_media_data_proto_depIdxs,
		MessageInfos:      file_chameleon_api_media_data_proto_msgTypes,
	}.Build()
	File_chameleon_api_media_data_proto = out.File
	file_chameleon_api_media_data_proto_rawDesc = nil
	file_chameleon_api_media_data_proto_goTypes = nil
	file_chameleon_api_media_data_proto_depIdxs = nil
}