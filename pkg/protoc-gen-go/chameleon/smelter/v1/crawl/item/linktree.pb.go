// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        v3.18.0
// source: chameleon/smelter/v1/crawl/item/linktree.proto

package item

import (
	_ "github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/api/http"
	_ "github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/api/media"
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

// Linktree
type Linktree struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *Linktree) Reset() {
	*x = Linktree{}
	if protoimpl.UnsafeEnabled {
		mi := &file_chameleon_smelter_v1_crawl_item_linktree_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Linktree) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Linktree) ProtoMessage() {}

func (x *Linktree) ProtoReflect() protoreflect.Message {
	mi := &file_chameleon_smelter_v1_crawl_item_linktree_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Linktree.ProtoReflect.Descriptor instead.
func (*Linktree) Descriptor() ([]byte, []int) {
	return file_chameleon_smelter_v1_crawl_item_linktree_proto_rawDescGZIP(), []int{0}
}

// Item
type Linktree_Item struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Profile
	Profile *Linktree_Item_Profile `protobuf:"bytes,5,opt,name=profile,proto3" json:"profile,omitempty"`
	// Links
	Links []*Linktree_Item_Link `protobuf:"bytes,6,rep,name=links,proto3" json:"links,omitempty"`
	// SocialLinks
	SocialLinks []*Linktree_Item_SocialLink `protobuf:"bytes,7,rep,name=socialLinks,proto3" json:"socialLinks,omitempty"`
}

func (x *Linktree_Item) Reset() {
	*x = Linktree_Item{}
	if protoimpl.UnsafeEnabled {
		mi := &file_chameleon_smelter_v1_crawl_item_linktree_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Linktree_Item) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Linktree_Item) ProtoMessage() {}

func (x *Linktree_Item) ProtoReflect() protoreflect.Message {
	mi := &file_chameleon_smelter_v1_crawl_item_linktree_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Linktree_Item.ProtoReflect.Descriptor instead.
func (*Linktree_Item) Descriptor() ([]byte, []int) {
	return file_chameleon_smelter_v1_crawl_item_linktree_proto_rawDescGZIP(), []int{0, 0}
}

func (x *Linktree_Item) GetProfile() *Linktree_Item_Profile {
	if x != nil {
		return x.Profile
	}
	return nil
}

func (x *Linktree_Item) GetLinks() []*Linktree_Item_Link {
	if x != nil {
		return x.Links
	}
	return nil
}

func (x *Linktree_Item) GetSocialLinks() []*Linktree_Item_SocialLink {
	if x != nil {
		return x.SocialLinks
	}
	return nil
}

// Profile
type Linktree_Item_Profile struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// ID
	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	// Name
	Name string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	// Avatar
	Avatar string `protobuf:"bytes,3,opt,name=avatar,proto3" json:"avatar,omitempty"`
	// Email
	Email string `protobuf:"bytes,8,opt,name=email,proto3" json:"email,omitempty"`
	// LinktreeUrl
	LinktreeUrl string `protobuf:"bytes,11,opt,name=linktreeUrl,proto3" json:"linktreeUrl,omitempty"`
}

func (x *Linktree_Item_Profile) Reset() {
	*x = Linktree_Item_Profile{}
	if protoimpl.UnsafeEnabled {
		mi := &file_chameleon_smelter_v1_crawl_item_linktree_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Linktree_Item_Profile) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Linktree_Item_Profile) ProtoMessage() {}

func (x *Linktree_Item_Profile) ProtoReflect() protoreflect.Message {
	mi := &file_chameleon_smelter_v1_crawl_item_linktree_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Linktree_Item_Profile.ProtoReflect.Descriptor instead.
func (*Linktree_Item_Profile) Descriptor() ([]byte, []int) {
	return file_chameleon_smelter_v1_crawl_item_linktree_proto_rawDescGZIP(), []int{0, 0, 0}
}

func (x *Linktree_Item_Profile) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *Linktree_Item_Profile) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Linktree_Item_Profile) GetAvatar() string {
	if x != nil {
		return x.Avatar
	}
	return ""
}

