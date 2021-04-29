// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: crawler.proto

package types

import (
	fmt "fmt"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	io "io"
	math "math"
	math_bits "math/bits"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

// Crawler
type Crawler struct {
	// ID
	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	// Version
	Version int32 `protobuf:"varint,2,opt,name=version,proto3" json:"version,omitempty"`
	// rawId
	RawId string `protobuf:"bytes,3,opt,name=rawId,proto3" json:"rawId,omitempty"`
	// StoreId
	StoreId string `protobuf:"bytes,4,opt,name=storeId,proto3" json:"storeId,omitempty"`
	// AllowedDomains
	AllowedDomains []string `protobuf:"bytes,6,rep,name=allowedDomains,proto3" json:"allowedDomains,omitempty"`
	// ServeAddr
	ServeAddr string `protobuf:"bytes,11,opt,name=serveAddr,proto3" json:"serveAddr,omitempty"`
	// Status
	HostStatus []*Crawler_Status `protobuf:"bytes,12,rep,name=hostStatus,proto3" json:"hostStatus,omitempty"`
	// OnlineUtc
	OnlineUtc int64 `protobuf:"varint,16,opt,name=onlineUtc,proto3" json:"onlineUtc,omitempty"`
}

func (m *Crawler) Reset()         { *m = Crawler{} }
func (m *Crawler) String() string { return proto.CompactTextString(m) }
func (*Crawler) ProtoMessage()    {}
func (*Crawler) Descriptor() ([]byte, []int) {
	return fileDescriptor_84c7eabcfe7807d1, []int{0}
}
func (m *Crawler) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Crawler) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Crawler.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Crawler) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Crawler.Merge(m, src)
}
func (m *Crawler) XXX_Size() int {
	return m.Size()
}
func (m *Crawler) XXX_DiscardUnknown() {
	xxx_messageInfo_Crawler.DiscardUnknown(m)
}

var xxx_messageInfo_Crawler proto.InternalMessageInfo

func (m *Crawler) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *Crawler) GetVersion() int32 {
	if m != nil {
		return m.Version
	}
	return 0
}

func (m *Crawler) GetRawId() string {
	if m != nil {
		return m.RawId
	}
	return ""
}

func (m *Crawler) GetStoreId() string {
	if m != nil {
		return m.StoreId
	}
	return ""
}

func (m *Crawler) GetAllowedDomains() []string {
	if m != nil {
		return m.AllowedDomains
	}
	return nil
}

func (m *Crawler) GetServeAddr() string {
	if m != nil {
		return m.ServeAddr
	}
	return ""
}

func (m *Crawler) GetHostStatus() []*Crawler_Status {
	if m != nil {
		return m.HostStatus
	}
	return nil
}

func (m *Crawler) GetOnlineUtc() int64 {
	if m != nil {
		return m.OnlineUtc
	}
	return 0
}

// Status
type Crawler_Status struct {
	// Hostname
	Hostname string `protobuf:"bytes,1,opt,name=hostname,proto3" json:"hostname,omitempty"`
	// MaxAPIConcurrency
	MaxAPIConcurrency int32 `protobuf:"varint,3,opt,name=maxAPIConcurrency,proto3" json:"maxAPIConcurrency,omitempty"`
	// MaxMQConcurrency
	MaxMQConcurrency int32 `protobuf:"varint,4,opt,name=maxMQConcurrency,proto3" json:"maxMQConcurrency,omitempty"`
	// CurrentConcurrency
	CurrentConcurrency int32 `protobuf:"varint,5,opt,name=currentConcurrency,proto3" json:"currentConcurrency,omitempty"`
	// CurrentMQConcurrency
	CurrentMQConcurrency int32 `protobuf:"varint,6,opt,name=currentMQConcurrency,proto3" json:"currentMQConcurrency,omitempty"`
}

