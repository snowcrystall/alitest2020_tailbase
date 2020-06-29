// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.24.0
// 	protoc        v3.12.3
// source: pb/service.proto

package pb

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

type TraceData struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Tracedata []byte `protobuf:"bytes,1,opt,name=tracedata,proto3" json:"tracedata,omitempty"`
}

func (x *TraceData) Reset() {
	*x = TraceData{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pb_service_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TraceData) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TraceData) ProtoMessage() {}

func (x *TraceData) ProtoReflect() protoreflect.Message {
	mi := &file_pb_service_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TraceData.ProtoReflect.Descriptor instead.
func (*TraceData) Descriptor() ([]byte, []int) {
	return file_pb_service_proto_rawDescGZIP(), []int{0}
}

func (x *TraceData) GetTracedata() []byte {
	if x != nil {
		return x.Tracedata
	}
	return nil
}

type TargetInfo struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Traceid  int64 `protobuf:"varint,1,opt,name=traceid,proto3" json:"traceid,omitempty"`
	Checkcur int64 `protobuf:"varint,2,opt,name=checkcur,proto3" json:"checkcur,omitempty"`
}

func (x *TargetInfo) Reset() {
	*x = TargetInfo{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pb_service_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TargetInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TargetInfo) ProtoMessage() {}

func (x *TargetInfo) ProtoReflect() protoreflect.Message {
	mi := &file_pb_service_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TargetInfo.ProtoReflect.Descriptor instead.
func (*TargetInfo) Descriptor() ([]byte, []int) {
	return file_pb_service_proto_rawDescGZIP(), []int{1}
}

func (x *TargetInfo) GetTraceid() int64 {
	if x != nil {
		return x.Traceid
	}
	return 0
}

func (x *TargetInfo) GetCheckcur() int64 {
	if x != nil {
		return x.Checkcur
	}
	return 0
}

type Reply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Reply []byte `protobuf:"bytes,1,opt,name=reply,proto3" json:"reply,omitempty"`
}

func (x *Reply) Reset() {
	*x = Reply{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pb_service_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Reply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Reply) ProtoMessage() {}

func (x *Reply) ProtoReflect() protoreflect.Message {
	mi := &file_pb_service_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Reply.ProtoReflect.Descriptor instead.
func (*Reply) Descriptor() ([]byte, []int) {
	return file_pb_service_proto_rawDescGZIP(), []int{2}
}

func (x *Reply) GetReply() []byte {
	if x != nil {
		return x.Reply
	}
	return nil
}

type Req struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Req []byte `protobuf:"bytes,1,opt,name=Req,proto3" json:"Req,omitempty"`
}

func (x *Req) Reset() {
	*x = Req{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pb_service_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Req) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Req) ProtoMessage() {}

func (x *Req) ProtoReflect() protoreflect.Message {
	mi := &file_pb_service_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Req.ProtoReflect.Descriptor instead.
func (*Req) Descriptor() ([]byte, []int) {
	return file_pb_service_proto_rawDescGZIP(), []int{3}
}

func (x *Req) GetReq() []byte {
	if x != nil {
		return x.Req
	}
	return nil
}

type Addr struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Addr string `protobuf:"bytes,1,opt,name=addr,proto3" json:"addr,omitempty"`
}

func (x *Addr) Reset() {
	*x = Addr{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pb_service_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Addr) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Addr) ProtoMessage() {}

func (x *Addr) ProtoReflect() protoreflect.Message {
	mi := &file_pb_service_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Addr.ProtoReflect.Descriptor instead.
func (*Addr) Descriptor() ([]byte, []int) {
	return file_pb_service_proto_rawDescGZIP(), []int{4}
}

func (x *Addr) GetAddr() string {
	if x != nil {
		return x.Addr
	}
	return ""
}

type TraceidRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Traceid []byte `protobuf:"bytes,1,opt,name=traceid,proto3" json:"traceid,omitempty"`
}

func (x *TraceidRequest) Reset() {
	*x = TraceidRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pb_service_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TraceidRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TraceidRequest) ProtoMessage() {}

func (x *TraceidRequest) ProtoReflect() protoreflect.Message {
	mi := &file_pb_service_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TraceidRequest.ProtoReflect.Descriptor instead.
func (*TraceidRequest) Descriptor() ([]byte, []int) {
	return file_pb_service_proto_rawDescGZIP(), []int{5}
}

func (x *TraceidRequest) GetTraceid() []byte {
	if x != nil {
		return x.Traceid
	}
	return nil
}

var File_pb_service_proto protoreflect.FileDescriptor

var file_pb_service_proto_rawDesc = []byte{
	0x0a, 0x10, 0x70, 0x62, 0x2f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x12, 0x02, 0x70, 0x62, 0x22, 0x29, 0x0a, 0x09, 0x54, 0x72, 0x61, 0x63, 0x65, 0x44,
	0x61, 0x74, 0x61, 0x12, 0x1c, 0x0a, 0x09, 0x74, 0x72, 0x61, 0x63, 0x65, 0x64, 0x61, 0x74, 0x61,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x09, 0x74, 0x72, 0x61, 0x63, 0x65, 0x64, 0x61, 0x74,
	0x61, 0x22, 0x42, 0x0a, 0x0a, 0x54, 0x61, 0x72, 0x67, 0x65, 0x74, 0x49, 0x6e, 0x66, 0x6f, 0x12,
	0x18, 0x0a, 0x07, 0x74, 0x72, 0x61, 0x63, 0x65, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03,
	0x52, 0x07, 0x74, 0x72, 0x61, 0x63, 0x65, 0x69, 0x64, 0x12, 0x1a, 0x0a, 0x08, 0x63, 0x68, 0x65,
	0x63, 0x6b, 0x63, 0x75, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x08, 0x63, 0x68, 0x65,
	0x63, 0x6b, 0x63, 0x75, 0x72, 0x22, 0x1d, 0x0a, 0x05, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x12, 0x14,
	0x0a, 0x05, 0x72, 0x65, 0x70, 0x6c, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x05, 0x72,
	0x65, 0x70, 0x6c, 0x79, 0x22, 0x17, 0x0a, 0x03, 0x52, 0x65, 0x71, 0x12, 0x10, 0x0a, 0x03, 0x52,
	0x65, 0x71, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x03, 0x52, 0x65, 0x71, 0x22, 0x1a, 0x0a,
	0x04, 0x41, 0x64, 0x64, 0x72, 0x12, 0x12, 0x0a, 0x04, 0x61, 0x64, 0x64, 0x72, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x04, 0x61, 0x64, 0x64, 0x72, 0x22, 0x2a, 0x0a, 0x0e, 0x54, 0x72, 0x61,
	0x63, 0x65, 0x69, 0x64, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x18, 0x0a, 0x07, 0x74,
	0x72, 0x61, 0x63, 0x65, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x07, 0x74, 0x72,
	0x61, 0x63, 0x65, 0x69, 0x64, 0x32, 0x72, 0x0a, 0x0c, 0x41, 0x67, 0x65, 0x6e, 0x74, 0x53, 0x65,
	0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x35, 0x0a, 0x14, 0x4e, 0x6f, 0x74, 0x69, 0x66, 0x79, 0x54,
	0x61, 0x72, 0x67, 0x65, 0x74, 0x54, 0x72, 0x61, 0x63, 0x65, 0x69, 0x64, 0x73, 0x12, 0x0e, 0x2e,
	0x70, 0x62, 0x2e, 0x54, 0x61, 0x72, 0x67, 0x65, 0x74, 0x49, 0x6e, 0x66, 0x6f, 0x1a, 0x09, 0x2e,
	0x70, 0x62, 0x2e, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x22, 0x00, 0x28, 0x01, 0x12, 0x2b, 0x0a, 0x13,
	0x4e, 0x6f, 0x74, 0x69, 0x66, 0x79, 0x41, 0x6c, 0x6c, 0x46, 0x69, 0x6c, 0x74, 0x65, 0x72, 0x4f,
	0x76, 0x65, 0x72, 0x12, 0x07, 0x2e, 0x70, 0x62, 0x2e, 0x52, 0x65, 0x71, 0x1a, 0x09, 0x2e, 0x70,
	0x62, 0x2e, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x22, 0x00, 0x32, 0xf8, 0x01, 0x0a, 0x0e, 0x50, 0x72,
	0x6f, 0x63, 0x65, 0x73, 0x73, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x33, 0x0a, 0x10,
	0x53, 0x65, 0x74, 0x54, 0x61, 0x72, 0x67, 0x65, 0x74, 0x54, 0x72, 0x61, 0x63, 0x65, 0x69, 0x64,
	0x12, 0x12, 0x2e, 0x70, 0x62, 0x2e, 0x54, 0x72, 0x61, 0x63, 0x65, 0x69, 0x64, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x1a, 0x09, 0x2e, 0x70, 0x62, 0x2e, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x22,
	0x00, 0x12, 0x2d, 0x0a, 0x0d, 0x53, 0x65, 0x6e, 0x64, 0x54, 0x72, 0x61, 0x63, 0x65, 0x44, 0x61,
	0x74, 0x61, 0x12, 0x0d, 0x2e, 0x70, 0x62, 0x2e, 0x54, 0x72, 0x61, 0x63, 0x65, 0x44, 0x61, 0x74,
	0x61, 0x1a, 0x09, 0x2e, 0x70, 0x62, 0x2e, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x22, 0x00, 0x28, 0x01,
	0x12, 0x2e, 0x0a, 0x0d, 0x53, 0x65, 0x6e, 0x64, 0x54, 0x61, 0x72, 0x67, 0x65, 0x74, 0x49, 0x64,
	0x73, 0x12, 0x0e, 0x2e, 0x70, 0x62, 0x2e, 0x54, 0x61, 0x72, 0x67, 0x65, 0x74, 0x49, 0x6e, 0x66,
	0x6f, 0x1a, 0x09, 0x2e, 0x70, 0x62, 0x2e, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x22, 0x00, 0x28, 0x01,
	0x12, 0x29, 0x0a, 0x10, 0x4e, 0x6f, 0x74, 0x69, 0x66, 0x79, 0x46, 0x69, 0x6c, 0x74, 0x65, 0x72,
	0x4f, 0x76, 0x65, 0x72, 0x12, 0x08, 0x2e, 0x70, 0x62, 0x2e, 0x41, 0x64, 0x64, 0x72, 0x1a, 0x09,
	0x2e, 0x70, 0x62, 0x2e, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x22, 0x00, 0x12, 0x27, 0x0a, 0x0e, 0x4e,
	0x6f, 0x74, 0x69, 0x66, 0x79, 0x53, 0x65, 0x6e, 0x64, 0x4f, 0x76, 0x65, 0x72, 0x12, 0x08, 0x2e,
	0x70, 0x62, 0x2e, 0x41, 0x64, 0x64, 0x72, 0x1a, 0x09, 0x2e, 0x70, 0x62, 0x2e, 0x52, 0x65, 0x70,
	0x6c, 0x79, 0x22, 0x00, 0x42, 0x05, 0x5a, 0x03, 0x2f, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x33,
}

var (
	file_pb_service_proto_rawDescOnce sync.Once
	file_pb_service_proto_rawDescData = file_pb_service_proto_rawDesc
)

func file_pb_service_proto_rawDescGZIP() []byte {
	file_pb_service_proto_rawDescOnce.Do(func() {
		file_pb_service_proto_rawDescData = protoimpl.X.CompressGZIP(file_pb_service_proto_rawDescData)
	})
	return file_pb_service_proto_rawDescData
}

var file_pb_service_proto_msgTypes = make([]protoimpl.MessageInfo, 6)
var file_pb_service_proto_goTypes = []interface{}{
	(*TraceData)(nil),      // 0: pb.TraceData
	(*TargetInfo)(nil),     // 1: pb.TargetInfo
	(*Reply)(nil),          // 2: pb.Reply
	(*Req)(nil),            // 3: pb.Req
	(*Addr)(nil),           // 4: pb.Addr
	(*TraceidRequest)(nil), // 5: pb.TraceidRequest
}
var file_pb_service_proto_depIdxs = []int32{
	1, // 0: pb.AgentService.NotifyTargetTraceids:input_type -> pb.TargetInfo
	3, // 1: pb.AgentService.NotifyAllFilterOver:input_type -> pb.Req
	5, // 2: pb.ProcessService.SetTargetTraceid:input_type -> pb.TraceidRequest
	0, // 3: pb.ProcessService.SendTraceData:input_type -> pb.TraceData
	1, // 4: pb.ProcessService.SendTargetIds:input_type -> pb.TargetInfo
	4, // 5: pb.ProcessService.NotifyFilterOver:input_type -> pb.Addr
	4, // 6: pb.ProcessService.NotifySendOver:input_type -> pb.Addr
	2, // 7: pb.AgentService.NotifyTargetTraceids:output_type -> pb.Reply
	2, // 8: pb.AgentService.NotifyAllFilterOver:output_type -> pb.Reply
	2, // 9: pb.ProcessService.SetTargetTraceid:output_type -> pb.Reply
	2, // 10: pb.ProcessService.SendTraceData:output_type -> pb.Reply
	2, // 11: pb.ProcessService.SendTargetIds:output_type -> pb.Reply
	2, // 12: pb.ProcessService.NotifyFilterOver:output_type -> pb.Reply
	2, // 13: pb.ProcessService.NotifySendOver:output_type -> pb.Reply
	7, // [7:14] is the sub-list for method output_type
	0, // [0:7] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_pb_service_proto_init() }
func file_pb_service_proto_init() {
	if File_pb_service_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_pb_service_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TraceData); i {
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
		file_pb_service_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TargetInfo); i {
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
		file_pb_service_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Reply); i {
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
		file_pb_service_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Req); i {
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
		file_pb_service_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Addr); i {
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
		file_pb_service_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TraceidRequest); i {
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
			RawDescriptor: file_pb_service_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   6,
			NumExtensions: 0,
			NumServices:   2,
		},
		GoTypes:           file_pb_service_proto_goTypes,
		DependencyIndexes: file_pb_service_proto_depIdxs,
		MessageInfos:      file_pb_service_proto_msgTypes,
	}.Build()
	File_pb_service_proto = out.File
	file_pb_service_proto_rawDesc = nil
	file_pb_service_proto_goTypes = nil
	file_pb_service_proto_depIdxs = nil
}
