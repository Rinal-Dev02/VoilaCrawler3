// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.14.0
// source: chameleon/api/event/data.proto

package event

import (
	proto "github.com/golang/protobuf/proto"
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

// This is a compile-time assertion that a sufficiently up-to-date version
// of the legacy proto package is being used.
const _ = proto.ProtoPackageIsVersion4

// EventType
type EventType int32

const (
	// Unknown
	EventType_Unknown EventType = 0
	// Created
	EventType_Created EventType = 1
	// Updated
	EventType_Updated EventType = 2
	// Deleted
	EventType_Deleted EventType = 3
	// Expired
	EventType_Expired EventType = 11
	// Custom
	EventType_Custom EventType = 12
)

// Enum value maps for EventType.
var (
	EventType_name = map[int32]string{
		0:  "Unknown",
		1:  "Created",
		2:  "Updated",
		3:  "Deleted",
		11: "Expired",
		12: "Custom",
	}
	EventType_value = map[string]int32{
		"Unknown": 0,
		"Created": 1,
		"Updated": 2,
		"Deleted": 3,
		"Expired": 11,
		"Custom":  12,
	}
)

func (x EventType) Enum() *EventType {
	p := new(EventType)
	*p = x
	return p
}

func (x EventType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (EventType) Descriptor() protoreflect.EnumDescriptor {
	return file_chameleon_api_event_data_proto_enumTypes[0].Descriptor()
}

func (EventType) Type() protoreflect.EnumType {
	return &file_chameleon_api_event_data_proto_enumTypes[0]
}

func (x EventType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use EventType.Descriptor instead.
func (EventType) EnumDescriptor() ([]byte, []int) {
	return file_chameleon_api_event_data_proto_rawDescGZIP(), []int{0}
}

// Header
type Header struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Key
	Key string `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	// Values
	Value string `protobuf:"bytes,2,opt,name=value,proto3" json:"value,omitempty"`
}

func (x *Header) Reset() {
	*x = Header{}
	if protoimpl.UnsafeEnabled {
		mi := &file_chameleon_api_event_data_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Header) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Header) ProtoMessage() {}

func (x *Header) ProtoReflect() protoreflect.Message {
	mi := &file_chameleon_api_event_data_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Header.ProtoReflect.Descriptor instead.
func (*Header) Descriptor() ([]byte, []int) {
	return file_chameleon_api_event_data_proto_rawDescGZIP(), []int{0}
}

func (x *Header) GetKey() string {
	if x != nil {
		return x.Key
	}
	return ""
}

func (x *Header) GetValue() string {
	if x != nil {
		return x.Value
	}
	return ""
}

// Event
type Event struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Id
	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	// EventType
	Type EventType `protobuf:"varint,2,opt,name=type,proto3,enum=chameleon.api.event.EventType" json:"type,omitempty"`
	// Headers
	Headers map[string]string `protobuf:"bytes,3,rep,name=headers,proto3" json:"headers,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	// Data
	Data []byte `protobuf:"bytes,6,opt,name=data,proto3" json:"data,omitempty"`
	// Timestamp
	Timestamp int64 `protobuf:"varint,11,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
}

func (x *Event) Reset() {
	*x = Event{}
	if protoimpl.UnsafeEnabled {
		mi := &file_chameleon_api_event_data_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Event) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Event) ProtoMessage() {}

func (x *Event) ProtoReflect() protoreflect.Message {
	mi := &file_chameleon_api_event_data_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Event.ProtoReflect.Descriptor instead.
func (*Event) Descriptor() ([]byte, []int) {
	return file_chameleon_api_event_data_proto_rawDescGZIP(), []int{1}
}

func (x *Event) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *Event) GetType() EventType {
	if x != nil {
		return x.Type
	}
	return EventType_Unknown
}

func (x *Event) GetHeaders() map[string]string {
	if x != nil {
		return x.Headers
	}
	return nil
}

func (x *Event) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

func (x *Event) GetTimestamp() int64 {
	if x != nil {
		return x.Timestamp
	}
	return 0
}

var File_chameleon_api_event_data_proto protoreflect.FileDescriptor

var file_chameleon_api_event_data_proto_rawDesc = []byte{
	0x0a, 0x1e, 0x63, 0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x2f, 0x61, 0x70, 0x69, 0x2f,
	0x65, 0x76, 0x65, 0x6e, 0x74, 0x2f, 0x64, 0x61, 0x74, 0x61, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x12, 0x13, 0x63, 0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x2e, 0x61, 0x70, 0x69, 0x2e,
	0x65, 0x76, 0x65, 0x6e, 0x74, 0x22, 0x30, 0x0a, 0x06, 0x48, 0x65, 0x61, 0x64, 0x65, 0x72, 0x12,
	0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65,
	0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x22, 0xfc, 0x01, 0x0a, 0x05, 0x45, 0x76, 0x65, 0x6e,
	0x74, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69,
	0x64, 0x12, 0x32, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0e, 0x32,
	0x1e, 0x2e, 0x63, 0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x2e, 0x61, 0x70, 0x69, 0x2e,
	0x65, 0x76, 0x65, 0x6e, 0x74, 0x2e, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x54, 0x79, 0x70, 0x65, 0x52,
	0x04, 0x74, 0x79, 0x70, 0x65, 0x12, 0x41, 0x0a, 0x07, 0x68, 0x65, 0x61, 0x64, 0x65, 0x72, 0x73,
	0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x27, 0x2e, 0x63, 0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65,
	0x6f, 0x6e, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x2e, 0x45, 0x76, 0x65,
	0x6e, 0x74, 0x2e, 0x48, 0x65, 0x61, 0x64, 0x65, 0x72, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52,
	0x07, 0x68, 0x65, 0x61, 0x64, 0x65, 0x72, 0x73, 0x12, 0x12, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61,
	0x18, 0x06, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x12, 0x1c, 0x0a, 0x09,
	0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x18, 0x0b, 0x20, 0x01, 0x28, 0x03, 0x52,
	0x09, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x1a, 0x3a, 0x0a, 0x0c, 0x48, 0x65,
	0x61, 0x64, 0x65, 0x72, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65,
	0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05,
	0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c,
	0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x2a, 0x58, 0x0a, 0x09, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x54,
	0x79, 0x70, 0x65, 0x12, 0x0b, 0x0a, 0x07, 0x55, 0x6e, 0x6b, 0x6e, 0x6f, 0x77, 0x6e, 0x10, 0x00,
	0x12, 0x0b, 0x0a, 0x07, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x10, 0x01, 0x12, 0x0b, 0x0a,
	0x07, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x64, 0x10, 0x02, 0x12, 0x0b, 0x0a, 0x07, 0x44, 0x65,
	0x6c, 0x65, 0x74, 0x65, 0x64, 0x10, 0x03, 0x12, 0x0b, 0x0a, 0x07, 0x45, 0x78, 0x70, 0x69, 0x72,
	0x65, 0x64, 0x10, 0x0b, 0x12, 0x0a, 0x0a, 0x06, 0x43, 0x75, 0x73, 0x74, 0x6f, 0x6d, 0x10, 0x0c,
	0x42, 0x1b, 0x5a, 0x19, 0x63, 0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x2f, 0x61, 0x70,
	0x69, 0x2f, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x3b, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x62, 0x06, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_chameleon_api_event_data_proto_rawDescOnce sync.Once
	file_chameleon_api_event_data_proto_rawDescData = file_chameleon_api_event_data_proto_rawDesc
)

func file_chameleon_api_event_data_proto_rawDescGZIP() []byte {
	file_chameleon_api_event_data_proto_rawDescOnce.Do(func() {
		file_chameleon_api_event_data_proto_rawDescData = protoimpl.X.CompressGZIP(file_chameleon_api_event_data_proto_rawDescData)
	})
	return file_chameleon_api_event_data_proto_rawDescData
}

var file_chameleon_api_event_data_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_chameleon_api_event_data_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_chameleon_api_event_data_proto_goTypes = []interface{}{
	(EventType)(0), // 0: chameleon.api.event.EventType
	(*Header)(nil), // 1: chameleon.api.event.Header
	(*Event)(nil),  // 2: chameleon.api.event.Event
	nil,            // 3: chameleon.api.event.Event.HeadersEntry
}
var file_chameleon_api_event_data_proto_depIdxs = []int32{
	0, // 0: chameleon.api.event.Event.type:type_name -> chameleon.api.event.EventType
	3, // 1: chameleon.api.event.Event.headers:type_name -> chameleon.api.event.Event.HeadersEntry
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_chameleon_api_event_data_proto_init() }
func file_chameleon_api_event_data_proto_init() {
	if File_chameleon_api_event_data_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_chameleon_api_event_data_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Header); i {
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
		file_chameleon_api_event_data_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Event); i {
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
			RawDescriptor: file_chameleon_api_event_data_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_chameleon_api_event_data_proto_goTypes,
		DependencyIndexes: file_chameleon_api_event_data_proto_depIdxs,
		EnumInfos:         file_chameleon_api_event_data_proto_enumTypes,
		MessageInfos:      file_chameleon_api_event_data_proto_msgTypes,
	}.Build()
	File_chameleon_api_event_data_proto = out.File
	file_chameleon_api_event_data_proto_rawDesc = nil
	file_chameleon_api_event_data_proto_goTypes = nil
	file_chameleon_api_event_data_proto_depIdxs = nil
}