func (m *Crawler_Status) Reset()         { *m = Crawler_Status{} }
func (m *Crawler_Status) String() string { return proto.CompactTextString(m) }
func (*Crawler_Status) ProtoMessage()    {}
func (*Crawler_Status) Descriptor() ([]byte, []int) {
	return fileDescriptor_84c7eabcfe7807d1, []int{0, 0}
}
func (m *Crawler_Status) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Crawler_Status) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Crawler_Status.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Crawler_Status) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Crawler_Status.Merge(m, src)
}
func (m *Crawler_Status) XXX_Size() int {
	return m.Size()
}
func (m *Crawler_Status) XXX_DiscardUnknown() {
	xxx_messageInfo_Crawler_Status.DiscardUnknown(m)
}

var xxx_messageInfo_Crawler_Status proto.InternalMessageInfo

func (m *Crawler_Status) GetHostname() string {
	if m != nil {
		return m.Hostname
	}
	return ""
}

func (m *Crawler_Status) GetMaxAPIConcurrency() int32 {
	if m != nil {
		return m.MaxAPIConcurrency
	}
	return 0
}

func (m *Crawler_Status) GetMaxMQConcurrency() int32 {
	if m != nil {
		return m.MaxMQConcurrency
	}
	return 0
}

func (m *Crawler_Status) GetCurrentConcurrency() int32 {
	if m != nil {
		return m.CurrentConcurrency
	}
	return 0
}

func (m *Crawler_Status) GetCurrentMQConcurrency() int32 {
	if m != nil {
		return m.CurrentMQConcurrency
	}
	return 0
}

func init() {
	proto.RegisterType((*Crawler)(nil), "voiladev.voilacrawl.pkg.types.Crawler")
	proto.RegisterType((*Crawler_Status)(nil), "voiladev.voilacrawl.pkg.types.Crawler.Status")
}

func init() { proto.RegisterFile("crawler.proto", fileDescriptor_84c7eabcfe7807d1) }

