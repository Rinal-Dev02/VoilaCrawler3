// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0-devel
// 	protoc        v3.14.0
// source: chameleon/smelter/v1/crawl/item/youtube.proto

package item

import (
	_ "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/api/http"
	media "github.com/voiladev/VoilaCrawl/protoc-gen-go/chameleon/api/media"
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

// Youtube
type Youtube struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *Youtube) Reset() {
	*x = Youtube{}
	if protoimpl.UnsafeEnabled {
		mi := &file_chameleon_smelter_v1_crawl_item_youtube_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Youtube) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Youtube) ProtoMessage() {}

func (x *Youtube) ProtoReflect() protoreflect.Message {
	mi := &file_chameleon_smelter_v1_crawl_item_youtube_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Youtube.ProtoReflect.Descriptor instead.
func (*Youtube) Descriptor() ([]byte, []int) {
	return file_chameleon_smelter_v1_crawl_item_youtube_proto_rawDescGZIP(), []int{0}
}

// Source
type Youtube_Source struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Id
	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	// CrawlUrl
	CrawlUrl string `protobuf:"bytes,2,opt,name=crawlUrl,proto3" json:"crawlUrl,omitempty"`
	// SourceUrl
	SourceUrl string `protobuf:"bytes,5,opt,name=sourceUrl,proto3" json:"sourceUrl,omitempty"`
	// PublishUtc
	PublishUtc int64 `protobuf:"varint,15,opt,name=publishUtc,proto3" json:"publishUtc,omitempty"`
}

func (x *Youtube_Source) Reset() {
	*x = Youtube_Source{}
	if protoimpl.UnsafeEnabled {
		mi := &file_chameleon_smelter_v1_crawl_item_youtube_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Youtube_Source) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Youtube_Source) ProtoMessage() {}

func (x *Youtube_Source) ProtoReflect() protoreflect.Message {
	mi := &file_chameleon_smelter_v1_crawl_item_youtube_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Youtube_Source.ProtoReflect.Descriptor instead.
func (*Youtube_Source) Descriptor() ([]byte, []int) {
	return file_chameleon_smelter_v1_crawl_item_youtube_proto_rawDescGZIP(), []int{0, 0}
}

func (x *Youtube_Source) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *Youtube_Source) GetCrawlUrl() string {
	if x != nil {
		return x.CrawlUrl
	}
	return ""
}

func (x *Youtube_Source) GetSourceUrl() string {
	if x != nil {
		return x.SourceUrl
	}
	return ""
}

func (x *Youtube_Source) GetPublishUtc() int64 {
	if x != nil {
		return x.PublishUtc
	}
	return 0
}

// Channel
type Youtube_Channel struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Id channel id
	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	// CanonicalUrl
	CanonicalUrl string `protobuf:"bytes,2,opt,name=canonicalUrl,proto3" json:"canonicalUrl,omitempty"`
	// Username
	Username string `protobuf:"bytes,3,opt,name=username,proto3" json:"username,omitempty"`
	// Nickname
	Title string `protobuf:"bytes,4,opt,name=title,proto3" json:"title,omitempty"`
	// Avatar
	Avatar string `protobuf:"bytes,5,opt,name=avatar,proto3" json:"avatar,omitempty"`
	// Description
	Description string `protobuf:"bytes,6,opt,name=description,proto3" json:"description,omitempty"`
	// Country
	Country string `protobuf:"bytes,7,opt,name=country,proto3" json:"country,omitempty"`
	// Stats
	Stats *Youtube_Channel_Stats `protobuf:"bytes,11,opt,name=stats,proto3" json:"stats,omitempty"`
	// PublishedUtc
	PublishedUtc int64 `protobuf:"varint,15,opt,name=publishedUtc,proto3" json:"publishedUtc,omitempty"`
}

func (x *Youtube_Channel) Reset() {
	*x = Youtube_Channel{}
	if protoimpl.UnsafeEnabled {
		mi := &file_chameleon_smelter_v1_crawl_item_youtube_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Youtube_Channel) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Youtube_Channel) ProtoMessage() {}