func (x *Linktree_Item_Profile) GetEmail() string {
	if x != nil {
		return x.Email
	}
	return ""
}

func (x *Linktree_Item_Profile) GetLinktreeUrl() string {
	if x != nil {
		return x.LinktreeUrl
	}
	return ""
}

// Link
type Linktree_Item_Link struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Id
	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	// Title
	Title string `protobuf:"bytes,2,opt,name=title,proto3" json:"title,omitempty"`
	// Url
	Url string `protobuf:"bytes,3,opt,name=url,proto3" json:"url,omitempty"`
	// Type
	Type string `protobuf:"bytes,4,opt,name=type,proto3" json:"type,omitempty"`
	// Icon
	Icon string `protobuf:"bytes,6,opt,name=icon,proto3" json:"icon,omitempty"`
	// Style
	Style string `protobuf:"bytes,7,opt,name=style,proto3" json:"style,omitempty"`
}

func (x *Linktree_Item_Link) Reset() {
	*x = Linktree_Item_Link{}
	if protoimpl.UnsafeEnabled {
		mi := &file_chameleon_smelter_v1_crawl_item_linktree_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Linktree_Item_Link) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Linktree_Item_Link) ProtoMessage() {}

func (x *Linktree_Item_Link) ProtoReflect() protoreflect.Message {
	mi := &file_chameleon_smelter_v1_crawl_item_linktree_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Linktree_Item_Link.ProtoReflect.Descriptor instead.
func (*Linktree_Item_Link) Descriptor() ([]byte, []int) {
	return file_chameleon_smelter_v1_crawl_item_linktree_proto_rawDescGZIP(), []int{0, 0, 1}
}

func (x *Linktree_Item_Link) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *Linktree_Item_Link) GetTitle() string {
	if x != nil {
		return x.Title
	}
	return ""
}

func (x *Linktree_Item_Link) GetUrl() string {
	if x != nil {
		return x.Url
	}
	return ""
}

func (x *Linktree_Item_Link) GetType() string {
	if x != nil {
		return x.Type
	}
	return ""
}

func (x *Linktree_Item_Link) GetIcon() string {
	if x != nil {
		return x.Icon
	}
	return ""
}

func (x *Linktree_Item_Link) GetStyle() string {
	if x != nil {
		return x.Style
	}
	return ""
}

// SocialLink
type Linktree_Item_SocialLink struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Type
	Type string `protobuf:"bytes,2,opt,name=type,proto3" json:"type,omitempty"`
	// Title
	Title string `protobuf:"bytes,3,opt,name=title,proto3" json:"title,omitempty"`
	// Url
	Url string `protobuf:"bytes,4,opt,name=url,proto3" json:"url,omitempty"`
}

func (x *Linktree_Item_SocialLink) Reset() {
	*x = Linktree_Item_SocialLink{}
	if protoimpl.UnsafeEnabled {
		mi := &file_chameleon_smelter_v1_crawl_item_linktree_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Linktree_Item_SocialLink) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Linktree_Item_SocialLink) ProtoMessage() {}

