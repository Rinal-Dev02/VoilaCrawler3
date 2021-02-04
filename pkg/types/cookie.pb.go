// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: cookie.proto

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

// Cookie
type Cookie struct {
	// Id
	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	// TracingId
	TracingId string `protobuf:"bytes,2,opt,name=tracingId,proto3" json:"tracingId,omitempty"`
	// Name
	Name string `protobuf:"bytes,6,opt,name=name,proto3" json:"name,omitempty"`
	// Value
	Value string `protobuf:"bytes,7,opt,name=value,proto3" json:"value,omitempty"`
	// Domain
	Domain string `protobuf:"bytes,8,opt,name=domain,proto3" json:"domain,omitempty"`
	// Path
	Path string `protobuf:"bytes,9,opt,name=path,proto3" json:"path,omitempty"`
	// Expires
	Expires int64 `protobuf:"varint,10,opt,name=expires,proto3" json:"expires,omitempty"`
	// Size
	Size_ int32 `protobuf:"varint,13,opt,name=size,proto3" json:"size,omitempty"`
	// HttpOnly
	HttpOnly bool `protobuf:"varint,14,opt,name=httpOnly,proto3" json:"httpOnly,omitempty"`
	// Session
	Session bool `protobuf:"varint,15,opt,name=session,proto3" json:"session,omitempty"`
	// SameSite Strict,Lax,None
	SameSite string `protobuf:"bytes,16,opt,name=sameSite,proto3" json:"sameSite,omitempty"`
	// Priority Low, Medium, High
	Priority string `protobuf:"bytes,17,opt,name=priority,proto3" json:"priority,omitempty"`
	// CreatedUtc
	CreatedUtc int64 `protobuf:"varint,21,opt,name=createdUtc,proto3" json:"createdUtc,omitempty"`
}

func (m *Cookie) Reset()         { *m = Cookie{} }
func (m *Cookie) String() string { return proto.CompactTextString(m) }
func (*Cookie) ProtoMessage()    {}
func (*Cookie) Descriptor() ([]byte, []int) {
	return fileDescriptor_5dfef0e1712b7cf5, []int{0}
}
func (m *Cookie) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Cookie) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Cookie.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Cookie) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Cookie.Merge(m, src)
}
func (m *Cookie) XXX_Size() int {
	return m.Size()
}
func (m *Cookie) XXX_DiscardUnknown() {
	xxx_messageInfo_Cookie.DiscardUnknown(m)
}

var xxx_messageInfo_Cookie proto.InternalMessageInfo

func (m *Cookie) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *Cookie) GetTracingId() string {
	if m != nil {
		return m.TracingId
	}
	return ""
}

func (m *Cookie) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *Cookie) GetValue() string {
	if m != nil {
		return m.Value
	}
	return ""
}

func (m *Cookie) GetDomain() string {
	if m != nil {
		return m.Domain
	}
	return ""
}

func (m *Cookie) GetPath() string {
	if m != nil {
		return m.Path
	}
	return ""
}

func (m *Cookie) GetExpires() int64 {
	if m != nil {
		return m.Expires
	}
	return 0
}

func (m *Cookie) GetSize_() int32 {
	if m != nil {
		return m.Size_
	}
	return 0
}

func (m *Cookie) GetHttpOnly() bool {
	if m != nil {
		return m.HttpOnly
	}
	return false
}

func (m *Cookie) GetSession() bool {
	if m != nil {
		return m.Session
	}
	return false
}

func (m *Cookie) GetSameSite() string {
	if m != nil {
		return m.SameSite
	}
	return ""
}

func (m *Cookie) GetPriority() string {
	if m != nil {
		return m.Priority
	}
	return ""
}

func (m *Cookie) GetCreatedUtc() int64 {
	if m != nil {
		return m.CreatedUtc
	}
	return 0
}

func init() {
	proto.RegisterType((*Cookie)(nil), "voiladev.voilacrawl.pkg.types.Cookie")
}

func init() { proto.RegisterFile("cookie.proto", fileDescriptor_5dfef0e1712b7cf5) }

