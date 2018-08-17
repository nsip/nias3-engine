// Code generated by protoc-gen-go. DO NOT EDIT.
// source: block.proto

package n3

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

// core message type for lowest-level data tuples
type SPOTuple struct {
	Subject              string   `protobuf:"bytes,1,opt,name=Subject,proto3" json:"Subject,omitempty"`
	Object               string   `protobuf:"bytes,2,opt,name=Object,proto3" json:"Object,omitempty"`
	Predicate            string   `protobuf:"bytes,3,opt,name=Predicate,proto3" json:"Predicate,omitempty"`
	PredicateFlat        string   `protobuf:"bytes,4,opt,name=PredicateFlat,proto3" json:"PredicateFlat,omitempty"`
	Version              uint64   `protobuf:"varint,5,opt,name=Version,proto3" json:"Version,omitempty"`
	Context              string   `protobuf:"bytes,6,opt,name=Context,proto3" json:"Context,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *SPOTuple) Reset()         { *m = SPOTuple{} }
func (m *SPOTuple) String() string { return proto.CompactTextString(m) }
func (*SPOTuple) ProtoMessage()    {}
func (*SPOTuple) Descriptor() ([]byte, []int) {
	return fileDescriptor_block_50b4836eeb1da4b9, []int{0}
}
func (m *SPOTuple) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SPOTuple.Unmarshal(m, b)
}
func (m *SPOTuple) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SPOTuple.Marshal(b, m, deterministic)
}
func (dst *SPOTuple) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SPOTuple.Merge(dst, src)
}
func (m *SPOTuple) XXX_Size() int {
	return xxx_messageInfo_SPOTuple.Size(m)
}
func (m *SPOTuple) XXX_DiscardUnknown() {
	xxx_messageInfo_SPOTuple.DiscardUnknown(m)
}

var xxx_messageInfo_SPOTuple proto.InternalMessageInfo

func (m *SPOTuple) GetSubject() string {
	if m != nil {
		return m.Subject
	}
	return ""
}

func (m *SPOTuple) GetObject() string {
	if m != nil {
		return m.Object
	}
	return ""
}

func (m *SPOTuple) GetPredicate() string {
	if m != nil {
		return m.Predicate
	}
	return ""
}

func (m *SPOTuple) GetPredicateFlat() string {
	if m != nil {
		return m.PredicateFlat
	}
	return ""
}

func (m *SPOTuple) GetVersion() uint64 {
	if m != nil {
		return m.Version
	}
	return 0
}

func (m *SPOTuple) GetContext() string {
	if m != nil {
		return m.Context
	}
	return ""
}

type Block struct {
	BlockId              string    `protobuf:"bytes,1,opt,name=BlockId,proto3" json:"BlockId,omitempty"`
	Data                 *SPOTuple `protobuf:"bytes,2,opt,name=Data,proto3" json:"Data,omitempty"`
	PrevBlockHash        string    `protobuf:"bytes,3,opt,name=PrevBlockHash,proto3" json:"PrevBlockHash,omitempty"`
	Hash                 string    `protobuf:"bytes,4,opt,name=Hash,proto3" json:"Hash,omitempty"`
	Sig                  string    `protobuf:"bytes,5,opt,name=Sig,proto3" json:"Sig,omitempty"`
	Author               string    `protobuf:"bytes,6,opt,name=Author,proto3" json:"Author,omitempty"`
	Sender               string    `protobuf:"bytes,7,opt,name=Sender,proto3" json:"Sender,omitempty"`
	Receiver             string    `protobuf:"bytes,8,opt,name=Receiver,proto3" json:"Receiver,omitempty"`
	XXX_NoUnkeyedLiteral struct{}  `json:"-"`
	XXX_unrecognized     []byte    `json:"-"`
	XXX_sizecache        int32     `json:"-"`
}

func (m *Block) Reset()         { *m = Block{} }
func (m *Block) String() string { return proto.CompactTextString(m) }
func (*Block) ProtoMessage()    {}
func (*Block) Descriptor() ([]byte, []int) {
	return fileDescriptor_block_50b4836eeb1da4b9, []int{1}
}
func (m *Block) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Block.Unmarshal(m, b)
}
func (m *Block) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Block.Marshal(b, m, deterministic)
}
func (dst *Block) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Block.Merge(dst, src)
}
func (m *Block) XXX_Size() int {
	return xxx_messageInfo_Block.Size(m)
}
func (m *Block) XXX_DiscardUnknown() {
	xxx_messageInfo_Block.DiscardUnknown(m)
}

var xxx_messageInfo_Block proto.InternalMessageInfo

func (m *Block) GetBlockId() string {
	if m != nil {
		return m.BlockId
	}
	return ""
}

func (m *Block) GetData() *SPOTuple {
	if m != nil {
		return m.Data
	}
	return nil
}

func (m *Block) GetPrevBlockHash() string {
	if m != nil {
		return m.PrevBlockHash
	}
	return ""
}

func (m *Block) GetHash() string {
	if m != nil {
		return m.Hash
	}
	return ""
}

func (m *Block) GetSig() string {
	if m != nil {
		return m.Sig
	}
	return ""
}

func (m *Block) GetAuthor() string {
	if m != nil {
		return m.Author
	}
	return ""
}

func (m *Block) GetSender() string {
	if m != nil {
		return m.Sender
	}
	return ""
}

func (m *Block) GetReceiver() string {
	if m != nil {
		return m.Receiver
	}
	return ""
}

type DbCommand struct {
	Verb                 string    `protobuf:"bytes,1,opt,name=Verb,proto3" json:"Verb,omitempty"`
	Data                 *SPOTuple `protobuf:"bytes,2,opt,name=Data,proto3" json:"Data,omitempty"`
	Sequence             uint64    `protobuf:"varint,3,opt,name=Sequence,proto3" json:"Sequence,omitempty"`
	XXX_NoUnkeyedLiteral struct{}  `json:"-"`
	XXX_unrecognized     []byte    `json:"-"`
	XXX_sizecache        int32     `json:"-"`
}

func (m *DbCommand) Reset()         { *m = DbCommand{} }
func (m *DbCommand) String() string { return proto.CompactTextString(m) }
func (*DbCommand) ProtoMessage()    {}
func (*DbCommand) Descriptor() ([]byte, []int) {
	return fileDescriptor_block_50b4836eeb1da4b9, []int{2}
}
func (m *DbCommand) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_DbCommand.Unmarshal(m, b)
}
func (m *DbCommand) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_DbCommand.Marshal(b, m, deterministic)
}
func (dst *DbCommand) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DbCommand.Merge(dst, src)
}
func (m *DbCommand) XXX_Size() int {
	return xxx_messageInfo_DbCommand.Size(m)
}
func (m *DbCommand) XXX_DiscardUnknown() {
	xxx_messageInfo_DbCommand.DiscardUnknown(m)
}

var xxx_messageInfo_DbCommand proto.InternalMessageInfo

func (m *DbCommand) GetVerb() string {
	if m != nil {
		return m.Verb
	}
	return ""
}

func (m *DbCommand) GetData() *SPOTuple {
	if m != nil {
		return m.Data
	}
	return nil
}

func (m *DbCommand) GetSequence() uint64 {
	if m != nil {
		return m.Sequence
	}
	return 0
}

func init() {
	proto.RegisterType((*SPOTuple)(nil), "n3.SPOTuple")
	proto.RegisterType((*Block)(nil), "n3.Block")
	proto.RegisterType((*DbCommand)(nil), "n3.DbCommand")
}

func init() { proto.RegisterFile("block.proto", fileDescriptor_block_50b4836eeb1da4b9) }

var fileDescriptor_block_50b4836eeb1da4b9 = []byte{
	// 304 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x91, 0x41, 0x4f, 0xc2, 0x30,
	0x1c, 0xc5, 0x33, 0x18, 0xb0, 0xfd, 0xd1, 0xc4, 0xf4, 0x60, 0x1a, 0xe2, 0x81, 0x10, 0x0f, 0x9c,
	0x38, 0xc8, 0x27, 0x50, 0x88, 0xd1, 0x13, 0xa4, 0x33, 0xdc, 0x3c, 0x74, 0xdb, 0x3f, 0x32, 0x85,
	0x16, 0x4b, 0x47, 0xfc, 0x60, 0x7e, 0x1b, 0xbf, 0x8c, 0xe9, 0xbf, 0xdd, 0x0c, 0x37, 0x6f, 0xef,
	0xf7, 0x5e, 0x97, 0xbe, 0xb7, 0xc2, 0x30, 0xdf, 0xe9, 0xe2, 0x63, 0x76, 0x30, 0xda, 0x6a, 0xd6,
	0x51, 0xf3, 0xc9, 0x77, 0x04, 0x49, 0xb6, 0x5e, 0xbd, 0xd4, 0x87, 0x1d, 0x32, 0x0e, 0x83, 0xac,
	0xce, 0xdf, 0xb1, 0xb0, 0x3c, 0x1a, 0x47, 0xd3, 0x54, 0x34, 0xc8, 0xae, 0xa1, 0xbf, 0xf2, 0x41,
	0x87, 0x82, 0x40, 0xec, 0x06, 0xd2, 0xb5, 0xc1, 0xb2, 0x2a, 0xa4, 0x45, 0xde, 0xa5, 0xe8, 0xcf,
	0x60, 0xb7, 0x70, 0xd9, 0xc2, 0xe3, 0x4e, 0x5a, 0x1e, 0xd3, 0x89, 0x73, 0xd3, 0xdd, 0xba, 0x41,
	0x73, 0xac, 0xb4, 0xe2, 0xbd, 0x71, 0x34, 0x8d, 0x45, 0x83, 0x2e, 0x59, 0x68, 0x65, 0xf1, 0xcb,
	0xf2, 0xbe, 0xef, 0x13, 0x70, 0xf2, 0x13, 0x41, 0xef, 0xc1, 0x4d, 0x71, 0x67, 0x48, 0x3c, 0x97,
	0x4d, 0xe7, 0x80, 0x6c, 0x0c, 0xf1, 0x52, 0x5a, 0x49, 0x8d, 0x87, 0x77, 0x17, 0x33, 0x35, 0x9f,
	0x35, 0x4b, 0x05, 0x25, 0xa1, 0xdf, 0x89, 0x3e, 0x78, 0x92, 0xc7, 0x6d, 0x58, 0x70, 0x6e, 0x32,
	0x06, 0x31, 0x85, 0xbe, 0x3c, 0x69, 0x76, 0x05, 0xdd, 0xac, 0x7a, 0xa3, 0xbe, 0xa9, 0x70, 0xd2,
	0xfd, 0xa1, 0xfb, 0xda, 0x6e, 0xb5, 0x09, 0x55, 0x03, 0x39, 0x3f, 0x43, 0x55, 0xa2, 0xe1, 0x03,
	0xef, 0x7b, 0x62, 0x23, 0x48, 0x04, 0x16, 0x58, 0x9d, 0xd0, 0xf0, 0x84, 0x92, 0x96, 0x27, 0xaf,
	0x90, 0x2e, 0xf3, 0x85, 0xde, 0xef, 0xa5, 0x2a, 0xdd, 0xf5, 0x1b, 0x34, 0x79, 0x58, 0x47, 0xfa,
	0x1f, 0xd3, 0x46, 0x90, 0x64, 0xf8, 0x59, 0xa3, 0x2a, 0xfc, 0xbb, 0xc4, 0xa2, 0xe5, 0xbc, 0x4f,
	0xcf, 0x3f, 0xff, 0x0d, 0x00, 0x00, 0xff, 0xff, 0x63, 0x74, 0xe3, 0x77, 0x0d, 0x02, 0x00, 0x00,
}
