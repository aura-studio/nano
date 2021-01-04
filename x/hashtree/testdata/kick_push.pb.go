// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.24.0
// 	protoc        v3.12.1
// source: kick_push.proto

package testdata

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

type KickPush struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Code int64 `protobuf:"varint,1,opt,name=Code,proto3" json:"Code,omitempty"`
}

func (x *KickPush) Reset() {
	*x = KickPush{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kick_push_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *KickPush) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*KickPush) ProtoMessage() {}

func (x *KickPush) ProtoReflect() protoreflect.Message {
	mi := &file_kick_push_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use KickPush.ProtoReflect.Descriptor instead.
func (*KickPush) Descriptor() ([]byte, []int) {
	return file_kick_push_proto_rawDescGZIP(), []int{0}
}

func (x *KickPush) GetCode() int64 {
	if x != nil {
		return x.Code
	}
	return 0
}

var File_kick_push_proto protoreflect.FileDescriptor

var file_kick_push_proto_rawDesc = []byte{
	0x0a, 0x0f, 0x6b, 0x69, 0x63, 0x6b, 0x5f, 0x70, 0x75, 0x73, 0x68, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x08, 0x74, 0x65, 0x73, 0x74, 0x64, 0x61, 0x74, 0x61, 0x22, 0x1e, 0x0a, 0x08, 0x4b,
	0x69, 0x63, 0x6b, 0x50, 0x75, 0x73, 0x68, 0x12, 0x12, 0x0a, 0x04, 0x43, 0x6f, 0x64, 0x65, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x04, 0x43, 0x6f, 0x64, 0x65, 0x62, 0x06, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x33,
}

var (
	file_kick_push_proto_rawDescOnce sync.Once
	file_kick_push_proto_rawDescData = file_kick_push_proto_rawDesc
)

func file_kick_push_proto_rawDescGZIP() []byte {
	file_kick_push_proto_rawDescOnce.Do(func() {
		file_kick_push_proto_rawDescData = protoimpl.X.CompressGZIP(file_kick_push_proto_rawDescData)
	})
	return file_kick_push_proto_rawDescData
}

var file_kick_push_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_kick_push_proto_goTypes = []interface{}{
	(*KickPush)(nil), // 0: testdata.KickPush
}
var file_kick_push_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_kick_push_proto_init() }
func file_kick_push_proto_init() {
	if File_kick_push_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_kick_push_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*KickPush); i {
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
			RawDescriptor: file_kick_push_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_kick_push_proto_goTypes,
		DependencyIndexes: file_kick_push_proto_depIdxs,
		MessageInfos:      file_kick_push_proto_msgTypes,
	}.Build()
	File_kick_push_proto = out.File
	file_kick_push_proto_rawDesc = nil
	file_kick_push_proto_goTypes = nil
	file_kick_push_proto_depIdxs = nil
}