var fileDescriptor_5dfef0e1712b7cf5 = []byte{
	// 332 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x4c, 0x91, 0xb1, 0x4e, 0xc3, 0x30,
	0x14, 0x45, 0xeb, 0x40, 0xd3, 0xd6, 0x82, 0x02, 0x16, 0xa0, 0xa7, 0x0a, 0xac, 0x08, 0x96, 0x2c,
	0xa4, 0x03, 0x7f, 0x00, 0x13, 0x13, 0x52, 0x10, 0x0b, 0x9b, 0x9b, 0x98, 0xd4, 0x6a, 0x12, 0x47,
	0xb6, 0x5b, 0x28, 0x5f, 0xc1, 0x67, 0x75, 0xec, 0xc8, 0x08, 0xed, 0x4f, 0x30, 0xa2, 0xbc, 0xd0,
	0xc2, 0x76, 0xcf, 0xbd, 0x7e, 0xd7, 0x96, 0x1f, 0xdd, 0x4b, 0xb4, 0x9e, 0x28, 0x19, 0x55, 0x46,
	0x3b, 0xcd, 0xce, 0x67, 0x5a, 0xe5, 0x22, 0x95, 0xb3, 0x08, 0x45, 0x62, 0xc4, 0x4b, 0x1e, 0x55,
	0x93, 0x2c, 0x72, 0xf3, 0x4a, 0xda, 0xc1, 0x55, 0xa6, 0xdc, 0x78, 0x3a, 0x8a, 0x12, 0x5d, 0x0c,
	0x33, 0x9d, 0xe9, 0x21, 0x4e, 0x8d, 0xa6, 0xcf, 0x48, 0x08, 0xa8, 0x9a, 0xb6, 0x8b, 0x85, 0x47,
	0xfd, 0x5b, 0xac, 0x67, 0x7d, 0xea, 0xa9, 0x14, 0x48, 0x40, 0xc2, 0x5e, 0xec, 0xa9, 0x94, 0x9d,
	0xd1, 0x9e, 0x33, 0x22, 0x51, 0x65, 0x76, 0x97, 0x82, 0x87, 0xf6, 0x9f, 0xc1, 0x18, 0xdd, 0x2d,
	0x45, 0x21, 0xc1, 0xc7, 0x00, 0x35, 0x3b, 0xa6, 0xed, 0x99, 0xc8, 0xa7, 0x12, 0x3a, 0x68, 0x36,
	0xc0, 0x4e, 0xa9, 0x9f, 0xea, 0x42, 0xa8, 0x12, 0xba, 0x68, 0xff, 0x52, 0xdd, 0x50, 0x09, 0x37,
	0x86, 0x5e, 0xd3, 0x50, 0x6b, 0x06, 0xb4, 0x23, 0x5f, 0x2b, 0x65, 0xa4, 0x05, 0x1a, 0x90, 0x70,
	0x27, 0xde, 0x60, 0x7d, 0xda, 0xaa, 0x37, 0x09, 0xfb, 0x01, 0x09, 0xdb, 0x31, 0x6a, 0x36, 0xa0,
	0xdd, 0xb1, 0x73, 0xd5, 0x7d, 0x99, 0xcf, 0xa1, 0x1f, 0x90, 0xb0, 0x1b, 0x6f, 0xb9, 0x6e, 0xb2,
	0xd2, 0x5a, 0xa5, 0x4b, 0x38, 0xc0, 0x68, 0x83, 0xf5, 0x94, 0x15, 0x85, 0x7c, 0x50, 0x4e, 0xc2,
	0x21, 0xde, 0xbd, 0xe5, 0x3a, 0xab, 0x8c, 0xd2, 0x46, 0xb9, 0x39, 0x1c, 0x35, 0xd9, 0x86, 0x19,
	0xa7, 0x34, 0x31, 0x52, 0x38, 0x99, 0x3e, 0xba, 0x04, 0x4e, 0xf0, 0x79, 0xff, 0x9c, 0x9b, 0xcb,
	0xef, 0x2f, 0x4e, 0x16, 0x2b, 0x4e, 0x96, 0x2b, 0x4e, 0x3e, 0x57, 0x9c, 0xbc, 0xaf, 0x79, 0x6b,
	0xb9, 0xe6, 0xad, 0x8f, 0x35, 0x6f, 0x3d, 0xb5, 0x71, 0x3d, 0x23, 0x1f, 0xbf, 0xfd, 0xfa, 0x27,
	0x00, 0x00, 0xff, 0xff, 0xf4, 0x86, 0x72, 0x64, 0xd4, 0x01, 0x00, 0x00,
}

