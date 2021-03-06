// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: sql/execinfra/stats.proto

package execinfra

import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"

import time "time"

import github_com_gogo_protobuf_types "github.com/gogo/protobuf/types"

import io "io"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf
var _ = time.Kitchen

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion2 // please upgrade the proto package

// InputStats represents the stats collected from an input.
type InputStats struct {
	// num_rows is the number of rows received from the input.
	NumRows int64 `protobuf:"varint,1,opt,name=num_rows,json=numRows,proto3" json:"num_rows,omitempty"`
	// Duration in nanoseconds of the cumulative time spent stalled.
	StallTime time.Duration `protobuf:"bytes,8,opt,name=stall_time,json=stallTime,proto3,stdduration" json:"stall_time"`
}

func (m *InputStats) Reset()         { *m = InputStats{} }
func (m *InputStats) String() string { return proto.CompactTextString(m) }
func (*InputStats) ProtoMessage()    {}
func (*InputStats) Descriptor() ([]byte, []int) {
	return fileDescriptor_stats_ec3b62ca502c9165, []int{0}
}
func (m *InputStats) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *InputStats) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	b = b[:cap(b)]
	n, err := m.MarshalTo(b)
	if err != nil {
		return nil, err
	}
	return b[:n], nil
}
func (dst *InputStats) XXX_Merge(src proto.Message) {
	xxx_messageInfo_InputStats.Merge(dst, src)
}
func (m *InputStats) XXX_Size() int {
	return m.Size()
}
func (m *InputStats) XXX_DiscardUnknown() {
	xxx_messageInfo_InputStats.DiscardUnknown(m)
}

var xxx_messageInfo_InputStats proto.InternalMessageInfo

// JoinReaderStats are the stats collected during a JoinReader run.
type JoinReaderStats struct {
	InputStats       InputStats `protobuf:"bytes,1,opt,name=input_stats,json=inputStats,proto3" json:"input_stats"`
	IndexLookupStats InputStats `protobuf:"bytes,2,opt,name=index_lookup_stats,json=indexLookupStats,proto3" json:"index_lookup_stats"`
}

func (m *JoinReaderStats) Reset()         { *m = JoinReaderStats{} }
func (m *JoinReaderStats) String() string { return proto.CompactTextString(m) }
func (*JoinReaderStats) ProtoMessage()    {}
func (*JoinReaderStats) Descriptor() ([]byte, []int) {
	return fileDescriptor_stats_ec3b62ca502c9165, []int{1}
}
func (m *JoinReaderStats) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *JoinReaderStats) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	b = b[:cap(b)]
	n, err := m.MarshalTo(b)
	if err != nil {
		return nil, err
	}
	return b[:n], nil
}
func (dst *JoinReaderStats) XXX_Merge(src proto.Message) {
	xxx_messageInfo_JoinReaderStats.Merge(dst, src)
}
func (m *JoinReaderStats) XXX_Size() int {
	return m.Size()
}
func (m *JoinReaderStats) XXX_DiscardUnknown() {
	xxx_messageInfo_JoinReaderStats.DiscardUnknown(m)
}

var xxx_messageInfo_JoinReaderStats proto.InternalMessageInfo

func init() {
	proto.RegisterType((*InputStats)(nil), "cockroach.sql.distsqlrun.InputStats")
	proto.RegisterType((*JoinReaderStats)(nil), "cockroach.sql.distsqlrun.JoinReaderStats")
}
func (m *InputStats) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *InputStats) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.NumRows != 0 {
		dAtA[i] = 0x8
		i++
		i = encodeVarintStats(dAtA, i, uint64(m.NumRows))
	}
	dAtA[i] = 0x42
	i++
	i = encodeVarintStats(dAtA, i, uint64(github_com_gogo_protobuf_types.SizeOfStdDuration(m.StallTime)))
	n1, err := github_com_gogo_protobuf_types.StdDurationMarshalTo(m.StallTime, dAtA[i:])
	if err != nil {
		return 0, err
	}
	i += n1
	return i, nil
}