var fileDescriptor_84c7eabcfe7807d1 = []byte{
	// 383 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x7c, 0x52, 0xcf, 0x8a, 0xda, 0x40,
	0x1c, 0x76, 0x8c, 0x89, 0x75, 0x6c, 0xc5, 0x0e, 0x1e, 0x06, 0x69, 0x43, 0x68, 0xa1, 0x84, 0x52,
	0x47, 0xb0, 0x4f, 0x60, 0xed, 0xc5, 0x83, 0xd0, 0xa6, 0xf4, 0xd2, 0xdb, 0x98, 0x99, 0xc6, 0xd0,
	0x24, 0x23, 0x93, 0x49, 0xd4, 0xb7, 0xe8, 0x63, 0xed, 0xd1, 0xe3, 0x1e, 0x77, 0xe3, 0x4b, 0x2c,
	0xec, 0x65, 0x71, 0x12, 0x57, 0x77, 0x95, 0xbd, 0xfd, 0xbe, 0xbf, 0x21, 0x1f, 0x03, 0xdf, 0xf8,
	0x92, 0xae, 0x22, 0x2e, 0xc9, 0x52, 0x0a, 0x25, 0xd0, 0xfb, 0x5c, 0x84, 0x11, 0x65, 0x3c, 0x27,
	0xfa, 0xd0, 0x22, 0x59, 0xfe, 0x0b, 0x88, 0xda, 0x2c, 0x79, 0xda, 0x1f, 0x04, 0xa1, 0x5a, 0x64,
	0x73, 0xe2, 0x8b, 0x78, 0x18, 0x88, 0x40, 0x0c, 0x75, 0x6a, 0x9e, 0xfd, 0xd5, 0x48, 0x03, 0x7d,
	0x95, 0x6d, 0x1f, 0xee, 0x0d, 0xd8, 0x9c, 0x94, 0xfd, 0xa8, 0x03, 0xeb, 0x21, 0xc3, 0xc0, 0x01,
	0x6e, 0xcb, 0xab, 0x87, 0x0c, 0x61, 0xd8, 0xcc, 0xb9, 0x4c, 0x43, 0x91, 0xe0, 0xba, 0x03, 0x5c,
	0xd3, 0x3b, 0x40, 0xd4, 0x83, 0xa6, 0xa4, 0xab, 0x29, 0xc3, 0x86, 0x36, 0x97, 0x60, 0xef, 0x4f,
	0x95, 0x90, 0x7c, 0xca, 0x70, 0x43, 0xf3, 0x07, 0x88, 0x3e, 0xc1, 0x0e, 0x8d, 0x22, 0xb1, 0xe2,
	0xec, 0xbb, 0x88, 0x69, 0x98, 0xa4, 0xd8, 0x72, 0x0c, 0xb7, 0xe5, 0x3d, 0x63, 0xd1, 0x3b, 0xd8,
	0x4a, 0xb9, 0xcc, 0xf9, 0x98, 0x31, 0x89, 0xdb, 0xba, 0xe3, 0x48, 0xa0, 0x19, 0x84, 0x0b, 0x91,
	0xaa, 0x5f, 0x8a, 0xaa, 0x2c, 0xc5, 0xaf, 0x1d, 0xc3, 0x6d, 0x8f, 0x06, 0xe4, 0xc5, 0x39, 0x48,
	0xf5, 0x6f, 0xa4, 0x0c, 0x79, 0x27, 0x05, 0xfb, 0x8f, 0x89, 0x24, 0x0a, 0x13, 0xfe, 0x5b, 0xf9,
	0xb8, 0xeb, 0x00, 0xd7, 0xf0, 0x8e, 0x44, 0xbf, 0x00, 0xd0, 0xaa, 0x8c, 0x7d, 0xf8, 0x6a, 0x1f,
	0x4b, 0x68, 0xcc, 0xab, 0x75, 0x1e, 0x31, 0xfa, 0x02, 0xdf, 0xc6, 0x74, 0x3d, 0xfe, 0x31, 0x9d,
	0x88, 0xc4, 0xcf, 0xa4, 0xe4, 0x89, 0xbf, 0xd1, 0xab, 0x98, 0xde, 0xb9, 0x80, 0x3e, 0xc3, 0x6e,
	0x4c, 0xd7, 0xb3, 0x9f, 0xa7, 0xe6, 0x86, 0x36, 0x9f, 0xf1, 0x88, 0x40, 0x54, 0xde, 0xea, 0xd4,
	0x6d, 0x6a, 0xf7, 0x05, 0x05, 0x8d, 0x60, 0xaf, 0x62, 0x9f, 0xf6, 0x5b, 0x3a, 0x71, 0x51, 0xfb,
	0xf6, 0xf1, 0xee, 0xd6, 0x06, 0x57, 0x85, 0x0d, 0xb6, 0x85, 0x0d, 0x6e, 0x0a, 0x1b, 0xfc, 0xdf,
	0xd9, 0xb5, 0xed, 0xce, 0xae, 0x5d, 0xef, 0xec, 0xda, 0x1f, 0x53, 0x4f, 0x38, 0xb7, 0xf4, 0x4b,
	0xf9, 0xfa, 0x10, 0x00, 0x00, 0xff, 0xff, 0x35, 0xbb, 0x82, 0x24, 0x88, 0x02, 0x00, 0x00,
}