func (m *Cookie) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Cookie) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Cookie) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.CreatedUtc != 0 {
		i = encodeVarintCookie(dAtA, i, uint64(m.CreatedUtc))
		i--
		dAtA[i] = 0x1
		i--
		dAtA[i] = 0xa8
	}
	if len(m.Priority) > 0 {
		i -= len(m.Priority)
		copy(dAtA[i:], m.Priority)
		i = encodeVarintCookie(dAtA, i, uint64(len(m.Priority)))
		i--
		dAtA[i] = 0x1
		i--
		dAtA[i] = 0x8a
	}
	if len(m.SameSite) > 0 {
		i -= len(m.SameSite)
		copy(dAtA[i:], m.SameSite)
		i = encodeVarintCookie(dAtA, i, uint64(len(m.SameSite)))
		i--
		dAtA[i] = 0x1
		i--
		dAtA[i] = 0x82
	}
	if m.Session {
		i--
		if m.Session {
			dAtA[i] = 1
		} else {
			dAtA[i] = 0
		}
		i--
		dAtA[i] = 0x78
	}
	if m.HttpOnly {
		i--
		if m.HttpOnly {
			dAtA[i] = 1
		} else {
			dAtA[i] = 0
		}
		i--
		dAtA[i] = 0x70
	}
	if m.Size_ != 0 {
		i = encodeVarintCookie(dAtA, i, uint64(m.Size_))
		i--
		dAtA[i] = 0x68
	}
	if m.Expires != 0 {
		i = encodeVarintCookie(dAtA, i, uint64(m.Expires))
		i--
		dAtA[i] = 0x50
	}
	if len(m.Path) > 0 {
		i -= len(m.Path)
		copy(dAtA[i:], m.Path)
		i = encodeVarintCookie(dAtA, i, uint64(len(m.Path)))
		i--
		dAtA[i] = 0x4a
	}
	if len(m.Domain) > 0 {
		i -= len(m.Domain)
		copy(dAtA[i:], m.Domain)
		i = encodeVarintCookie(dAtA, i, uint64(len(m.Domain)))
		i--
		dAtA[i] = 0x42
	}
	if len(m.Value) > 0 {
		i -= len(m.Value)
		copy(dAtA[i:], m.Value)
		i = encodeVarintCookie(dAtA, i, uint64(len(m.Value)))
		i--
		dAtA[i] = 0x3a
	}
	if len(m.Name) > 0 {
		i -= len(m.Name)
		copy(dAtA[i:], m.Name)
		i = encodeVarintCookie(dAtA, i, uint64(len(m.Name)))
		i--
		dAtA[i] = 0x32
	}
	if len(m.TracingId) > 0 {
		i -= len(m.TracingId)
		copy(dAtA[i:], m.TracingId)
		i = encodeVarintCookie(dAtA, i, uint64(len(m.TracingId)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.Id) > 0 {
		i -= len(m.Id)
		copy(dAtA[i:], m.Id)
		i = encodeVarintCookie(dAtA, i, uint64(len(m.Id)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func encodeVarintCookie(dAtA []byte, offset int, v uint64) int {
	offset -= sovCookie(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func NewPopulatedCookie(r randyCookie, easy bool) *Cookie {
	this := &Cookie{}
	this.Id = string(randStringCookie(r))
	this.TracingId = string(randStringCookie(r))
	this.Name = string(randStringCookie(r))
	this.Value = string(randStringCookie(r))
	this.Domain = string(randStringCookie(r))
	this.Path = string(randStringCookie(r))
	this.Expires = int64(r.Int63())
	if r.Intn(2) == 0 {
		this.Expires *= -1
	}
	this.Size_ = int32(r.Int31())
	if r.Intn(2) == 0 {
		this.Size_ *= -1
	}
	this.HttpOnly = bool(bool(r.Intn(2) == 0))
	this.Session = bool(bool(r.Intn(2) == 0))
	this.SameSite = string(randStringCookie(r))
	this.Priority = string(randStringCookie(r))
	this.CreatedUtc = int64(r.Int63())
	if r.Intn(2) == 0 {
		this.CreatedUtc *= -1
	}
	if !easy && r.Intn(10) != 0 {
	}
	return this
}

type randyCookie interface {
	Float32() float32
	Float64() float64
	Int63() int64
	Int31() int32
	Uint32() uint32
	Intn(n int) int
}

func randUTF8RuneCookie(r randyCookie) rune {
	ru := r.Intn(62)
	if ru < 10 {
		return rune(ru + 48)
	} else if ru < 36 {
		return rune(ru + 55)
	}
	return rune(ru + 61)
}
func randStringCookie(r randyCookie) string {
	v1 := r.Intn(100)
	tmps := make([]rune, v1)
	for i := 0; i < v1; i++ {
		tmps[i] = randUTF8RuneCookie(r)
	}
	return string(tmps)
}
func randUnrecognizedCookie(r randyCookie, maxFieldNumber int) (dAtA []byte) {
	l := r.Intn(5)
	for i := 0; i < l; i++ {
		wire := r.Intn(4)
		if wire == 3 {
			wire = 5
		}
		fieldNumber := maxFieldNumber + r.Intn(100)
		dAtA = randFieldCookie(dAtA, r, fieldNumber, wire)
	}
	return dAtA
}
func randFieldCookie(dAtA []byte, r randyCookie, fieldNumber int, wire int) []byte {
	key := uint32(fieldNumber)<<3 | uint32(wire)
	switch wire {
	case 0:
		dAtA = encodeVarintPopulateCookie(dAtA, uint64(key))
		v2 := r.Int63()
		if r.Intn(2) == 0 {
			v2 *= -1
		}
		dAtA = encodeVarintPopulateCookie(dAtA, uint64(v2))
	case 1:
		dAtA = encodeVarintPopulateCookie(dAtA, uint64(key))
		dAtA = append(dAtA, byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)))
	case 2:
		dAtA = encodeVarintPopulateCookie(dAtA, uint64(key))
		ll := r.Intn(100)
		dAtA = encodeVarintPopulateCookie(dAtA, uint64(ll))
		for j := 0; j < ll; j++ {
			dAtA = append(dAtA, byte(r.Intn(256)))
		}
	default:
		dAtA = encodeVarintPopulateCookie(dAtA, uint64(key))
		dAtA = append(dAtA, byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)))
	}
	return dAtA
}
func encodeVarintPopulateCookie(dAtA []byte, v uint64) []byte {
	for v >= 1<<7 {
		dAtA = append(dAtA, uint8(uint64(v)&0x7f|0x80))
		v >>= 7
	}
	dAtA = append(dAtA, uint8(v))
	return dAtA
}
func (m *Cookie) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Id)
	if l > 0 {
		n += 1 + l + sovCookie(uint64(l))
	}
	l = len(m.TracingId)
	if l > 0 {
		n += 1 + l + sovCookie(uint64(l))
	}
	l = len(m.Name)
	if l > 0 {
		n += 1 + l + sovCookie(uint64(l))
	}
	l = len(m.Value)
	if l > 0 {
		n += 1 + l + sovCookie(uint64(l))
	}
	l = len(m.Domain)
	if l > 0 {
		n += 1 + l + sovCookie(uint64(l))
	}
	l = len(m.Path)
	if l > 0 {
		n += 1 + l + sovCookie(uint64(l))
	}
	if m.Expires != 0 {
		n += 1 + sovCookie(uint64(m.Expires))
	}
	if m.Size_ != 0 {
		n += 1 + sovCookie(uint64(m.Size_))
	}
	if m.HttpOnly {
		n += 2
	}
	if m.Session {
		n += 2
	}
	l = len(m.SameSite)
	if l > 0 {
		n += 2 + l + sovCookie(uint64(l))
	}
	l = len(m.Priority)
	if l > 0 {
		n += 2 + l + sovCookie(uint64(l))
	}
	if m.CreatedUtc != 0 {
		n += 2 + sovCookie(uint64(m.CreatedUtc))
	}
	return n
}

func sovCookie(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozCookie(x uint64) (n int) {
	return sovCookie(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *Cookie) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowCookie
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
			return fmt.Errorf("proto: Cookie: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Cookie: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Id", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCookie
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
				return ErrInvalidLengthCookie
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthCookie
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Id = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field TracingId", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCookie
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
				return ErrInvalidLengthCookie
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthCookie
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.TracingId = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 6:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Name", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCookie
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
				return ErrInvalidLengthCookie
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthCookie
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Name = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 7:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Value", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCookie
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
				return ErrInvalidLengthCookie
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthCookie
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Value = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 8:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Domain", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCookie
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
				return ErrInvalidLengthCookie
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthCookie
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Domain = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 9:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Path", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCookie
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
				return ErrInvalidLengthCookie
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthCookie
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Path = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 10:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Expires", wireType)
			}
			m.Expires = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCookie
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Expires |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 13:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Size_", wireType)
			}
			m.Size_ = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCookie
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Size_ |= int32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 14:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field HttpOnly", wireType)
			}
			var v int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCookie
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				v |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			m.HttpOnly = bool(v != 0)
		case 15:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Session", wireType)
			}
			var v int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCookie
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				v |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			m.Session = bool(v != 0)
		case 16:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field SameSite", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCookie
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
				return ErrInvalidLengthCookie
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthCookie
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.SameSite = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 17:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Priority", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCookie
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
				return ErrInvalidLengthCookie
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthCookie
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Priority = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 21:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field CreatedUtc", wireType)
			}
			m.CreatedUtc = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCookie
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.CreatedUtc |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipCookie(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthCookie
			}
			if (iNdEx + skippy) < 0 {
				return ErrInvalidLengthCookie
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
func skipCookie(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowCookie
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
					return 0, ErrIntOverflowCookie
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
					return 0, ErrIntOverflowCookie
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
				return 0, ErrInvalidLengthCookie
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupCookie
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthCookie
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthCookie        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowCookie          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupCookie = fmt.Errorf("proto: unexpected end of group")
)