func (x *Youtube_Channel) ProtoReflect() protoreflect.Message {
	mi := &file_chameleon_smelter_v1_crawl_item_youtube_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Youtube_Channel.ProtoReflect.Descriptor instead.
func (*Youtube_Channel) Descriptor() ([]byte, []int) {
	return file_chameleon_smelter_v1_crawl_item_youtube_proto_rawDescGZIP(), []int{0, 1}
}

func (x *Youtube_Channel) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *Youtube_Channel) GetCanonicalUrl() string {
	if x != nil {
		return x.CanonicalUrl
	}
	return ""
}

func (x *Youtube_Channel) GetUsername() string {
	if x != nil {
		return x.Username
	}
	return ""
}

func (x *Youtube_Channel) GetTitle() string {
	if x != nil {
		return x.Title
	}
	return ""
}

func (x *Youtube_Channel) GetAvatar() string {
	if x != nil {
		return x.Avatar
	}
	return ""
}

func (x *Youtube_Channel) GetDescription() string {
	if x != nil {
		return x.Description
	}
	return ""
}

func (x *Youtube_Channel) GetCountry() string {
	if x != nil {
		return x.Country
	}
	return ""
}

func (x *Youtube_Channel) GetStats() *Youtube_Channel_Stats {
	if x != nil {
		return x.Stats
	}
	return nil
}

func (x *Youtube_Channel) GetPublishedUtc() int64 {
	if x != nil {
		return x.PublishedUtc
	}
	return 0
}

// Stats
type Youtube_Stats struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// SubscribeCount
	SubscribeCount int32 `protobuf:"varint,1,opt,name=subscribeCount,proto3" json:"subscribeCount,omitempty"`
	// VideoCount
	VideoCount int32 `protobuf:"varint,2,opt,name=videoCount,proto3" json:"videoCount,omitempty"`
	// ViewCount
	ViewCount int32 `protobuf:"varint,3,opt,name=viewCount,proto3" json:"viewCount,omitempty"`
	// CommentCount
	CommentCount int32 `protobuf:"varint,5,opt,name=commentCount,proto3" json:"commentCount,omitempty"`
}

func (x *Youtube_Stats) Reset() {
	*x = Youtube_Stats{}
	if protoimpl.UnsafeEnabled {
		mi := &file_chameleon_smelter_v1_crawl_item_youtube_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Youtube_Stats) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Youtube_Stats) ProtoMessage() {}

