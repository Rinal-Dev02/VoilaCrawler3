// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        v3.18.0
// source: chameleon/smelter/v1/crawl/item/tiktok.proto

package item

import (
	_ "github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/api/http"
	media "github.com/voiladev/VoilaCrawler/pkg/protoc-gen-go/chameleon/api/media"
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

// Tiktok
type Tiktok struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *Tiktok) Reset() {
	*x = Tiktok{}
	if protoimpl.UnsafeEnabled {
		mi := &file_chameleon_smelter_v1_crawl_item_tiktok_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Tiktok) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Tiktok) ProtoMessage() {}

func (x *Tiktok) ProtoReflect() protoreflect.Message {
	mi := &file_chameleon_smelter_v1_crawl_item_tiktok_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Tiktok.ProtoReflect.Descriptor instead.
func (*Tiktok) Descriptor() ([]byte, []int) {
	return file_chameleon_smelter_v1_crawl_item_tiktok_proto_rawDescGZIP(), []int{0}
}

// Source
type Tiktok_Source struct {
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

func (x *Tiktok_Source) Reset() {
	*x = Tiktok_Source{}
	if protoimpl.UnsafeEnabled {
		mi := &file_chameleon_smelter_v1_crawl_item_tiktok_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Tiktok_Source) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Tiktok_Source) ProtoMessage() {}

func (x *Tiktok_Source) ProtoReflect() protoreflect.Message {
	mi := &file_chameleon_smelter_v1_crawl_item_tiktok_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Tiktok_Source.ProtoReflect.Descriptor instead.
func (*Tiktok_Source) Descriptor() ([]byte, []int) {
	return file_chameleon_smelter_v1_crawl_item_tiktok_proto_rawDescGZIP(), []int{0, 0}
}

func (x *Tiktok_Source) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *Tiktok_Source) GetCrawlUrl() string {
	if x != nil {
		return x.CrawlUrl
	}
	return ""
}

func (x *Tiktok_Source) GetSourceUrl() string {
	if x != nil {
		return x.SourceUrl
	}
	return ""
}

func (x *Tiktok_Source) GetPublishUtc() int64 {
	if x != nil {
		return x.PublishUtc
	}
	return 0
}

// Author
type Tiktok_Author struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Id
	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	// Name
	Name string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	// Nickname
	Nickname string `protobuf:"bytes,3,opt,name=nickname,proto3" json:"nickname,omitempty"`
	// Avatar
	Avatar string `protobuf:"bytes,4,opt,name=avatar,proto3" json:"avatar,omitempty"`
	// Description
	Description string `protobuf:"bytes,6,opt,name=description,proto3" json:"description,omitempty"`
	// Stats
	Stats *Tiktok_Author_Stats `protobuf:"bytes,11,opt,name=stats,proto3" json:"stats,omitempty"`
	// RegisterUtc
	RegisterUtc int64 `protobuf:"varint,15,opt,name=registerUtc,proto3" json:"registerUtc,omitempty"`
}

func (x *Tiktok_Author) Reset() {
	*x = Tiktok_Author{}
	if protoimpl.UnsafeEnabled {
		mi := &file_chameleon_smelter_v1_crawl_item_tiktok_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Tiktok_Author) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Tiktok_Author) ProtoMessage() {}

func (x *Tiktok_Author) ProtoReflect() protoreflect.Message {
	mi := &file_chameleon_smelter_v1_crawl_item_tiktok_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Tiktok_Author.ProtoReflect.Descriptor instead.
func (*Tiktok_Author) Descriptor() ([]byte, []int) {
	return file_chameleon_smelter_v1_crawl_item_tiktok_proto_rawDescGZIP(), []int{0, 1}
}

func (x *Tiktok_Author) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *Tiktok_Author) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Tiktok_Author) GetNickname() string {
	if x != nil {
		return x.Nickname
	}
	return ""
}

func (x *Tiktok_Author) GetAvatar() string {
	if x != nil {
		return x.Avatar
	}
	return ""
}

func (x *Tiktok_Author) GetDescription() string {
	if x != nil {
		return x.Description
	}
	return ""
}

func (x *Tiktok_Author) GetStats() *Tiktok_Author_Stats {
	if x != nil {
		return x.Stats
	}
	return nil
}