func (m *JoinReaderStats) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *JoinReaderStats) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	dAtA[i] = 0xa
	i++
	i = encodeVarintStats(dAtA, i, uint64(m.InputStats.Size()))
	n2, err := m.InputStats.MarshalTo(dAtA[i:])
	if err != nil {
		return 0, err
	}
	i += n2
	dAtA[i] = 0x12
	i++
	i = encodeVarintStats(dAtA, i, uint64(m.IndexLookupStats.Size()))
	n3, err := m.IndexLookupStats.MarshalTo(dAtA[i:])
	if err != nil {
		return 0, err
	}
	i += n3
	return i, nil
}

func encodeVarintStats(dAtA []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return offset + 1
}
func (m *InputStats) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.NumRows != 0 {
		n += 1 + sovStats(uint64(m.NumRows))
	}
	l = github_com_gogo_protobuf_types.SizeOfStdDuration(m.StallTime)
	n += 1 + l + sovStats(uint64(l))
	return n
}

func (m *JoinReaderStats) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = m.InputStats.Size()
	n += 1 + l + sovStats(uint64(l))
	l = m.IndexLookupStats.Size()
	n += 1 + l + sovStats(uint64(l))
	return n
}

func sovStats(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func sozStats(x uint64) (n int) {
	return sovStats(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *InputStats) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowStats
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: InputStats: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: InputStats: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field NumRows", wireType)
			}
			m.NumRows = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowStats
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.NumRows |= (int64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 8:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field StallTime", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowStats
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthStats
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := github_com_gogo_protobuf_types.StdDurationUnmarshal(&m.StallTime, dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipStats(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthStats
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
func (m *JoinReaderStats) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowStats
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: JoinReaderStats: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: JoinReaderStats: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field InputStats", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowStats
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthStats
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.InputStats.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field IndexLookupStats", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowStats
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthStats
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.IndexLookupStats.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipStats(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthStats
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
func skipStats(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowStats
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
					return 0, ErrIntOverflowStats
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
			return iNdEx, nil
		case 1:
			iNdEx += 8
			return iNdEx, nil
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowStats
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
			iNdEx += length
			if length < 0 {
				return 0, ErrInvalidLengthStats
			}
			return iNdEx, nil
		case 3:
			for {
				var innerWire uint64
				var start int = iNdEx
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return 0, ErrIntOverflowStats
					}
					if iNdEx >= l {
						return 0, io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					innerWire |= (uint64(b) & 0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				innerWireType := int(innerWire & 0x7)
				if innerWireType == 4 {
					break
				}
				next, err := skipStats(dAtA[start:])
				if err != nil {
					return 0, err
				}
				iNdEx = start + next
			}
			return iNdEx, nil
		case 4:
			return iNdEx, nil
		case 5:
			iNdEx += 4
			return iNdEx, nil
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
	}
	panic("unreachable")
}

var (
	ErrInvalidLengthStats = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowStats   = fmt.Errorf("proto: integer overflow")
)

func init() { proto.RegisterFile("sql/execinfra/stats.proto", fileDescriptor_stats_ec3b62ca502c9165) }

var fileDescriptor_stats_ec3b62ca502c9165 = []byte{
	// 327 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x91, 0xbd, 0x6e, 0xea, 0x30,
	0x14, 0xc7, 0xe3, 0x0b, 0xba, 0x37, 0x98, 0xe1, 0xa2, 0xe8, 0x0e, 0xc0, 0x60, 0x10, 0xba, 0x03,
	0x52, 0x25, 0x47, 0xa2, 0x6f, 0x10, 0x75, 0x29, 0xed, 0x94, 0x76, 0xa8, 0xba, 0x44, 0x21, 0x31,
	0xa9, 0x45, 0xe2, 0x03, 0xfe, 0x10, 0x3c, 0x46, 0xc7, 0x3e, 0x4d, 0x67, 0x46, 0x46, 0xa6, 0x7e,
	0x84, 0x17, 0xa9, 0xe2, 0xd0, 0x32, 0x75, 0xe8, 0xe6, 0x8f, 0xe3, 0xdf, 0xef, 0x7f, 0x7c, 0x70,
	0x4f, 0xad, 0x72, 0x9f, 0x6d, 0x58, 0xc2, 0xc5, 0x5c, 0xc6, 0xbe, 0xd2, 0xb1, 0x56, 0x74, 0x29,
	0x41, 0x83, 0xd7, 0x4d, 0x20, 0x59, 0x48, 0x88, 0x93, 0x07, 0xaa, 0x56, 0x39, 0x4d, 0xb9, 0xd2,
	0x6a, 0x95, 0x4b, 0x23, 0xfa, 0xff, 0x32, 0xc8, 0xc0, 0x16, 0xf9, 0xd5, 0xaa, 0xae, 0xef, 0x93,
	0x0c, 0x20, 0xcb, 0x99, 0x6f, 0x77, 0x33, 0x33, 0xf7, 0x53, 0x23, 0x63, 0xcd, 0x41, 0xd4, 0xf7,
	0xa3, 0x05, 0xc6, 0x97, 0x62, 0x69, 0xf4, 0x4d, 0xe5, 0xf0, 0x7a, 0xd8, 0x15, 0xa6, 0x88, 0x24,
	0xac, 0x55, 0x17, 0x0d, 0xd1, 0xb8, 0x11, 0xfe, 0x11, 0xa6, 0x08, 0x61, 0xad, 0xbc, 0x00, 0x63,
	0xa5, 0xe3, 0x3c, 0x8f, 0x34, 0x2f, 0x58, 0xd7, 0x1d, 0xa2, 0x71, 0x7b, 0xd2, 0xa3, 0x35, 0x9d,
	0x7e, 0xd2, 0xe9, 0xc5, 0x91, 0x1e, 0xb8, 0xdb, 0x97, 0x81, 0xf3, 0xf4, 0x3a, 0x40, 0x61, 0xcb,
	0x3e, 0xbb, 0xe5, 0x05, 0x1b, 0x3d, 0x23, 0xfc, 0x77, 0x0a, 0x5c, 0x84, 0x2c, 0x4e, 0x99, 0xac,
	0x95, 0x57, 0xb8, 0xcd, 0xab, 0x00, 0x91, 0xed, 0xd2, 0x5a, 0xdb, 0x93, 0xff, 0xf4, 0xbb, 0x36,
	0xe9, 0x29, 0x6d, 0xd0, 0xac, 0x1c, 0x21, 0xe6, 0xa7, 0xfc, 0x77, 0xd8, 0xe3, 0x22, 0x65, 0x9b,
	0x28, 0x07, 0x58, 0x98, 0xe5, 0x91, 0xf9, 0xeb, 0xc7, 0xcc, 0x8e, 0xa5, 0x5c, 0x5b, 0x88, 0x3d,
	0x9f, 0x36, 0xdd, 0x46, 0xa7, 0x19, 0x9c, 0x6d, 0xdf, 0x89, 0xb3, 0x2d, 0x09, 0xda, 0x95, 0x04,
	0xed, 0x4b, 0x82, 0xde, 0x4a, 0x82, 0x1e, 0x0f, 0xc4, 0xd9, 0x1d, 0x88, 0xb3, 0x3f, 0x10, 0xe7,
	0xbe, 0xf5, 0x35, 0xb6, 0xd9, 0x6f, 0xfb, 0x2b, 0xe7, 0x1f, 0x01, 0x00, 0x00, 0xff, 0xff, 0x45,
	0xf8, 0xdb, 0x61, 0xce, 0x01, 0x00, 0x00,
}