func (m *Crawler) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Crawler) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Crawler) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.OnlineUtc != 0 {
		i = encodeVarintCrawler(dAtA, i, uint64(m.OnlineUtc))
		i--
		dAtA[i] = 0x1
		i--
		dAtA[i] = 0x80
	}
	if len(m.HostStatus) > 0 {
		for iNdEx := len(m.HostStatus) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.HostStatus[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintCrawler(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x62
		}
	}
	if len(m.ServeAddr) > 0 {
		i -= len(m.ServeAddr)
		copy(dAtA[i:], m.ServeAddr)
		i = encodeVarintCrawler(dAtA, i, uint64(len(m.ServeAddr)))
		i--
		dAtA[i] = 0x5a
	}
	if len(m.AllowedDomains) > 0 {
		for iNdEx := len(m.AllowedDomains) - 1; iNdEx >= 0; iNdEx-- {
			i -= len(m.AllowedDomains[iNdEx])
			copy(dAtA[i:], m.AllowedDomains[iNdEx])
			i = encodeVarintCrawler(dAtA, i, uint64(len(m.AllowedDomains[iNdEx])))
			i--
			dAtA[i] = 0x32
		}
	}
	if len(m.StoreId) > 0 {
		i -= len(m.StoreId)
		copy(dAtA[i:], m.StoreId)
		i = encodeVarintCrawler(dAtA, i, uint64(len(m.StoreId)))
		i--
		dAtA[i] = 0x22
	}
	if len(m.RawId) > 0 {
		i -= len(m.RawId)
		copy(dAtA[i:], m.RawId)
		i = encodeVarintCrawler(dAtA, i, uint64(len(m.RawId)))
		i--
		dAtA[i] = 0x1a
	}
	if m.Version != 0 {
		i = encodeVarintCrawler(dAtA, i, uint64(m.Version))
		i--
		dAtA[i] = 0x10
	}
	if len(m.Id) > 0 {
		i -= len(m.Id)
		copy(dAtA[i:], m.Id)
		i = encodeVarintCrawler(dAtA, i, uint64(len(m.Id)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *Crawler_Status) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Crawler_Status) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Crawler_Status) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.CurrentMQConcurrency != 0 {
		i = encodeVarintCrawler(dAtA, i, uint64(m.CurrentMQConcurrency))
		i--
		dAtA[i] = 0x30
	}
	if m.CurrentConcurrency != 0 {
		i = encodeVarintCrawler(dAtA, i, uint64(m.CurrentConcurrency))
		i--
		dAtA[i] = 0x28
	}
	if m.MaxMQConcurrency != 0 {
		i = encodeVarintCrawler(dAtA, i, uint64(m.MaxMQConcurrency))
		i--
		dAtA[i] = 0x20
	}
	if m.MaxAPIConcurrency != 0 {
		i = encodeVarintCrawler(dAtA, i, uint64(m.MaxAPIConcurrency))
		i--
		dAtA[i] = 0x18
	}
	if len(m.Hostname) > 0 {
		i -= len(m.Hostname)
		copy(dAtA[i:], m.Hostname)
		i = encodeVarintCrawler(dAtA, i, uint64(len(m.Hostname)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func encodeVarintCrawler(dAtA []byte, offset int, v uint64) int {
	offset -= sovCrawler(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func NewPopulatedCrawler(r randyCrawler, easy bool) *Crawler {
	this := &Crawler{}
	this.Id = string(randStringCrawler(r))
	this.Version = int32(r.Int31())
	if r.Intn(2) == 0 {
		this.Version *= -1
	}
	this.RawId = string(randStringCrawler(r))
	this.StoreId = string(randStringCrawler(r))
	v1 := r.Intn(10)
	this.AllowedDomains = make([]string, v1)
	for i := 0; i < v1; i++ {
		this.AllowedDomains[i] = string(randStringCrawler(r))
	}
	this.ServeAddr = string(randStringCrawler(r))
	if r.Intn(5) != 0 {
		v2 := r.Intn(5)
		this.HostStatus = make([]*Crawler_Status, v2)
		for i := 0; i < v2; i++ {
			this.HostStatus[i] = NewPopulatedCrawler_Status(r, easy)
		}
	}
	this.OnlineUtc = int64(r.Int63())
	if r.Intn(2) == 0 {
		this.OnlineUtc *= -1
	}
	if !easy && r.Intn(10) != 0 {
	}
	return this
}

func NewPopulatedCrawler_Status(r randyCrawler, easy bool) *Crawler_Status {
	this := &Crawler_Status{}
	this.Hostname = string(randStringCrawler(r))
	this.MaxAPIConcurrency = int32(r.Int31())
	if r.Intn(2) == 0 {
		this.MaxAPIConcurrency *= -1
	}
	this.MaxMQConcurrency = int32(r.Int31())
	if r.Intn(2) == 0 {
		this.MaxMQConcurrency *= -1
	}
	this.CurrentConcurrency = int32(r.Int31())
	if r.Intn(2) == 0 {
		this.CurrentConcurrency *= -1
	}
	this.CurrentMQConcurrency = int32(r.Int31())
	if r.Intn(2) == 0 {
		this.CurrentMQConcurrency *= -1
	}
	if !easy && r.Intn(10) != 0 {
	}
	return this
}

type randyCrawler interface {
	Float32() float32
	Float64() float64
	Int63() int64
	Int31() int32
	Uint32() uint32
	Intn(n int) int
}

func randUTF8RuneCrawler(r randyCrawler) rune {
	ru := r.Intn(62)
	if ru < 10 {
		return rune(ru + 48)
	} else if ru < 36 {
		return rune(ru + 55)
	}
	return rune(ru + 61)
}
func randStringCrawler(r randyCrawler) string {
	v3 := r.Intn(100)
	tmps := make([]rune, v3)
	for i := 0; i < v3; i++ {
		tmps[i] = randUTF8RuneCrawler(r)
	}
	return string(tmps)
}
func randUnrecognizedCrawler(r randyCrawler, maxFieldNumber int) (dAtA []byte) {
	l := r.Intn(5)
	for i := 0; i < l; i++ {
		wire := r.Intn(4)
		if wire == 3 {
			wire = 5
		}
		fieldNumber := maxFieldNumber + r.Intn(100)
		dAtA = randFieldCrawler(dAtA, r, fieldNumber, wire)
	}
	return dAtA
}
func randFieldCrawler(dAtA []byte, r randyCrawler, fieldNumber int, wire int) []byte {
	key := uint32(fieldNumber)<<3 | uint32(wire)
	switch wire {
	case 0:
		dAtA = encodeVarintPopulateCrawler(dAtA, uint64(key))
		v4 := r.Int63()
		if r.Intn(2) == 0 {
			v4 *= -1
		}
		dAtA = encodeVarintPopulateCrawler(dAtA, uint64(v4))
	case 1:
		dAtA = encodeVarintPopulateCrawler(dAtA, uint64(key))
		dAtA = append(dAtA, byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)))
	case 2:
		dAtA = encodeVarintPopulateCrawler(dAtA, uint64(key))
		ll := r.Intn(100)
		dAtA = encodeVarintPopulateCrawler(dAtA, uint64(ll))
		for j := 0; j < ll; j++ {
			dAtA = append(dAtA, byte(r.Intn(256)))
		}
	default:
		dAtA = encodeVarintPopulateCrawler(dAtA, uint64(key))
		dAtA = append(dAtA, byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)))
	}
	return dAtA
}
func encodeVarintPopulateCrawler(dAtA []byte, v uint64) []byte {
	for v >= 1<<7 {
		dAtA = append(dAtA, uint8(uint64(v)&0x7f|0x80))
		v >>= 7
	}
	dAtA = append(dAtA, uint8(v))
	return dAtA
}
func (m *Crawler) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Id)
	if l > 0 {
		n += 1 + l + sovCrawler(uint64(l))
	}
	if m.Version != 0 {
		n += 1 + sovCrawler(uint64(m.Version))
	}
	l = len(m.RawId)
	if l > 0 {
		n += 1 + l + sovCrawler(uint64(l))
	}
	l = len(m.StoreId)
	if l > 0 {
		n += 1 + l + sovCrawler(uint64(l))
	}
	if len(m.AllowedDomains) > 0 {
		for _, s := range m.AllowedDomains {
			l = len(s)
			n += 1 + l + sovCrawler(uint64(l))
		}
	}
	l = len(m.ServeAddr)
	if l > 0 {
		n += 1 + l + sovCrawler(uint64(l))
	}
	if len(m.HostStatus) > 0 {
		for _, e := range m.HostStatus {
			l = e.Size()
			n += 1 + l + sovCrawler(uint64(l))
		}
	}
	if m.OnlineUtc != 0 {
		n += 2 + sovCrawler(uint64(m.OnlineUtc))
	}
	return n
}