func (x *Tiktok_Author) GetRegisterUtc() int64 {
	if x != nil {
		return x.RegisterUtc
	}
	return 0
}

// Stats
type Tiktok_Stats struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// DiggCount
	DiggCount int32 `protobuf:"varint,1,opt,name=diggCount,proto3" json:"diggCount,omitempty"`
	// ShareCount
	ShareCount int32 `protobuf:"varint,2,opt,name=shareCount,proto3" json:"shareCount,omitempty"`
	// CommentCount
	CommentCount int32 `protobuf:"varint,3,opt,name=commentCount,proto3" json:"commentCount,omitempty"`
	// PlayCount
	PlayCount int32 `protobuf:"varint,4,opt,name=playCount,proto3" json:"playCount,omitempty"`
}

func (x *Tiktok_Stats) Reset() {
	*x = Tiktok_Stats{}
	if protoimpl.UnsafeEnabled {
		mi := &file_chameleon_smelter_v1_crawl_item_tiktok_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Tiktok_Stats) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Tiktok_Stats) ProtoMessage() {}

func (x *Tiktok_Stats) ProtoReflect() protoreflect.Message {
	mi := &file_chameleon_smelter_v1_crawl_item_tiktok_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Tiktok_Stats.ProtoReflect.Descriptor instead.
func (*Tiktok_Stats) Descriptor() ([]byte, []int) {
	return file_chameleon_smelter_v1_crawl_item_tiktok_proto_rawDescGZIP(), []int{0, 2}
}

func (x *Tiktok_Stats) GetDiggCount() int32 {
	if x != nil {
		return x.DiggCount
	}
	return 0
}

func (x *Tiktok_Stats) GetShareCount() int32 {
	if x != nil {
		return x.ShareCount
	}
	return 0
}

func (x *Tiktok_Stats) GetCommentCount() int32 {
	if x != nil {
		return x.CommentCount
	}
	return 0
}

func (x *Tiktok_Stats) GetPlayCount() int32 {
	if x != nil {
		return x.PlayCount
	}
	return 0
}

// Item
type Tiktok_Item struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Source
	Source *Tiktok_Source `protobuf:"bytes,2,opt,name=source,proto3" json:"source,omitempty"`
	// Title
	Title string `protobuf:"bytes,4,opt,name=title,proto3" json:"title,omitempty"`
	// Description
	Description string `protobuf:"bytes,5,opt,name=description,proto3" json:"description,omitempty"`
	// Author
	Author *Tiktok_Author `protobuf:"bytes,6,opt,name=author,proto3" json:"author,omitempty"`
	// Video
	Video *media.Media_Video `protobuf:"bytes,11,opt,name=video,proto3" json:"video,omitempty"`
	// Header
	Headers map[string]string `protobuf:"bytes,15,rep,name=headers,proto3" json:"headers,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	// Stats
	Stats *Tiktok_Stats `protobuf:"bytes,21,opt,name=stats,proto3" json:"stats,omitempty"`
	// CrawledUtc
	CrawledUtc int64 `protobuf:"varint,31,opt,name=crawledUtc,proto3" json:"crawledUtc,omitempty"`
	// ExpiresUtc which decided by cookie or url expire time
	ExpiresUtc int64 `protobuf:"varint,32,opt,name=expiresUtc,proto3" json:"expiresUtc,omitempty"`
}

func (x *Tiktok_Item) Reset() {
	*x = Tiktok_Item{}
	if protoimpl.UnsafeEnabled {
		mi := &file_chameleon_smelter_v1_crawl_item_tiktok_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Tiktok_Item) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Tiktok_Item) ProtoMessage() {}

func (x *Tiktok_Item) ProtoReflect() protoreflect.Message {
	mi := &file_chameleon_smelter_v1_crawl_item_tiktok_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Tiktok_Item.ProtoReflect.Descriptor instead.
func (*Tiktok_Item) Descriptor() ([]byte, []int) {
	return file_chameleon_smelter_v1_crawl_item_tiktok_proto_rawDescGZIP(), []int{0, 3}
}

func (x *Tiktok_Item) GetSource() *Tiktok_Source {
	if x != nil {
		return x.Source
	}
	return nil
}

func (x *Tiktok_Item) GetTitle() string {
	if x != nil {
		return x.Title
	}
	return ""
}

func (x *Tiktok_Item) GetDescription() string {
	if x != nil {
		return x.Description
	}
	return ""
}

func (x *Tiktok_Item) GetAuthor() *Tiktok_Author {
	if x != nil {
		return x.Author
	}
	return nil
}

func (x *Tiktok_Item) GetVideo() *media.Media_Video {
	if x != nil {
		return x.Video
	}
	return nil
}

func (x *Tiktok_Item) GetHeaders() map[string]string {
	if x != nil {
		return x.Headers
	}
	return nil
}

func (x *Tiktok_Item) GetStats() *Tiktok_Stats {
	if x != nil {
		return x.Stats
	}
	return nil
}

func (x *Tiktok_Item) GetCrawledUtc() int64 {
	if x != nil {
		return x.CrawledUtc
	}
	return 0
}

func (x *Tiktok_Item) GetExpiresUtc() int64 {
	if x != nil {
		return x.ExpiresUtc
	}
	return 0
}

// Stats
type Tiktok_Author_Stats struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// FollowingCount
	FollowingCount int32 `protobuf:"varint,1,opt,name=followingCount,proto3" json:"followingCount,omitempty"`
	// FollowerCount
	FollowerCount int32 `protobuf:"varint,2,opt,name=followerCount,proto3" json:"followerCount,omitempty"`
	// LikeCount
	LikeCount int32 `protobuf:"varint,3,opt,name=likeCount,proto3" json:"likeCount,omitempty"`
	// VideoCount
	VideoCount int32 `protobuf:"varint,6,opt,name=videoCount,proto3" json:"videoCount,omitempty"`
	// Diggcount 用户推荐数
	DiggCount int32 `protobuf:"varint,7,opt,name=diggCount,proto3" json:"diggCount,omitempty"`
}

func (x *Tiktok_Author_Stats) Reset() {
	*x = Tiktok_Author_Stats{}
	if protoimpl.UnsafeEnabled {
		mi := &file_chameleon_smelter_v1_crawl_item_tiktok_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Tiktok_Author_Stats) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Tiktok_Author_Stats) ProtoMessage() {}

func (x *Tiktok_Author_Stats) ProtoReflect() protoreflect.Message {
	mi := &file_chameleon_smelter_v1_crawl_item_tiktok_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Tiktok_Author_Stats.ProtoReflect.Descriptor instead.
func (*Tiktok_Author_Stats) Descriptor() ([]byte, []int) {
	return file_chameleon_smelter_v1_crawl_item_tiktok_proto_rawDescGZIP(), []int{0, 1, 0}
}

func (x *Tiktok_Author_Stats) GetFollowingCount() int32 {
	if x != nil {
		return x.FollowingCount
	}
	return 0
}

func (x *Tiktok_Author_Stats) GetFollowerCount() int32 {
	if x != nil {
		return x.FollowerCount
	}
	return 0
}

func (x *Tiktok_Author_Stats) GetLikeCount() int32 {
	if x != nil {
		return x.LikeCount
	}
	return 0
}

func (x *Tiktok_Author_Stats) GetVideoCount() int32 {
	if x != nil {
		return x.VideoCount
	}
	return 0
}

func (x *Tiktok_Author_Stats) GetDiggCount() int32 {
	if x != nil {
		return x.DiggCount
	}
	return 0
}

var File_chameleon_smelter_v1_crawl_item_tiktok_proto protoreflect.FileDescriptor

var file_chameleon_smelter_v1_crawl_item_tiktok_proto_rawDesc = []byte{
	0x0a, 0x2c, 0x63, 0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x2f, 0x73, 0x6d, 0x65, 0x6c,
	0x74, 0x65, 0x72, 0x2f, 0x76, 0x31, 0x2f, 0x63, 0x72, 0x61, 0x77, 0x6c, 0x2f, 0x69, 0x74, 0x65,
	0x6d, 0x2f, 0x74, 0x69, 0x6b, 0x74, 0x6f, 0x6b, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x1f,
	0x63, 0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x2e, 0x73, 0x6d, 0x65, 0x6c, 0x74, 0x65,
	0x72, 0x2e, 0x76, 0x31, 0x2e, 0x63, 0x72, 0x61, 0x77, 0x6c, 0x2e, 0x69, 0x74, 0x65, 0x6d, 0x1a,
	0x1e, 0x63, 0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x6d,
	0x65, 0x64, 0x69, 0x61, 0x2f, 0x64, 0x61, 0x74, 0x61, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a,
	0x1d, 0x63, 0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x68,
	0x74, 0x74, 0x70, 0x2f, 0x64, 0x61, 0x74, 0x61, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xcc,
	0x09, 0x0a, 0x06, 0x54, 0x69, 0x6b, 0x74, 0x6f, 0x6b, 0x1a, 0x72, 0x0a, 0x06, 0x53, 0x6f, 0x75,
	0x72, 0x63, 0x65, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x02, 0x69, 0x64, 0x12, 0x1a, 0x0a, 0x08, 0x63, 0x72, 0x61, 0x77, 0x6c, 0x55, 0x72, 0x6c, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x63, 0x72, 0x61, 0x77, 0x6c, 0x55, 0x72, 0x6c, 0x12,
	0x1c, 0x0a, 0x09, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x55, 0x72, 0x6c, 0x18, 0x05, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x09, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x55, 0x72, 0x6c, 0x12, 0x1e, 0x0a,
	0x0a, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x73, 0x68, 0x55, 0x74, 0x63, 0x18, 0x0f, 0x20, 0x01, 0x28,
	0x03, 0x52, 0x0a, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x73, 0x68, 0x55, 0x74, 0x63, 0x1a, 0xa4, 0x03,
	0x0a, 0x06, 0x41, 0x75, 0x74, 0x68, 0x6f, 0x72, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x1a, 0x0a, 0x08,
	0x6e, 0x69, 0x63, 0x6b, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08,
	0x6e, 0x69, 0x63, 0x6b, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x61, 0x76, 0x61, 0x74,
	0x61, 0x72, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x61, 0x76, 0x61, 0x74, 0x61, 0x72,
	0x12, 0x20, 0x0a, 0x0b, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x18,
	0x06, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69,
	0x6f, 0x6e, 0x12, 0x4a, 0x0a, 0x05, 0x73, 0x74, 0x61, 0x74, 0x73, 0x18, 0x0b, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x34, 0x2e, 0x63, 0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x2e, 0x73, 0x6d,
	0x65, 0x6c, 0x74, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x63, 0x72, 0x61, 0x77, 0x6c, 0x2e, 0x69,
	0x74, 0x65, 0x6d, 0x2e, 0x54, 0x69, 0x6b, 0x74, 0x6f, 0x6b, 0x2e, 0x41, 0x75, 0x74, 0x68, 0x6f,
	0x72, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x73, 0x52, 0x05, 0x73, 0x74, 0x61, 0x74, 0x73, 0x12, 0x20,
	0x0a, 0x0b, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x65, 0x72, 0x55, 0x74, 0x63, 0x18, 0x0f, 0x20,
	0x01, 0x28, 0x03, 0x52, 0x0b, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x65, 0x72, 0x55, 0x74, 0x63,
	0x1a, 0xb1, 0x01, 0x0a, 0x05, 0x53, 0x74, 0x61, 0x74, 0x73, 0x12, 0x26, 0x0a, 0x0e, 0x66, 0x6f,
	0x6c, 0x6c, 0x6f, 0x77, 0x69, 0x6e, 0x67, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x05, 0x52, 0x0e, 0x66, 0x6f, 0x6c, 0x6c, 0x6f, 0x77, 0x69, 0x6e, 0x67, 0x43, 0x6f, 0x75,
	0x6e, 0x74, 0x12, 0x24, 0x0a, 0x0d, 0x66, 0x6f, 0x6c, 0x6c, 0x6f, 0x77, 0x65, 0x72, 0x43, 0x6f,
	0x75, 0x6e, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0d, 0x66, 0x6f, 0x6c, 0x6c, 0x6f,
	0x77, 0x65, 0x72, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x1c, 0x0a, 0x09, 0x6c, 0x69, 0x6b, 0x65,
	0x43, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x05, 0x52, 0x09, 0x6c, 0x69, 0x6b,
	0x65, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x1e, 0x0a, 0x0a, 0x76, 0x69, 0x64, 0x65, 0x6f, 0x43,
	0x6f, 0x75, 0x6e, 0x74, 0x18, 0x06, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0a, 0x76, 0x69, 0x64, 0x65,
	0x6f, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x1c, 0x0a, 0x09, 0x64, 0x69, 0x67, 0x67, 0x43, 0x6f,
	0x75, 0x6e, 0x74, 0x18, 0x07, 0x20, 0x01, 0x28, 0x05, 0x52, 0x09, 0x64, 0x69, 0x67, 0x67, 0x43,
	0x6f, 0x75, 0x6e, 0x74, 0x1a, 0x87, 0x01, 0x0a, 0x05, 0x53, 0x74, 0x61, 0x74, 0x73, 0x12, 0x1c,
	0x0a, 0x09, 0x64, 0x69, 0x67, 0x67, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x05, 0x52, 0x09, 0x64, 0x69, 0x67, 0x67, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x1e, 0x0a, 0x0a,
	0x73, 0x68, 0x61, 0x72, 0x65, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05,
	0x52, 0x0a, 0x73, 0x68, 0x61, 0x72, 0x65, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x22, 0x0a, 0x0c,
	0x63, 0x6f, 0x6d, 0x6d, 0x65, 0x6e, 0x74, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x05, 0x52, 0x0c, 0x63, 0x6f, 0x6d, 0x6d, 0x65, 0x6e, 0x74, 0x43, 0x6f, 0x75, 0x6e, 0x74,
	0x12, 0x1c, 0x0a, 0x09, 0x70, 0x6c, 0x61, 0x79, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x04, 0x20,
	0x01, 0x28, 0x05, 0x52, 0x09, 0x70, 0x6c, 0x61, 0x79, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x1a, 0x9c,
	0x04, 0x0a, 0x04, 0x49, 0x74, 0x65, 0x6d, 0x12, 0x46, 0x0a, 0x06, 0x73, 0x6f, 0x75, 0x72, 0x63,
	0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x2e, 0x2e, 0x63, 0x68, 0x61, 0x6d, 0x65, 0x6c,
	0x65, 0x6f, 0x6e, 0x2e, 0x73, 0x6d, 0x65, 0x6c, 0x74, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x63,
	0x72, 0x61, 0x77, 0x6c, 0x2e, 0x69, 0x74, 0x65, 0x6d, 0x2e, 0x54, 0x69, 0x6b, 0x74, 0x6f, 0x6b,
	0x2e, 0x53, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x52, 0x06, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x12,
	0x14, 0x0a, 0x05, 0x74, 0x69, 0x74, 0x6c, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05,
	0x74, 0x69, 0x74, 0x6c, 0x65, 0x12, 0x20, 0x0a, 0x0b, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70,
	0x74, 0x69, 0x6f, 0x6e, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x64, 0x65, 0x73, 0x63,
	0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x46, 0x0a, 0x06, 0x61, 0x75, 0x74, 0x68, 0x6f,
	0x72, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x2e, 0x2e, 0x63, 0x68, 0x61, 0x6d, 0x65, 0x6c,
	0x65, 0x6f, 0x6e, 0x2e, 0x73, 0x6d, 0x65, 0x6c, 0x74, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x63,
	0x72, 0x61, 0x77, 0x6c, 0x2e, 0x69, 0x74, 0x65, 0x6d, 0x2e, 0x54, 0x69, 0x6b, 0x74, 0x6f, 0x6b,
	0x2e, 0x41, 0x75, 0x74, 0x68, 0x6f, 0x72, 0x52, 0x06, 0x61, 0x75, 0x74, 0x68, 0x6f, 0x72, 0x12,
	0x36, 0x0a, 0x05, 0x76, 0x69, 0x64, 0x65, 0x6f, 0x18, 0x0b, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x20,
	0x2e, 0x63, 0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x6d,
	0x65, 0x64, 0x69, 0x61, 0x2e, 0x4d, 0x65, 0x64, 0x69, 0x61, 0x2e, 0x56, 0x69, 0x64, 0x65, 0x6f,
	0x52, 0x05, 0x76, 0x69, 0x64, 0x65, 0x6f, 0x12, 0x53, 0x0a, 0x07, 0x68, 0x65, 0x61, 0x64, 0x65,
	0x72, 0x73, 0x18, 0x0f, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x39, 0x2e, 0x63, 0x68, 0x61, 0x6d, 0x65,
	0x6c, 0x65, 0x6f, 0x6e, 0x2e, 0x73, 0x6d, 0x65, 0x6c, 0x74, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e,
	0x63, 0x72, 0x61, 0x77, 0x6c, 0x2e, 0x69, 0x74, 0x65, 0x6d, 0x2e, 0x54, 0x69, 0x6b, 0x74, 0x6f,
	0x6b, 0x2e, 0x49, 0x74, 0x65, 0x6d, 0x2e, 0x48, 0x65, 0x61, 0x64, 0x65, 0x72, 0x73, 0x45, 0x6e,
	0x74, 0x72, 0x79, 0x52, 0x07, 0x68, 0x65, 0x61, 0x64, 0x65, 0x72, 0x73, 0x12, 0x43, 0x0a, 0x05,
	0x73, 0x74, 0x61, 0x74, 0x73, 0x18, 0x15, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x2d, 0x2e, 0x63, 0x68,
	0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x2e, 0x73, 0x6d, 0x65, 0x6c, 0x74, 0x65, 0x72, 0x2e,
	0x76, 0x31, 0x2e, 0x63, 0x72, 0x61, 0x77, 0x6c, 0x2e, 0x69, 0x74, 0x65, 0x6d, 0x2e, 0x54, 0x69,
	0x6b, 0x74, 0x6f, 0x6b, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x73, 0x52, 0x05, 0x73, 0x74, 0x61, 0x74,
	0x73, 0x12, 0x1e, 0x0a, 0x0a, 0x63, 0x72, 0x61, 0x77, 0x6c, 0x65, 0x64, 0x55, 0x74, 0x63, 0x18,
	0x1f, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0a, 0x63, 0x72, 0x61, 0x77, 0x6c, 0x65, 0x64, 0x55, 0x74,
	0x63, 0x12, 0x1e, 0x0a, 0x0a, 0x65, 0x78, 0x70, 0x69, 0x72, 0x65, 0x73, 0x55, 0x74, 0x63, 0x18,
	0x20, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0a, 0x65, 0x78, 0x70, 0x69, 0x72, 0x65, 0x73, 0x55, 0x74,
	0x63, 0x1a, 0x3a, 0x0a, 0x0c, 0x48, 0x65, 0x61, 0x64, 0x65, 0x72, 0x73, 0x45, 0x6e, 0x74, 0x72,
	0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03,
	0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x42, 0x26, 0x5a,
	0x24, 0x63, 0x68, 0x61, 0x6d, 0x65, 0x6c, 0x65, 0x6f, 0x6e, 0x2f, 0x73, 0x6d, 0x65, 0x6c, 0x74,
	0x65, 0x72, 0x2f, 0x76, 0x31, 0x2f, 0x63, 0x72, 0x61, 0x77, 0x6c, 0x2f, 0x69, 0x74, 0x65, 0x6d,
	0x3b, 0x69, 0x74, 0x65, 0x6d, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_chameleon_smelter_v1_crawl_item_tiktok_proto_rawDescOnce sync.Once
	file_chameleon_smelter_v1_crawl_item_tiktok_proto_rawDescData = file_chameleon_smelter_v1_crawl_item_tiktok_proto_rawDesc
)

func file_chameleon_smelter_v1_crawl_item_tiktok_proto_rawDescGZIP() []byte {
	file_chameleon_smelter_v1_crawl_item_tiktok_proto_rawDescOnce.Do(func() {
		file_chameleon_smelter_v1_crawl_item_tiktok_proto_rawDescData = protoimpl.X.CompressGZIP(file_chameleon_smelter_v1_crawl_item_tiktok_proto_rawDescData)
	})
	return file_chameleon_smelter_v1_crawl_item_tiktok_proto_rawDescData
}

var file_chameleon_smelter_v1_crawl_item_tiktok_proto_msgTypes = make([]protoimpl.MessageInfo, 7)
var file_chameleon_smelter_v1_crawl_item_tiktok_proto_goTypes = []interface{}{
	(*Tiktok)(nil),              // 0: chameleon.smelter.v1.crawl.item.Tiktok
	(*Tiktok_Source)(nil),       // 1: chameleon.smelter.v1.crawl.item.Tiktok.Source
	(*Tiktok_Author)(nil),       // 2: chameleon.smelter.v1.crawl.item.Tiktok.Author
	(*Tiktok_Stats)(nil),        // 3: chameleon.smelter.v1.crawl.item.Tiktok.Stats
	(*Tiktok_Item)(nil),         // 4: chameleon.smelter.v1.crawl.item.Tiktok.Item
	(*Tiktok_Author_Stats)(nil), // 5: chameleon.smelter.v1.crawl.item.Tiktok.Author.Stats
	nil,                         // 6: chameleon.smelter.v1.crawl.item.Tiktok.Item.HeadersEntry
	(*media.Media_Video)(nil),   // 7: chameleon.api.media.Media.Video
}
var file_chameleon_smelter_v1_crawl_item_tiktok_proto_depIdxs = []int32{
	5, // 0: chameleon.smelter.v1.crawl.item.Tiktok.Author.stats:type_name -> chameleon.smelter.v1.crawl.item.Tiktok.Author.Stats
	1, // 1: chameleon.smelter.v1.crawl.item.Tiktok.Item.source:type_name -> chameleon.smelter.v1.crawl.item.Tiktok.Source
	2, // 2: chameleon.smelter.v1.crawl.item.Tiktok.Item.author:type_name -> chameleon.smelter.v1.crawl.item.Tiktok.Author
	7, // 3: chameleon.smelter.v1.crawl.item.Tiktok.Item.video:type_name -> chameleon.api.media.Media.Video
	6, // 4: chameleon.smelter.v1.crawl.item.Tiktok.Item.headers:type_name -> chameleon.smelter.v1.crawl.item.Tiktok.Item.HeadersEntry
	3, // 5: chameleon.smelter.v1.crawl.item.Tiktok.Item.stats:type_name -> chameleon.smelter.v1.crawl.item.Tiktok.Stats
	6, // [6:6] is the sub-list for method output_type
	6, // [6:6] is the sub-list for method input_type
	6, // [6:6] is the sub-list for extension type_name
	6, // [6:6] is the sub-list for extension extendee
	0, // [0:6] is the sub-list for field type_name
}

func init() { file_chameleon_smelter_v1_crawl_item_tiktok_proto_init() }
func file_chameleon_smelter_v1_crawl_item_tiktok_proto_init() {
	if File_chameleon_smelter_v1_crawl_item_tiktok_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_chameleon_smelter_v1_crawl_item_tiktok_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Tiktok); i {
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
		file_chameleon_smelter_v1_crawl_item_tiktok_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Tiktok_Source); i {
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
		file_chameleon_smelter_v1_crawl_item_tiktok_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Tiktok_Author); i {
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
		file_chameleon_smelter_v1_crawl_item_tiktok_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Tiktok_Stats); i {
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
		file_chameleon_smelter_v1_crawl_item_tiktok_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Tiktok_Item); i {
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
		file_chameleon_smelter_v1_crawl_item_tiktok_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Tiktok_Author_Stats); i {
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
			RawDescriptor: file_chameleon_smelter_v1_crawl_item_tiktok_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   7,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_chameleon_smelter_v1_crawl_item_tiktok_proto_goTypes,
		DependencyIndexes: file_chameleon_smelter_v1_crawl_item_tiktok_proto_depIdxs,
		MessageInfos:      file_chameleon_smelter_v1_crawl_item_tiktok_proto_msgTypes,
	}.Build()
	File_chameleon_smelter_v1_crawl_item_tiktok_proto = out.File
	file_chameleon_smelter_v1_crawl_item_tiktok_proto_rawDesc = nil
	file_chameleon_smelter_v1_crawl_item_tiktok_proto_goTypes = nil
	file_chameleon_smelter_v1_crawl_item_tiktok_proto_depIdxs = nil
}
