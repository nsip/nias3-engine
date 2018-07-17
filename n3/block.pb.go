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
	Version              uint64   `protobuf:"varint,4,opt,name=Version,proto3" json:"Version,omitempty"`
	Context              string   `protobuf:"bytes,5,opt,name=Context,proto3" json:"Context,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *SPOTuple) Reset()         { *m = SPOTuple{} }
func (m *SPOTuple) String() string { return proto.CompactTextString(m) }
func (*SPOTuple) ProtoMessage()    {}
func (*SPOTuple) Descriptor() ([]byte, []int) {
	return fileDescriptor_block_d22b36f5a6932c14, []int{0}
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
	return fileDescriptor_block_d22b36f5a6932c14, []int{1}
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
	Verb                 string   `protobuf:"bytes,1,opt,name=verb,proto3" json:"verb,omitempty"`
	Key                  []byte   `protobuf:"bytes,2,opt,name=key,proto3" json:"key,omitempty"`
	Value                []byte   `protobuf:"bytes,3,opt,name=value,proto3" json:"value,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *DbCommand) Reset()         { *m = DbCommand{} }
func (m *DbCommand) String() string { return proto.CompactTextString(m) }
func (*DbCommand) ProtoMessage()    {}
func (*DbCommand) Descriptor() ([]byte, []int) {
	return fileDescriptor_block_d22b36f5a6932c14, []int{2}
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

func (m *DbCommand) GetKey() []byte {
	if m != nil {
		return m.Key
	}
	return nil
}

func (m *DbCommand) GetValue() []byte {
	if m != nil {
		return m.Value
	}
	return nil
}

func init() {
	proto.RegisterType((*SPOTuple)(nil), "n3.SPOTuple")
	proto.RegisterType((*Block)(nil), "n3.Block")
	proto.RegisterType((*DbCommand)(nil), "n3.DbCommand")
}

func init() { proto.RegisterFile("block.proto", fileDescriptor_block_d22b36f5a6932c14) }

var fileDescriptor_block_d22b36f5a6932c14 = []byte{
	// 300 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x54, 0x91, 0x4f, 0x4e, 0xf3, 0x30,
	0x10, 0xc5, 0x95, 0x36, 0xfd, 0x37, 0xed, 0x27, 0x7d, 0xb2, 0x10, 0xb2, 0x10, 0x8b, 0xaa, 0x62,
	0xd1, 0x55, 0x16, 0xe4, 0x04, 0xd0, 0x4a, 0xc0, 0xaa, 0x91, 0x83, 0xd8, 0x3b, 0xc9, 0x88, 0x86,
	0xa6, 0x76, 0xe5, 0x3a, 0x11, 0x5c, 0x82, 0x0b, 0x72, 0x19, 0xe4, 0xb1, 0x03, 0x62, 0xf7, 0x7e,
	0xf3, 0x12, 0xcf, 0x7b, 0x36, 0xcc, 0x8b, 0x46, 0x97, 0x87, 0xe4, 0x64, 0xb4, 0xd5, 0x6c, 0xa0,
	0xd2, 0xd5, 0x67, 0x04, 0xd3, 0x3c, 0xdb, 0x3d, 0xb7, 0xa7, 0x06, 0x19, 0x87, 0x49, 0xde, 0x16,
	0x6f, 0x58, 0x5a, 0x1e, 0x2d, 0xa3, 0xf5, 0x4c, 0xf4, 0xc8, 0x2e, 0x61, 0xbc, 0xf3, 0xc6, 0x80,
	0x8c, 0x40, 0xec, 0x1a, 0x66, 0x99, 0xc1, 0xaa, 0x2e, 0xa5, 0x45, 0x3e, 0x24, 0xeb, 0x77, 0xe0,
	0xce, 0x7b, 0x41, 0x73, 0xae, 0xb5, 0xe2, 0xf1, 0x32, 0x5a, 0xc7, 0xa2, 0x47, 0xe7, 0x6c, 0xb4,
	0xb2, 0xf8, 0x6e, 0xf9, 0xc8, 0x6f, 0x0a, 0xb8, 0xfa, 0x8a, 0x60, 0x74, 0xef, 0x42, 0xba, 0x6f,
	0x48, 0x3c, 0x55, 0x7d, 0x9a, 0x80, 0x6c, 0x09, 0xf1, 0x56, 0x5a, 0x49, 0x59, 0xe6, 0xb7, 0x8b,
	0x44, 0xa5, 0x49, 0xdf, 0x41, 0x90, 0xc3, 0x6e, 0xe0, 0x5f, 0x66, 0xb0, 0xa3, 0x1f, 0x1e, 0xe5,
	0x79, 0x1f, 0xb2, 0xfd, 0x1d, 0x32, 0x06, 0x31, 0x99, 0x31, 0x99, 0xa4, 0xd9, 0x7f, 0x18, 0xe6,
	0xf5, 0x6b, 0x48, 0xe5, 0xa4, 0xeb, 0x7e, 0xd7, 0xda, 0xbd, 0x36, 0x7c, 0xec, 0xbb, 0x7b, 0x72,
	0xf3, 0x1c, 0x55, 0x85, 0x86, 0x4f, 0xfc, 0xdc, 0x13, 0xbb, 0x82, 0xa9, 0xc0, 0x12, 0xeb, 0x0e,
	0x0d, 0x9f, 0x92, 0xf3, 0xc3, 0xab, 0x07, 0x98, 0x6d, 0x8b, 0x8d, 0x3e, 0x1e, 0xa5, 0xaa, 0xdc,
	0xfa, 0x0e, 0x4d, 0x11, 0xda, 0x91, 0x76, 0xeb, 0x0f, 0xf8, 0x41, 0xcd, 0x16, 0xc2, 0x49, 0x76,
	0x01, 0xa3, 0x4e, 0x36, 0xad, 0xbf, 0xde, 0x85, 0xf0, 0x50, 0x8c, 0xe9, 0x09, 0xd3, 0xef, 0x00,
	0x00, 0x00, 0xff, 0xff, 0x79, 0xf1, 0x02, 0x1b, 0xd1, 0x01, 0x00, 0x00,
}
