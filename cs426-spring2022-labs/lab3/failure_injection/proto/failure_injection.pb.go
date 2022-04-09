// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0
// 	protoc        v3.19.4
// source: failure_injection/proto/failure_injection.proto

package proto

import (
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

type InjectionConfig struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// 0 == no latency injection
	SleepNs int64 `protobuf:"varint,1,opt,name=sleep_ns,json=sleepNs,proto3" json:"sleep_ns,omitempty"`
	// inject hard failure one in failure_rate requests
	// 0 == turn off injection
	FailureRate int64 `protobuf:"varint,2,opt,name=failure_rate,json=failureRate,proto3" json:"failure_rate,omitempty"`
	// inject response omission failure one in omission_response_rate requests
	// 0 == turn off injection
	ResponseOmissionRate int64 `protobuf:"varint,3,opt,name=response_omission_rate,json=responseOmissionRate,proto3" json:"response_omission_rate,omitempty"`
}

func (x *InjectionConfig) Reset() {
	*x = InjectionConfig{}
	if protoimpl.UnsafeEnabled {
		mi := &file_failure_injection_proto_failure_injection_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *InjectionConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*InjectionConfig) ProtoMessage() {}

func (x *InjectionConfig) ProtoReflect() protoreflect.Message {
	mi := &file_failure_injection_proto_failure_injection_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use InjectionConfig.ProtoReflect.Descriptor instead.
func (*InjectionConfig) Descriptor() ([]byte, []int) {
	return file_failure_injection_proto_failure_injection_proto_rawDescGZIP(), []int{0}
}

func (x *InjectionConfig) GetSleepNs() int64 {
	if x != nil {
		return x.SleepNs
	}
	return 0
}

func (x *InjectionConfig) GetFailureRate() int64 {
	if x != nil {
		return x.FailureRate
	}
	return 0
}

func (x *InjectionConfig) GetResponseOmissionRate() int64 {
	if x != nil {
		return x.ResponseOmissionRate
	}
	return 0
}

type SetInjectionConfigRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Config *InjectionConfig `protobuf:"bytes,1,opt,name=config,proto3" json:"config,omitempty"`
}

func (x *SetInjectionConfigRequest) Reset() {
	*x = SetInjectionConfigRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_failure_injection_proto_failure_injection_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SetInjectionConfigRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SetInjectionConfigRequest) ProtoMessage() {}

func (x *SetInjectionConfigRequest) ProtoReflect() protoreflect.Message {
	mi := &file_failure_injection_proto_failure_injection_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SetInjectionConfigRequest.ProtoReflect.Descriptor instead.
func (*SetInjectionConfigRequest) Descriptor() ([]byte, []int) {
	return file_failure_injection_proto_failure_injection_proto_rawDescGZIP(), []int{1}
}

func (x *SetInjectionConfigRequest) GetConfig() *InjectionConfig {
	if x != nil {
		return x.Config
	}
	return nil
}

type SetInjectionConfigResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *SetInjectionConfigResponse) Reset() {
	*x = SetInjectionConfigResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_failure_injection_proto_failure_injection_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SetInjectionConfigResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SetInjectionConfigResponse) ProtoMessage() {}

func (x *SetInjectionConfigResponse) ProtoReflect() protoreflect.Message {
	mi := &file_failure_injection_proto_failure_injection_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SetInjectionConfigResponse.ProtoReflect.Descriptor instead.
func (*SetInjectionConfigResponse) Descriptor() ([]byte, []int) {
	return file_failure_injection_proto_failure_injection_proto_rawDescGZIP(), []int{2}
}

var File_failure_injection_proto_failure_injection_proto protoreflect.FileDescriptor

var file_failure_injection_proto_failure_injection_proto_rawDesc = []byte{
	0x0a, 0x2f, 0x66, 0x61, 0x69, 0x6c, 0x75, 0x72, 0x65, 0x5f, 0x69, 0x6e, 0x6a, 0x65, 0x63, 0x74,
	0x69, 0x6f, 0x6e, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x66, 0x61, 0x69, 0x6c, 0x75, 0x72,
	0x65, 0x5f, 0x69, 0x6e, 0x6a, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x11, 0x66, 0x61, 0x69, 0x6c, 0x75, 0x72, 0x65, 0x5f, 0x69, 0x6e, 0x6a, 0x65, 0x63,
	0x74, 0x69, 0x6f, 0x6e, 0x22, 0x85, 0x01, 0x0a, 0x0f, 0x49, 0x6e, 0x6a, 0x65, 0x63, 0x74, 0x69,
	0x6f, 0x6e, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x19, 0x0a, 0x08, 0x73, 0x6c, 0x65, 0x65,
	0x70, 0x5f, 0x6e, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x07, 0x73, 0x6c, 0x65, 0x65,
	0x70, 0x4e, 0x73, 0x12, 0x21, 0x0a, 0x0c, 0x66, 0x61, 0x69, 0x6c, 0x75, 0x72, 0x65, 0x5f, 0x72,
	0x61, 0x74, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0b, 0x66, 0x61, 0x69, 0x6c, 0x75,
	0x72, 0x65, 0x52, 0x61, 0x74, 0x65, 0x12, 0x34, 0x0a, 0x16, 0x72, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x5f, 0x6f, 0x6d, 0x69, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x5f, 0x72, 0x61, 0x74, 0x65,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x03, 0x52, 0x14, 0x72, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x4f, 0x6d, 0x69, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x52, 0x61, 0x74, 0x65, 0x22, 0x57, 0x0a, 0x19,
	0x53, 0x65, 0x74, 0x49, 0x6e, 0x6a, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x43, 0x6f, 0x6e, 0x66,
	0x69, 0x67, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x3a, 0x0a, 0x06, 0x63, 0x6f, 0x6e,
	0x66, 0x69, 0x67, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x22, 0x2e, 0x66, 0x61, 0x69, 0x6c,
	0x75, 0x72, 0x65, 0x5f, 0x69, 0x6e, 0x6a, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x49, 0x6e,
	0x6a, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x52, 0x06, 0x63,
	0x6f, 0x6e, 0x66, 0x69, 0x67, 0x22, 0x1c, 0x0a, 0x1a, 0x53, 0x65, 0x74, 0x49, 0x6e, 0x6a, 0x65,
	0x63, 0x74, 0x69, 0x6f, 0x6e, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x42, 0x2d, 0x5a, 0x2b, 0x63, 0x73, 0x34, 0x32, 0x36, 0x2e, 0x79, 0x61, 0x6c,
	0x65, 0x2e, 0x65, 0x64, 0x75, 0x2f, 0x6c, 0x61, 0x62, 0x31, 0x2f, 0x66, 0x61, 0x69, 0x6c, 0x75,
	0x72, 0x65, 0x5f, 0x69, 0x6e, 0x6a, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x2f, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_failure_injection_proto_failure_injection_proto_rawDescOnce sync.Once
	file_failure_injection_proto_failure_injection_proto_rawDescData = file_failure_injection_proto_failure_injection_proto_rawDesc
)

func file_failure_injection_proto_failure_injection_proto_rawDescGZIP() []byte {
	file_failure_injection_proto_failure_injection_proto_rawDescOnce.Do(func() {
		file_failure_injection_proto_failure_injection_proto_rawDescData = protoimpl.X.CompressGZIP(file_failure_injection_proto_failure_injection_proto_rawDescData)
	})
	return file_failure_injection_proto_failure_injection_proto_rawDescData
}

var file_failure_injection_proto_failure_injection_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_failure_injection_proto_failure_injection_proto_goTypes = []interface{}{
	(*InjectionConfig)(nil),            // 0: failure_injection.InjectionConfig
	(*SetInjectionConfigRequest)(nil),  // 1: failure_injection.SetInjectionConfigRequest
	(*SetInjectionConfigResponse)(nil), // 2: failure_injection.SetInjectionConfigResponse
}
var file_failure_injection_proto_failure_injection_proto_depIdxs = []int32{
	0, // 0: failure_injection.SetInjectionConfigRequest.config:type_name -> failure_injection.InjectionConfig
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_failure_injection_proto_failure_injection_proto_init() }
func file_failure_injection_proto_failure_injection_proto_init() {
	if File_failure_injection_proto_failure_injection_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_failure_injection_proto_failure_injection_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*InjectionConfig); i {
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
		file_failure_injection_proto_failure_injection_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SetInjectionConfigRequest); i {
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
		file_failure_injection_proto_failure_injection_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SetInjectionConfigResponse); i {
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
			RawDescriptor: file_failure_injection_proto_failure_injection_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_failure_injection_proto_failure_injection_proto_goTypes,
		DependencyIndexes: file_failure_injection_proto_failure_injection_proto_depIdxs,
		MessageInfos:      file_failure_injection_proto_failure_injection_proto_msgTypes,
	}.Build()
	File_failure_injection_proto_failure_injection_proto = out.File
	file_failure_injection_proto_failure_injection_proto_rawDesc = nil
	file_failure_injection_proto_failure_injection_proto_goTypes = nil
	file_failure_injection_proto_failure_injection_proto_depIdxs = nil
}