func (x *Youtube_Stats) ProtoReflect() protoreflect.Message {
	mi := &file_chameleon_smelter_v1_crawl_item_youtube_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Youtube_Stats.ProtoReflect.Descriptor instead.
func (*Youtube_Stats) Descriptor() ([]byte, []int) {
	return file_chameleon_smelter_v1_crawl_item_youtube_proto_rawDescGZIP(), []int{0, 2}
}

func (x *Youtube_Stats) GetSubscribeCount() int32 {
	if x != nil {
		return x.SubscribeCount
	}
	return 0
}

func (x *Youtube_Stats) GetVideoCount() int32 {
	if x != nil {
		return x.VideoCount
	}
	return 0
}

func (x *Youtube_Stats) GetViewCount() int32 {
	if x != nil {
		return x.ViewCount
	}
	return 0
}

func (x *Youtube_Stats) GetCommentCount() int32 {
	if x != nil {
		return x.CommentCount
	}
	return 0
}

// Video
type Youtube_Video struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Source
	Source *Youtube_Source `protobuf:"bytes,2,opt,name=source,proto3" json:"source,omitempty"`
	// Title
	Title string `protobuf:"bytes,4,opt,name=title,proto3" json:"title,omitempty"`
	// Description
	Description string `protobuf:"bytes,5,opt,name=description,proto3" json:"description,omitempty"`
	// Channel
	Channel *Youtube_Channel `protobuf:"bytes,6,opt,name=channel,proto3" json:"channel,omitempty"`
	// Video
	Video *media.Media_Video `protobuf:"bytes,11,opt,name=video,proto3" json:"video,omitempty"`
	// Header
	Headers map[string]string `protobuf:"bytes,15,rep,name=headers,proto3" json:"headers,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	// Stats
	Stats *Youtube_Stats `protobuf:"bytes,21,opt,name=stats,proto3" json:"stats,omitempty"`
	// CrawledUtc
	CrawledUtc int64 `protobuf:"varint,31,opt,name=crawledUtc,proto3" json:"crawledUtc,omitempty"`
	// ExpiresUtc which decided by cookie or url expire time
	ExpiresUtc int64 `protobuf:"varint,32,opt,name=expiresUtc,proto3" json:"expiresUtc,omitempty"`
}

func (x *Youtube_Video) Reset() {
	*x = Youtube_Video{}
	if protoimpl.UnsafeEnabled {
		mi := &file_chameleon_smelter_v1_crawl_item_youtube_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Youtube_Video) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Youtube_Video) ProtoMessage() {}

func (x *Youtube_Video) ProtoReflect() protoreflect.Message {
	mi := &file_chameleon_smelter_v1_crawl_item_youtube_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Youtube_Video.ProtoReflect.Descriptor instead.
func (*Youtube_Video) Descriptor() ([]byte, []int) {
	return file_chameleon_smelter_v1_crawl_item_youtube_proto_rawDescGZIP(), []int{0, 3}
}

func (x *Youtube_Video) GetSource() *Youtube_Source {
	if x != nil {
		return x.Source
	}
	return nil
}

func (x *Youtube_Video) GetTitle() string {
	if x != nil {
		return x.Title
	}
	return ""
}

func (x *Youtube_Video) GetDescription() string {
	if x != nil {
		return x.Description
	}
	return ""
}

func (x *Youtube_Video) GetChannel() *Youtube_Channel {
	if x != nil {
		return x.Channel
	}
	return nil
}

func (x *Youtube_Video) GetVideo() *media.Media_Video {
	if x != nil {
		return x.Video
	}
	return nil
}

func (x *Youtube_Video) GetHeaders() map[string]string {
	if x != nil {
		return x.Headers
	}
	return nil
}

func (x *Youtube_Video) GetStats() *Youtube_Stats {
	if x != nil {
		return x.Stats
	}
	return nil
}

func (x *Youtube_Video) GetCrawledUtc() int64 {
	if x != nil {
		return x.CrawledUtc
	}
	return 0
}

func (x *Youtube_Video) GetExpiresUtc() int64 {
	if x != nil {
		return x.ExpiresUtc
	}
	return 0
}

// Stats
type Youtube_Channel_Stats struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// SubscribeCount
	SubscribeCount int32 `protobuf:"varint,1,opt,name=subscribeCount,proto3" json:"subscribeCount,omitempty"`
	// VideoCount
	VideoCount int32 `protobuf:"varint,2,opt,name=videoCount,proto3" json:"videoCount,omitempty"`
	// ViewCount
	ViewCount int32 `protobuf:"varint,3,opt,name=viewCount,proto3" json:"viewCount,omitempty"`
	// CommentCount
	CommentCount int32 `protobuf:"varint,5,opt,name=commentCount,proto3" json:"commentCount,omitempty"`
}

func (x *Youtube_Channel_Stats) Reset() {
	*x = Youtube_Channel_Stats{}
	if protoimpl.UnsafeEnabled {
		mi := &file_chameleon_smelter_v1_crawl_item_youtube_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Youtube_Channel_Stats) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Youtube_Channel_Stats) ProtoMessage() {}

func (x *Youtube_Channel_Stats) ProtoReflect() protoreflect.Message {
	mi := &file_chameleon_smelter_v1_crawl_item_youtube_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Youtube_Channel_Stats.ProtoReflect.Descriptor instead.
func (*Youtube_Channel_Stats) Descriptor() ([]byte, []int) {
	return file_chameleon_smelter_v1_crawl_item_youtube_proto_rawDescGZIP(), []int{0, 1, 0}
}

func (x *Youtube_Channel_Stats) GetSubscribeCount() int32 {
	if x != nil {
		return x.SubscribeCount
	}
	return 0
}

func (x *Youtube_Channel_Stats) GetVideoCount() int32 {
	if x != nil {
		return x.VideoCount
	}
	return 0
}

func (x *Youtube_Channel_Stats) GetViewCount() int32 {
	if x != nil {
		return x.ViewCount
	}
	return 0
}

func (x *Youtube_Channel_Stats) GetCommentCount() int32 {
	if x != nil {
		return x.CommentCount
	}
	return 0
}

var File_chameleon_smelter_v1_crawl_item_youtube_proto protoreflect.FileDescriptor

var file_chameleon_smelter_v1_crawl_item_youtube_proto_rawDesc = []byte{
	0x0a, 0x2d, 0x63, 0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x2f, 0x73, 0x6d, 0x65, 0x6c,
	0x74, 0x65, 0x72, 0x2f, 0x76, 0x31, 0x2f, 0x63, 0x72, 0x61, 0x77, 0x6c, 0x2f, 0x69, 0x74, 0x65,
	0x6d, 0x2f, 0x79, 0x6f, 0x75, 0x74, 0x75, 0x62, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x1f, 0x63, 0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x2e, 0x73, 0x6d, 0x65, 0x6c, 0x74,
	0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x63, 0x72, 0x61, 0x77, 0x6c, 0x2e, 0x69, 0x74, 0x65, 0x6d,
	0x1a, 0x1e, 0x63, 0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x2f, 0x61, 0x70, 0x69, 0x2f,
	0x6d, 0x65, 0x64, 0x69, 0x61, 0x2f, 0x64, 0x61, 0x74, 0x61, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x1a, 0x1d, 0x63, 0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x2f, 0x61, 0x70, 0x69, 0x2f,
	0x68, 0x74, 0x74, 0x70, 0x2f, 0x64, 0x61, 0x74, 0x61, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22,
	0x85, 0x0a, 0x0a, 0x07, 0x59, 0x6f, 0x75, 0x74, 0x75, 0x62, 0x65, 0x1a, 0x72, 0x0a, 0x06, 0x53,
	0x6f, 0x75, 0x72, 0x63, 0x65, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x1a, 0x0a, 0x08, 0x63, 0x72, 0x61, 0x77, 0x6c, 0x55, 0x72,
	0x6c, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x63, 0x72, 0x61, 0x77, 0x6c, 0x55, 0x72,
	0x6c, 0x12, 0x1c, 0x0a, 0x09, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x55, 0x72, 0x6c, 0x18, 0x05,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x55, 0x72, 0x6c, 0x12,
	0x1e, 0x0a, 0x0a, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x73, 0x68, 0x55, 0x74, 0x63, 0x18, 0x0f, 0x20,
	0x01, 0x28, 0x03, 0x52, 0x0a, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x73, 0x68, 0x55, 0x74, 0x63, 0x1a,
	0xc9, 0x03, 0x0a, 0x07, 0x43, 0x68, 0x61, 0x6e, 0x6e, 0x65, 0x6c, 0x12, 0x0e, 0x0a, 0x02, 0x69,
	0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x22, 0x0a, 0x0c, 0x63,
	0x61, 0x6e, 0x6f, 0x6e, 0x69, 0x63, 0x61, 0x6c, 0x55, 0x72, 0x6c, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x0c, 0x63, 0x61, 0x6e, 0x6f, 0x6e, 0x69, 0x63, 0x61, 0x6c, 0x55, 0x72, 0x6c, 0x12,
	0x1a, 0x0a, 0x08, 0x75, 0x73, 0x65, 0x72, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x08, 0x75, 0x73, 0x65, 0x72, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x74,
	0x69, 0x74, 0x6c, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x74, 0x69, 0x74, 0x6c,
	0x65, 0x12, 0x16, 0x0a, 0x06, 0x61, 0x76, 0x61, 0x74, 0x61, 0x72, 0x18, 0x05, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x06, 0x61, 0x76, 0x61, 0x74, 0x61, 0x72, 0x12, 0x20, 0x0a, 0x0b, 0x64, 0x65, 0x73,
	0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b,
	0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x18, 0x0a, 0x07, 0x63,
	0x6f, 0x75, 0x6e, 0x74, 0x72, 0x79, 0x18, 0x07, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x63, 0x6f,
	0x75, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x4c, 0x0a, 0x05, 0x73, 0x74, 0x61, 0x74, 0x73, 0x18, 0x0b,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x36, 0x2e, 0x63, 0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e,
	0x2e, 0x73, 0x6d, 0x65, 0x6c, 0x74, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x63, 0x72, 0x61, 0x77,
	0x6c, 0x2e, 0x69, 0x74, 0x65, 0x6d, 0x2e, 0x59, 0x6f, 0x75, 0x74, 0x75, 0x62, 0x65, 0x2e, 0x43,
	0x68, 0x61, 0x6e, 0x6e, 0x65, 0x6c, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x73, 0x52, 0x05, 0x73, 0x74,
	0x61, 0x74, 0x73, 0x12, 0x22, 0x0a, 0x0c, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x73, 0x68, 0x65, 0x64,
	0x55, 0x74, 0x63, 0x18, 0x0f, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0c, 0x70, 0x75, 0x62, 0x6c, 0x69,
	0x73, 0x68, 0x65, 0x64, 0x55, 0x74, 0x63, 0x1a, 0x91, 0x01, 0x0a, 0x05, 0x53, 0x74, 0x61, 0x74,
	0x73, 0x12, 0x26, 0x0a, 0x0e, 0x73, 0x75, 0x62, 0x73, 0x63, 0x72, 0x69, 0x62, 0x65, 0x43, 0x6f,
	0x75, 0x6e, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0e, 0x73, 0x75, 0x62, 0x73, 0x63,
	0x72, 0x69, 0x62, 0x65, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x1e, 0x0a, 0x0a, 0x76, 0x69, 0x64,
	0x65, 0x6f, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0a, 0x76,
	0x69, 0x64, 0x65, 0x6f, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x1c, 0x0a, 0x09, 0x76, 0x69, 0x65,
	0x77, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x05, 0x52, 0x09, 0x76, 0x69,
	0x65, 0x77, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x22, 0x0a, 0x0c, 0x63, 0x6f, 0x6d, 0x6d, 0x65,
	0x6e, 0x74, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x05, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0c, 0x63,
	0x6f, 0x6d, 0x6d, 0x65, 0x6e, 0x74, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x1a, 0x91, 0x01, 0x0a, 0x05,
	0x53, 0x74, 0x61, 0x74, 0x73, 0x12, 0x26, 0x0a, 0x0e, 0x73, 0x75, 0x62, 0x73, 0x63, 0x72, 0x69,
	0x62, 0x65, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0e, 0x73,
	0x75, 0x62, 0x73, 0x63, 0x72, 0x69, 0x62, 0x65, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x1e, 0x0a,
	0x0a, 0x76, 0x69, 0x64, 0x65, 0x6f, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x05, 0x52, 0x0a, 0x76, 0x69, 0x64, 0x65, 0x6f, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x1c, 0x0a,
	0x09, 0x76, 0x69, 0x65, 0x77, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x05,
	0x52, 0x09, 0x76, 0x69, 0x65, 0x77, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x22, 0x0a, 0x0c, 0x63,
	0x6f, 0x6d, 0x6d, 0x65, 0x6e, 0x74, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x05, 0x20, 0x01, 0x28,
	0x05, 0x52, 0x0c, 0x63, 0x6f, 0x6d, 0x6d, 0x65, 0x6e, 0x74, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x1a,
	0xa5, 0x04, 0x0a, 0x05, 0x56, 0x69, 0x64, 0x65, 0x6f, 0x12, 0x47, 0x0a, 0x06, 0x73, 0x6f, 0x75,
	0x72, 0x63, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x2f, 0x2e, 0x63, 0x68, 0x61, 0x6d,
	0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x2e, 0x73, 0x6d, 0x65, 0x6c, 0x74, 0x65, 0x72, 0x2e, 0x76, 0x31,
	0x2e, 0x63, 0x72, 0x61, 0x77, 0x6c, 0x2e, 0x69, 0x74, 0x65, 0x6d, 0x2e, 0x59, 0x6f, 0x75, 0x74,
	0x75, 0x62, 0x65, 0x2e, 0x53, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x52, 0x06, 0x73, 0x6f, 0x75, 0x72,
	0x63, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x74, 0x69, 0x74, 0x6c, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x05, 0x74, 0x69, 0x74, 0x6c, 0x65, 0x12, 0x20, 0x0a, 0x0b, 0x64, 0x65, 0x73, 0x63,
	0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x64,
	0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x4a, 0x0a, 0x07, 0x63, 0x68,
	0x61, 0x6e, 0x6e, 0x65, 0x6c, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x30, 0x2e, 0x63, 0x68,
	0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x2e, 0x73, 0x6d, 0x65, 0x6c, 0x74, 0x65, 0x72, 0x2e,
	0x76, 0x31, 0x2e, 0x63, 0x72, 0x61, 0x77, 0x6c, 0x2e, 0x69, 0x74, 0x65, 0x6d, 0x2e, 0x59, 0x6f,
	0x75, 0x74, 0x75, 0x62, 0x65, 0x2e, 0x43, 0x68, 0x61, 0x6e, 0x6e, 0x65, 0x6c, 0x52, 0x07, 0x63,
	0x68, 0x61, 0x6e, 0x6e, 0x65, 0x6c, 0x12, 0x36, 0x0a, 0x05, 0x76, 0x69, 0x64, 0x65, 0x6f, 0x18,
	0x0b, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x20, 0x2e, 0x63, 0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f,
	0x6e, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x6d, 0x65, 0x64, 0x69, 0x61, 0x2e, 0x4d, 0x65, 0x64, 0x69,
	0x61, 0x2e, 0x56, 0x69, 0x64, 0x65, 0x6f, 0x52, 0x05, 0x76, 0x69, 0x64, 0x65, 0x6f, 0x12, 0x55,
	0x0a, 0x07, 0x68, 0x65, 0x61, 0x64, 0x65, 0x72, 0x73, 0x18, 0x0f, 0x20, 0x03, 0x28, 0x0b, 0x32,
	0x3b, 0x2e, 0x63, 0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x2e, 0x73, 0x6d, 0x65, 0x6c,
	0x74, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x63, 0x72, 0x61, 0x77, 0x6c, 0x2e, 0x69, 0x74, 0x65,
	0x6d, 0x2e, 0x59, 0x6f, 0x75, 0x74, 0x75, 0x62, 0x65, 0x2e, 0x56, 0x69, 0x64, 0x65, 0x6f, 0x2e,
	0x48, 0x65, 0x61, 0x64, 0x65, 0x72, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x07, 0x68, 0x65,
	0x61, 0x64, 0x65, 0x72, 0x73, 0x12, 0x44, 0x0a, 0x05, 0x73, 0x74, 0x61, 0x74, 0x73, 0x18, 0x15,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x2e, 0x2e, 0x63, 0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e,
	0x2e, 0x73, 0x6d, 0x65, 0x6c, 0x74, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x63, 0x72, 0x61, 0x77,
	0x6c, 0x2e, 0x69, 0x74, 0x65, 0x6d, 0x2e, 0x59, 0x6f, 0x75, 0x74, 0x75, 0x62, 0x65, 0x2e, 0x53,
	0x74, 0x61, 0x74, 0x73, 0x52, 0x05, 0x73, 0x74, 0x61, 0x74, 0x73, 0x12, 0x1e, 0x0a, 0x0a, 0x63,
	0x72, 0x61, 0x77, 0x6c, 0x65, 0x64, 0x55, 0x74, 0x63, 0x18, 0x1f, 0x20, 0x01, 0x28, 0x03, 0x52,
	0x0a, 0x63, 0x72, 0x61, 0x77, 0x6c, 0x65, 0x64, 0x55, 0x74, 0x63, 0x12, 0x1e, 0x0a, 0x0a, 0x65,
	0x78, 0x70, 0x69, 0x72, 0x65, 0x73, 0x55, 0x74, 0x63, 0x18, 0x20, 0x20, 0x01, 0x28, 0x03, 0x52,
	0x0a, 0x65, 0x78, 0x70, 0x69, 0x72, 0x65, 0x73, 0x55, 0x74, 0x63, 0x1a, 0x3a, 0x0a, 0x0c, 0x48,
	0x65, 0x61, 0x64, 0x65, 0x72, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b,
	0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a,
	0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61,
	0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x42, 0x26, 0x5a, 0x24, 0x63, 0x68, 0x61, 0x6d, 0x65,
	0x6c, 0x65, 0x6f, 0x6e, 0x2f, 0x73, 0x6d, 0x65, 0x6c, 0x74, 0x65, 0x72, 0x2f, 0x76, 0x31, 0x2f,
	0x63, 0x72, 0x61, 0x77, 0x6c, 0x2f, 0x69, 0x74, 0x65, 0x6d, 0x3b, 0x69, 0x74, 0x65, 0x6d, 0x62,
	0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_chameleon_smelter_v1_crawl_item_youtube_proto_rawDescOnce sync.Once
	file_chameleon_smelter_v1_crawl_item_youtube_proto_rawDescData = file_chameleon_smelter_v1_crawl_item_youtube_proto_rawDesc
)

func file_chameleon_smelter_v1_crawl_item_youtube_proto_rawDescGZIP() []byte {
	file_chameleon_smelter_v1_crawl_item_youtube_proto_rawDescOnce.Do(func() {
		file_chameleon_smelter_v1_crawl_item_youtube_proto_rawDescData = protoimpl.X.CompressGZIP(file_chameleon_smelter_v1_crawl_item_youtube_proto_rawDescData)
	})
	return file_chameleon_smelter_v1_crawl_item_youtube_proto_rawDescData
}

var file_chameleon_smelter_v1_crawl_item_youtube_proto_msgTypes = make([]protoimpl.MessageInfo, 7)
var file_chameleon_smelter_v1_crawl_item_youtube_proto_goTypes = []interface{}{
	(*Youtube)(nil),               // 0: chameleon.smelter.v1.crawl.item.Youtube
	(*Youtube_Source)(nil),        // 1: chameleon.smelter.v1.crawl.item.Youtube.Source
	(*Youtube_Channel)(nil),       // 2: chameleon.smelter.v1.crawl.item.Youtube.Channel
	(*Youtube_Stats)(nil),         // 3: chameleon.smelter.v1.crawl.item.Youtube.Stats
	(*Youtube_Video)(nil),         // 4: chameleon.smelter.v1.crawl.item.Youtube.Video
	(*Youtube_Channel_Stats)(nil), // 5: chameleon.smelter.v1.crawl.item.Youtube.Channel.Stats
	nil,                           // 6: chameleon.smelter.v1.crawl.item.Youtube.Video.HeadersEntry
	(*media.Media_Video)(nil),     // 7: chameleon.api.media.Media.Video
}
var file_chameleon_smelter_v1_crawl_item_youtube_proto_depIdxs = []int32{
	5, // 0: chameleon.smelter.v1.crawl.item.Youtube.Channel.stats:type_name -> chameleon.smelter.v1.crawl.item.Youtube.Channel.Stats
	1, // 1: chameleon.smelter.v1.crawl.item.Youtube.Video.source:type_name -> chameleon.smelter.v1.crawl.item.Youtube.Source
	2, // 2: chameleon.smelter.v1.crawl.item.Youtube.Video.channel:type_name -> chameleon.smelter.v1.crawl.item.Youtube.Channel
	7, // 3: chameleon.smelter.v1.crawl.item.Youtube.Video.video:type_name -> chameleon.api.media.Media.Video
	6, // 4: chameleon.smelter.v1.crawl.item.Youtube.Video.headers:type_name -> chameleon.smelter.v1.crawl.item.Youtube.Video.HeadersEntry
	3, // 5: chameleon.smelter.v1.crawl.item.Youtube.Video.stats:type_name -> chameleon.smelter.v1.crawl.item.Youtube.Stats
	6, // [6:6] is the sub-list for method output_type
	6, // [6:6] is the sub-list for method input_type
	6, // [6:6] is the sub-list for extension type_name
	6, // [6:6] is the sub-list for extension extendee
	0, // [0:6] is the sub-list for field type_name
}

func init() { file_chameleon_smelter_v1_crawl_item_youtube_proto_init() }
func file_chameleon_smelter_v1_crawl_item_youtube_proto_init() {
	if File_chameleon_smelter_v1_crawl_item_youtube_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_chameleon_smelter_v1_crawl_item_youtube_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Youtube); i {
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
		file_chameleon_smelter_v1_crawl_item_youtube_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Youtube_Source); i {
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
		file_chameleon_smelter_v1_crawl_item_youtube_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Youtube_Channel); i {
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
		file_chameleon_smelter_v1_crawl_item_youtube_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Youtube_Stats); i {
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
		file_chameleon_smelter_v1_crawl_item_youtube_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Youtube_Video); i {
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
		file_chameleon_smelter_v1_crawl_item_youtube_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Youtube_Channel_Stats); i {
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
			RawDescriptor: file_chameleon_smelter_v1_crawl_item_youtube_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   7,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_chameleon_smelter_v1_crawl_item_youtube_proto_goTypes,
		DependencyIndexes: file_chameleon_smelter_v1_crawl_item_youtube_proto_depIdxs,
		MessageInfos:      file_chameleon_smelter_v1_crawl_item_youtube_proto_msgTypes,
	}.Build()
	File_chameleon_smelter_v1_crawl_item_youtube_proto = out.File
	file_chameleon_smelter_v1_crawl_item_youtube_proto_rawDesc = nil
	file_chameleon_smelter_v1_crawl_item_youtube_proto_goTypes = nil
	file_chameleon_smelter_v1_crawl_item_youtube_proto_depIdxs = nil
}