func (m *Crawler_Status) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Hostname)
	if l > 0 {
		n += 1 + l + sovCrawler(uint64(l))
	}
	if m.MaxAPIConcurrency != 0 {
		n += 1 + sovCrawler(uint64(m.MaxAPIConcurrency))
	}
	if m.MaxMQConcurrency != 0 {
		n += 1 + sovCrawler(uint64(m.MaxMQConcurrency))
	}
	if m.CurrentConcurrency != 0 {
		n += 1 + sovCrawler(uint64(m.CurrentConcurrency))
	}
	if m.CurrentMQConcurrency != 0 {
		n += 1 + sovCrawler(uint64(m.CurrentMQConcurrency))
	}
	return n
}

func sovCrawler(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozCrawler(x uint64) (n int) {
	return sovCrawler(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *Crawler) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowCrawler
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: Crawler: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Crawler: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Id", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCrawler
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthCrawler
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthCrawler
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Id = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Version", wireType)
			}
			m.Version = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCrawler
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Version |= int32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field RawId", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCrawler
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthCrawler
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthCrawler
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.RawId = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field StoreId", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCrawler
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthCrawler
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthCrawler
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.StoreId = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 6:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field AllowedDomains", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCrawler
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthCrawler
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthCrawler
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.AllowedDomains = append(m.AllowedDomains, string(dAtA[iNdEx:postIndex]))
			iNdEx = postIndex
		case 11:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ServeAddr", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCrawler
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthCrawler
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthCrawler
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ServeAddr = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 12:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field HostStatus", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCrawler
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthCrawler
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthCrawler
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.HostStatus = append(m.HostStatus, &Crawler_Status{})
			if err := m.HostStatus[len(m.HostStatus)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 16:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field OnlineUtc", wireType)
			}
			m.OnlineUtc = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCrawler
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.OnlineUtc |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipCrawler(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthCrawler
			}
			if (iNdEx + skippy) < 0 {
				return ErrInvalidLengthCrawler
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *Crawler_Status) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowCrawler
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: Status: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Status: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Hostname", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCrawler
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthCrawler
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthCrawler
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Hostname = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field MaxAPIConcurrency", wireType)
			}
			m.MaxAPIConcurrency = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCrawler
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.MaxAPIConcurrency |= int32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field MaxMQConcurrency", wireType)
			}
			m.MaxMQConcurrency = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCrawler
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.MaxMQConcurrency |= int32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 5:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field CurrentConcurrency", wireType)
			}
			m.CurrentConcurrency = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCrawler
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.CurrentConcurrency |= int32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 6:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field CurrentMQConcurrency", wireType)
			}
			m.CurrentMQConcurrency = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCrawler
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.CurrentMQConcurrency |= int32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipCrawler(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthCrawler
			}
			if (iNdEx + skippy) < 0 {
				return ErrInvalidLengthCrawler
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipCrawler(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowCrawler
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowCrawler
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowCrawler
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if length < 0 {
				return 0, ErrInvalidLengthCrawler
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupCrawler
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthCrawler
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthCrawler        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowCrawler          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupCrawler = fmt.Errorf("proto: unexpected end of group")
)