func (x *Linktree_Item_SocialLink) ProtoReflect() protoreflect.Message {
	mi := &file_chameleon_smelter_v1_crawl_item_linktree_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Linktree_Item_SocialLink.ProtoReflect.Descriptor instead.
func (*Linktree_Item_SocialLink) Descriptor() ([]byte, []int) {
	return file_chameleon_smelter_v1_crawl_item_linktree_proto_rawDescGZIP(), []int{0, 0, 2}
}

func (x *Linktree_Item_SocialLink) GetType() string {
	if x != nil {
		return x.Type
	}
	return ""
}

func (x *Linktree_Item_SocialLink) GetTitle() string {
	if x != nil {
		return x.Title
	}
	return ""
}

func (x *Linktree_Item_SocialLink) GetUrl() string {
	if x != nil {
		return x.Url
	}
	return ""
}

var File_chameleon_smelter_v1_crawl_item_linktree_proto protoreflect.FileDescriptor

var file_chameleon_smelter_v1_crawl_item_linktree_proto_rawDesc = []byte{
	0x0a, 0x2e, 0x63, 0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x2f, 0x73, 0x6d, 0x65, 0x6c,
	0x74, 0x65, 0x72, 0x2f, 0x76, 0x31, 0x2f, 0x63, 0x72, 0x61, 0x77, 0x6c, 0x2f, 0x69, 0x74, 0x65,
	0x6d, 0x2f, 0x6c, 0x69, 0x6e, 0x6b, 0x74, 0x72, 0x65, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x12, 0x1f, 0x63, 0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x2e, 0x73, 0x6d, 0x65, 0x6c,
	0x74, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x63, 0x72, 0x61, 0x77, 0x6c, 0x2e, 0x69, 0x74, 0x65,
	0x6d, 0x1a, 0x1e, 0x63, 0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x2f, 0x61, 0x70, 0x69,
	0x2f, 0x6d, 0x65, 0x64, 0x69, 0x61, 0x2f, 0x64, 0x61, 0x74, 0x61, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x1a, 0x1d, 0x63, 0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x2f, 0x61, 0x70, 0x69,
	0x2f, 0x68, 0x74, 0x74, 0x70, 0x2f, 0x64, 0x61, 0x74, 0x61, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x22, 0xd4, 0x04, 0x0a, 0x08, 0x4c, 0x69, 0x6e, 0x6b, 0x74, 0x72, 0x65, 0x65, 0x1a, 0xc7, 0x04,
	0x0a, 0x04, 0x49, 0x74, 0x65, 0x6d, 0x12, 0x50, 0x0a, 0x07, 0x70, 0x72, 0x6f, 0x66, 0x69, 0x6c,
	0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x36, 0x2e, 0x63, 0x68, 0x61, 0x6d, 0x65, 0x6c,
	0x65, 0x6f, 0x6e, 0x2e, 0x73, 0x6d, 0x65, 0x6c, 0x74, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x63,
	0x72, 0x61, 0x77, 0x6c, 0x2e, 0x69, 0x74, 0x65, 0x6d, 0x2e, 0x4c, 0x69, 0x6e, 0x6b, 0x74, 0x72,
	0x65, 0x65, 0x2e, 0x49, 0x74, 0x65, 0x6d, 0x2e, 0x50, 0x72, 0x6f, 0x66, 0x69, 0x6c, 0x65, 0x52,
	0x07, 0x70, 0x72, 0x6f, 0x66, 0x69, 0x6c, 0x65, 0x12, 0x49, 0x0a, 0x05, 0x6c, 0x69, 0x6e, 0x6b,
	0x73, 0x18, 0x06, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x33, 0x2e, 0x63, 0x68, 0x61, 0x6d, 0x65, 0x6c,
	0x65, 0x6f, 0x6e, 0x2e, 0x73, 0x6d, 0x65, 0x6c, 0x74, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x63,
	0x72, 0x61, 0x77, 0x6c, 0x2e, 0x69, 0x74, 0x65, 0x6d, 0x2e, 0x4c, 0x69, 0x6e, 0x6b, 0x74, 0x72,
	0x65, 0x65, 0x2e, 0x49, 0x74, 0x65, 0x6d, 0x2e, 0x4c, 0x69, 0x6e, 0x6b, 0x52, 0x05, 0x6c, 0x69,
	0x6e, 0x6b, 0x73, 0x12, 0x5b, 0x0a, 0x0b, 0x73, 0x6f, 0x63, 0x69, 0x61, 0x6c, 0x4c, 0x69, 0x6e,
	0x6b, 0x73, 0x18, 0x07, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x39, 0x2e, 0x63, 0x68, 0x61, 0x6d, 0x65,
	0x6c, 0x65, 0x6f, 0x6e, 0x2e, 0x73, 0x6d, 0x65, 0x6c, 0x74, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e,
	0x63, 0x72, 0x61, 0x77, 0x6c, 0x2e, 0x69, 0x74, 0x65, 0x6d, 0x2e, 0x4c, 0x69, 0x6e, 0x6b, 0x74,
	0x72, 0x65, 0x65, 0x2e, 0x49, 0x74, 0x65, 0x6d, 0x2e, 0x53, 0x6f, 0x63, 0x69, 0x61, 0x6c, 0x4c,
	0x69, 0x6e, 0x6b, 0x52, 0x0b, 0x73, 0x6f, 0x63, 0x69, 0x61, 0x6c, 0x4c, 0x69, 0x6e, 0x6b, 0x73,
	0x1a, 0x7d, 0x0a, 0x07, 0x50, 0x72, 0x6f, 0x66, 0x69, 0x6c, 0x65, 0x12, 0x0e, 0x0a, 0x02, 0x69,
	0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x6e,
	0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12,
	0x16, 0x0a, 0x06, 0x61, 0x76, 0x61, 0x74, 0x61, 0x72, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x06, 0x61, 0x76, 0x61, 0x74, 0x61, 0x72, 0x12, 0x14, 0x0a, 0x05, 0x65, 0x6d, 0x61, 0x69, 0x6c,
	0x18, 0x08, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x65, 0x6d, 0x61, 0x69, 0x6c, 0x12, 0x20, 0x0a,
	0x0b, 0x6c, 0x69, 0x6e, 0x6b, 0x74, 0x72, 0x65, 0x65, 0x55, 0x72, 0x6c, 0x18, 0x0b, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x0b, 0x6c, 0x69, 0x6e, 0x6b, 0x74, 0x72, 0x65, 0x65, 0x55, 0x72, 0x6c, 0x1a,
	0x7c, 0x0a, 0x04, 0x4c, 0x69, 0x6e, 0x6b, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x14, 0x0a, 0x05, 0x74, 0x69, 0x74, 0x6c, 0x65,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x74, 0x69, 0x74, 0x6c, 0x65, 0x12, 0x10, 0x0a,
	0x03, 0x75, 0x72, 0x6c, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x75, 0x72, 0x6c, 0x12,
	0x12, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x74,
	0x79, 0x70, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x69, 0x63, 0x6f, 0x6e, 0x18, 0x06, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x04, 0x69, 0x63, 0x6f, 0x6e, 0x12, 0x14, 0x0a, 0x05, 0x73, 0x74, 0x79, 0x6c, 0x65,
	0x18, 0x07, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x73, 0x74, 0x79, 0x6c, 0x65, 0x1a, 0x48, 0x0a,
	0x0a, 0x53, 0x6f, 0x63, 0x69, 0x61, 0x6c, 0x4c, 0x69, 0x6e, 0x6b, 0x12, 0x12, 0x0a, 0x04, 0x74,
	0x79, 0x70, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x74, 0x79, 0x70, 0x65, 0x12,
	0x14, 0x0a, 0x05, 0x74, 0x69, 0x74, 0x6c, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05,
	0x74, 0x69, 0x74, 0x6c, 0x65, 0x12, 0x10, 0x0a, 0x03, 0x75, 0x72, 0x6c, 0x18, 0x04, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x03, 0x75, 0x72, 0x6c, 0x42, 0x26, 0x5a, 0x24, 0x63, 0x68, 0x61, 0x6d, 0x65,
	0x6c, 0x65, 0x6f, 0x6e, 0x2f, 0x73, 0x6d, 0x65, 0x6c, 0x74, 0x65, 0x72, 0x2f, 0x76, 0x31, 0x2f,
	0x63, 0x72, 0x61, 0x77, 0x6c, 0x2f, 0x69, 0x74, 0x65, 0x6d, 0x3b, 0x69, 0x74, 0x65, 0x6d, 0x62,
	0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_chameleon_smelter_v1_crawl_item_linktree_proto_rawDescOnce sync.Once
	file_chameleon_smelter_v1_crawl_item_linktree_proto_rawDescData = file_chameleon_smelter_v1_crawl_item_linktree_proto_rawDesc
)

func file_chameleon_smelter_v1_crawl_item_linktree_proto_rawDescGZIP() []byte {
	file_chameleon_smelter_v1_crawl_item_linktree_proto_rawDescOnce.Do(func() {
		file_chameleon_smelter_v1_crawl_item_linktree_proto_rawDescData = protoimpl.X.CompressGZIP(file_chameleon_smelter_v1_crawl_item_linktree_proto_rawDescData)
	})
	return file_chameleon_smelter_v1_crawl_item_linktree_proto_rawDescData
}

var file_chameleon_smelter_v1_crawl_item_linktree_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_chameleon_smelter_v1_crawl_item_linktree_proto_goTypes = []interface{}{
	(*Linktree)(nil),                 // 0: chameleon.smelter.v1.crawl.item.Linktree
	(*Linktree_Item)(nil),            // 1: chameleon.smelter.v1.crawl.item.Linktree.Item
	(*Linktree_Item_Profile)(nil),    // 2: chameleon.smelter.v1.crawl.item.Linktree.Item.Profile
	(*Linktree_Item_Link)(nil),       // 3: chameleon.smelter.v1.crawl.item.Linktree.Item.Link
	(*Linktree_Item_SocialLink)(nil), // 4: chameleon.smelter.v1.crawl.item.Linktree.Item.SocialLink
}
var file_chameleon_smelter_v1_crawl_item_linktree_proto_depIdxs = []int32{
	2, // 0: chameleon.smelter.v1.crawl.item.Linktree.Item.profile:type_name -> chameleon.smelter.v1.crawl.item.Linktree.Item.Profile
	3, // 1: chameleon.smelter.v1.crawl.item.Linktree.Item.links:type_name -> chameleon.smelter.v1.crawl.item.Linktree.Item.Link
	4, // 2: chameleon.smelter.v1.crawl.item.Linktree.Item.socialLinks:type_name -> chameleon.smelter.v1.crawl.item.Linktree.Item.SocialLink
	3, // [3:3] is the sub-list for method output_type
	3, // [3:3] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_chameleon_smelter_v1_crawl_item_linktree_proto_init() }
func file_chameleon_smelter_v1_crawl_item_linktree_proto_init() {
	if File_chameleon_smelter_v1_crawl_item_linktree_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_chameleon_smelter_v1_crawl_item_linktree_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Linktree); i {
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
		file_chameleon_smelter_v1_crawl_item_linktree_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Linktree_Item); i {
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
		file_chameleon_smelter_v1_crawl_item_linktree_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Linktree_Item_Profile); i {
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
		file_chameleon_smelter_v1_crawl_item_linktree_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Linktree_Item_Link); i {
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
		file_chameleon_smelter_v1_crawl_item_linktree_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Linktree_Item_SocialLink); i {
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
			RawDescriptor: file_chameleon_smelter_v1_crawl_item_linktree_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_chameleon_smelter_v1_crawl_item_linktree_proto_goTypes,
		DependencyIndexes: file_chameleon_smelter_v1_crawl_item_linktree_proto_depIdxs,
		MessageInfos:      file_chameleon_smelter_v1_crawl_item_linktree_proto_msgTypes,
	}.Build()
	File_chameleon_smelter_v1_crawl_item_linktree_proto = out.File
	file_chameleon_smelter_v1_crawl_item_linktree_proto_rawDesc = nil
	file_chameleon_smelter_v1_crawl_item_linktree_proto_goTypes = nil
	file_chameleon_smelter_v1_crawl_item_linktree_proto_depIdxs = nil
}