// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.30.0
// 	protoc        v4.22.0
// source: test.proto

package testpb

import (
	reflect "reflect"
	sync "sync"

	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type PingEmptyRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *PingEmptyRequest) Reset() {
	*x = PingEmptyRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_test_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PingEmptyRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PingEmptyRequest) ProtoMessage() {}

func (x *PingEmptyRequest) ProtoReflect() protoreflect.Message {
	mi := &file_test_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PingEmptyRequest.ProtoReflect.Descriptor instead.
func (*PingEmptyRequest) Descriptor() ([]byte, []int) {
	return file_test_proto_rawDescGZIP(), []int{0}
}

type PingEmptyResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *PingEmptyResponse) Reset() {
	*x = PingEmptyResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_test_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PingEmptyResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PingEmptyResponse) ProtoMessage() {}

func (x *PingEmptyResponse) ProtoReflect() protoreflect.Message {
	mi := &file_test_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PingEmptyResponse.ProtoReflect.Descriptor instead.
func (*PingEmptyResponse) Descriptor() ([]byte, []int) {
	return file_test_proto_rawDescGZIP(), []int{1}
}

type PingRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Value             string `protobuf:"bytes,1,opt,name=value,proto3" json:"value,omitempty"`
	SleepTimeMs       int32  `protobuf:"varint,2,opt,name=sleep_time_ms,json=sleepTimeMs,proto3" json:"sleep_time_ms,omitempty"`
	ErrorCodeReturned uint32 `protobuf:"varint,3,opt,name=error_code_returned,json=errorCodeReturned,proto3" json:"error_code_returned,omitempty"`
}

func (x *PingRequest) Reset() {
	*x = PingRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_test_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PingRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PingRequest) ProtoMessage() {}

func (x *PingRequest) ProtoReflect() protoreflect.Message {
	mi := &file_test_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PingRequest.ProtoReflect.Descriptor instead.
func (*PingRequest) Descriptor() ([]byte, []int) {
	return file_test_proto_rawDescGZIP(), []int{2}
}

func (x *PingRequest) GetValue() string {
	if x != nil {
		return x.Value
	}
	return ""
}

func (x *PingRequest) GetSleepTimeMs() int32 {
	if x != nil {
		return x.SleepTimeMs
	}
	return 0
}

func (x *PingRequest) GetErrorCodeReturned() uint32 {
	if x != nil {
		return x.ErrorCodeReturned
	}
	return 0
}

type PingResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Value   string `protobuf:"bytes,1,opt,name=value,proto3" json:"value,omitempty"`
	Counter int32  `protobuf:"varint,2,opt,name=counter,proto3" json:"counter,omitempty"`
}

func (x *PingResponse) Reset() {
	*x = PingResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_test_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PingResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PingResponse) ProtoMessage() {}

func (x *PingResponse) ProtoReflect() protoreflect.Message {
	mi := &file_test_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PingResponse.ProtoReflect.Descriptor instead.
func (*PingResponse) Descriptor() ([]byte, []int) {
	return file_test_proto_rawDescGZIP(), []int{3}
}

func (x *PingResponse) GetValue() string {
	if x != nil {
		return x.Value
	}
	return ""
}

func (x *PingResponse) GetCounter() int32 {
	if x != nil {
		return x.Counter
	}
	return 0
}

type PingErrorRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Value             string `protobuf:"bytes,1,opt,name=value,proto3" json:"value,omitempty"`
	SleepTimeMs       int32  `protobuf:"varint,2,opt,name=sleep_time_ms,json=sleepTimeMs,proto3" json:"sleep_time_ms,omitempty"`
	ErrorCodeReturned uint32 `protobuf:"varint,3,opt,name=error_code_returned,json=errorCodeReturned,proto3" json:"error_code_returned,omitempty"`
}

func (x *PingErrorRequest) Reset() {
	*x = PingErrorRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_test_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PingErrorRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PingErrorRequest) ProtoMessage() {}

func (x *PingErrorRequest) ProtoReflect() protoreflect.Message {
	mi := &file_test_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PingErrorRequest.ProtoReflect.Descriptor instead.
func (*PingErrorRequest) Descriptor() ([]byte, []int) {
	return file_test_proto_rawDescGZIP(), []int{4}
}

func (x *PingErrorRequest) GetValue() string {
	if x != nil {
		return x.Value
	}
	return ""
}

func (x *PingErrorRequest) GetSleepTimeMs() int32 {
	if x != nil {
		return x.SleepTimeMs
	}
	return 0
}

func (x *PingErrorRequest) GetErrorCodeReturned() uint32 {
	if x != nil {
		return x.ErrorCodeReturned
	}
	return 0
}

type PingErrorResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Value   string `protobuf:"bytes,1,opt,name=value,proto3" json:"value,omitempty"`
	Counter int32  `protobuf:"varint,2,opt,name=counter,proto3" json:"counter,omitempty"`
}

func (x *PingErrorResponse) Reset() {
	*x = PingErrorResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_test_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PingErrorResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PingErrorResponse) ProtoMessage() {}

func (x *PingErrorResponse) ProtoReflect() protoreflect.Message {
	mi := &file_test_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PingErrorResponse.ProtoReflect.Descriptor instead.
func (*PingErrorResponse) Descriptor() ([]byte, []int) {
	return file_test_proto_rawDescGZIP(), []int{5}
}

func (x *PingErrorResponse) GetValue() string {
	if x != nil {
		return x.Value
	}
	return ""
}

func (x *PingErrorResponse) GetCounter() int32 {
	if x != nil {
		return x.Counter
	}
	return 0
}

type PingListRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Value             string `protobuf:"bytes,1,opt,name=value,proto3" json:"value,omitempty"`
	SleepTimeMs       int32  `protobuf:"varint,2,opt,name=sleep_time_ms,json=sleepTimeMs,proto3" json:"sleep_time_ms,omitempty"`
	ErrorCodeReturned uint32 `protobuf:"varint,3,opt,name=error_code_returned,json=errorCodeReturned,proto3" json:"error_code_returned,omitempty"`
}

func (x *PingListRequest) Reset() {
	*x = PingListRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_test_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PingListRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PingListRequest) ProtoMessage() {}

func (x *PingListRequest) ProtoReflect() protoreflect.Message {
	mi := &file_test_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PingListRequest.ProtoReflect.Descriptor instead.
func (*PingListRequest) Descriptor() ([]byte, []int) {
	return file_test_proto_rawDescGZIP(), []int{6}
}

func (x *PingListRequest) GetValue() string {
	if x != nil {
		return x.Value
	}
	return ""
}

func (x *PingListRequest) GetSleepTimeMs() int32 {
	if x != nil {
		return x.SleepTimeMs
	}
	return 0
}

func (x *PingListRequest) GetErrorCodeReturned() uint32 {
	if x != nil {
		return x.ErrorCodeReturned
	}
	return 0
}

type PingListResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Value   string `protobuf:"bytes,1,opt,name=value,proto3" json:"value,omitempty"`
	Counter int32  `protobuf:"varint,2,opt,name=counter,proto3" json:"counter,omitempty"`
}

func (x *PingListResponse) Reset() {
	*x = PingListResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_test_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PingListResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PingListResponse) ProtoMessage() {}

func (x *PingListResponse) ProtoReflect() protoreflect.Message {
	mi := &file_test_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PingListResponse.ProtoReflect.Descriptor instead.
func (*PingListResponse) Descriptor() ([]byte, []int) {
	return file_test_proto_rawDescGZIP(), []int{7}
}

func (x *PingListResponse) GetValue() string {
	if x != nil {
		return x.Value
	}
	return ""
}

func (x *PingListResponse) GetCounter() int32 {
	if x != nil {
		return x.Counter
	}
	return 0
}

type PingStreamRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Value             string `protobuf:"bytes,1,opt,name=value,proto3" json:"value,omitempty"`
	SleepTimeMs       int32  `protobuf:"varint,2,opt,name=sleep_time_ms,json=sleepTimeMs,proto3" json:"sleep_time_ms,omitempty"`
	ErrorCodeReturned uint32 `protobuf:"varint,3,opt,name=error_code_returned,json=errorCodeReturned,proto3" json:"error_code_returned,omitempty"`
}

func (x *PingStreamRequest) Reset() {
	*x = PingStreamRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_test_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PingStreamRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PingStreamRequest) ProtoMessage() {}

func (x *PingStreamRequest) ProtoReflect() protoreflect.Message {
	mi := &file_test_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PingStreamRequest.ProtoReflect.Descriptor instead.
func (*PingStreamRequest) Descriptor() ([]byte, []int) {
	return file_test_proto_rawDescGZIP(), []int{8}
}

func (x *PingStreamRequest) GetValue() string {
	if x != nil {
		return x.Value
	}
	return ""
}

func (x *PingStreamRequest) GetSleepTimeMs() int32 {
	if x != nil {
		return x.SleepTimeMs
	}
	return 0
}

func (x *PingStreamRequest) GetErrorCodeReturned() uint32 {
	if x != nil {
		return x.ErrorCodeReturned
	}
	return 0
}

type PingStreamResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Value   string `protobuf:"bytes,1,opt,name=value,proto3" json:"value,omitempty"`
	Counter int32  `protobuf:"varint,2,opt,name=counter,proto3" json:"counter,omitempty"`
}

func (x *PingStreamResponse) Reset() {
	*x = PingStreamResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_test_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PingStreamResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PingStreamResponse) ProtoMessage() {}

func (x *PingStreamResponse) ProtoReflect() protoreflect.Message {
	mi := &file_test_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PingStreamResponse.ProtoReflect.Descriptor instead.
func (*PingStreamResponse) Descriptor() ([]byte, []int) {
	return file_test_proto_rawDescGZIP(), []int{9}
}

func (x *PingStreamResponse) GetValue() string {
	if x != nil {
		return x.Value
	}
	return ""
}

func (x *PingStreamResponse) GetCounter() int32 {
	if x != nil {
		return x.Counter
	}
	return 0
}

var File_test_proto protoreflect.FileDescriptor

var file_test_proto_rawDesc = []byte{
	0x0a, 0x0a, 0x74, 0x65, 0x73, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x11, 0x74, 0x65,
	0x73, 0x74, 0x69, 0x6e, 0x67, 0x2e, 0x74, 0x65, 0x73, 0x74, 0x70, 0x62, 0x2e, 0x76, 0x31, 0x22,
	0x12, 0x0a, 0x10, 0x50, 0x69, 0x6e, 0x67, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x22, 0x13, 0x0a, 0x11, 0x50, 0x69, 0x6e, 0x67, 0x45, 0x6d, 0x70, 0x74, 0x79,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x77, 0x0a, 0x0b, 0x50, 0x69, 0x6e, 0x67,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x12, 0x22, 0x0a,
	0x0d, 0x73, 0x6c, 0x65, 0x65, 0x70, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x5f, 0x6d, 0x73, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x05, 0x52, 0x0b, 0x73, 0x6c, 0x65, 0x65, 0x70, 0x54, 0x69, 0x6d, 0x65, 0x4d,
	0x73, 0x12, 0x2e, 0x0a, 0x13, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x5f, 0x63, 0x6f, 0x64, 0x65, 0x5f,
	0x72, 0x65, 0x74, 0x75, 0x72, 0x6e, 0x65, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x11,
	0x65, 0x72, 0x72, 0x6f, 0x72, 0x43, 0x6f, 0x64, 0x65, 0x52, 0x65, 0x74, 0x75, 0x72, 0x6e, 0x65,
	0x64, 0x22, 0x3e, 0x0a, 0x0c, 0x50, 0x69, 0x6e, 0x67, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x63, 0x6f, 0x75, 0x6e, 0x74,
	0x65, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x07, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x65,
	0x72, 0x22, 0x7c, 0x0a, 0x10, 0x50, 0x69, 0x6e, 0x67, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x12, 0x22, 0x0a, 0x0d, 0x73,
	0x6c, 0x65, 0x65, 0x70, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x5f, 0x6d, 0x73, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x05, 0x52, 0x0b, 0x73, 0x6c, 0x65, 0x65, 0x70, 0x54, 0x69, 0x6d, 0x65, 0x4d, 0x73, 0x12,
	0x2e, 0x0a, 0x13, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x5f, 0x63, 0x6f, 0x64, 0x65, 0x5f, 0x72, 0x65,
	0x74, 0x75, 0x72, 0x6e, 0x65, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x11, 0x65, 0x72,
	0x72, 0x6f, 0x72, 0x43, 0x6f, 0x64, 0x65, 0x52, 0x65, 0x74, 0x75, 0x72, 0x6e, 0x65, 0x64, 0x22,
	0x43, 0x0a, 0x11, 0x50, 0x69, 0x6e, 0x67, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x63, 0x6f,
	0x75, 0x6e, 0x74, 0x65, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x07, 0x63, 0x6f, 0x75,
	0x6e, 0x74, 0x65, 0x72, 0x22, 0x7b, 0x0a, 0x0f, 0x50, 0x69, 0x6e, 0x67, 0x4c, 0x69, 0x73, 0x74,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x12, 0x22, 0x0a,
	0x0d, 0x73, 0x6c, 0x65, 0x65, 0x70, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x5f, 0x6d, 0x73, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x05, 0x52, 0x0b, 0x73, 0x6c, 0x65, 0x65, 0x70, 0x54, 0x69, 0x6d, 0x65, 0x4d,
	0x73, 0x12, 0x2e, 0x0a, 0x13, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x5f, 0x63, 0x6f, 0x64, 0x65, 0x5f,
	0x72, 0x65, 0x74, 0x75, 0x72, 0x6e, 0x65, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x11,
	0x65, 0x72, 0x72, 0x6f, 0x72, 0x43, 0x6f, 0x64, 0x65, 0x52, 0x65, 0x74, 0x75, 0x72, 0x6e, 0x65,
	0x64, 0x22, 0x42, 0x0a, 0x10, 0x50, 0x69, 0x6e, 0x67, 0x4c, 0x69, 0x73, 0x74, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x63,
	0x6f, 0x75, 0x6e, 0x74, 0x65, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x07, 0x63, 0x6f,
	0x75, 0x6e, 0x74, 0x65, 0x72, 0x22, 0x7d, 0x0a, 0x11, 0x50, 0x69, 0x6e, 0x67, 0x53, 0x74, 0x72,
	0x65, 0x61, 0x6d, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61,
	0x6c, 0x75, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65,
	0x12, 0x22, 0x0a, 0x0d, 0x73, 0x6c, 0x65, 0x65, 0x70, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x5f, 0x6d,
	0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0b, 0x73, 0x6c, 0x65, 0x65, 0x70, 0x54, 0x69,
	0x6d, 0x65, 0x4d, 0x73, 0x12, 0x2e, 0x0a, 0x13, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x5f, 0x63, 0x6f,
	0x64, 0x65, 0x5f, 0x72, 0x65, 0x74, 0x75, 0x72, 0x6e, 0x65, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x0d, 0x52, 0x11, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x43, 0x6f, 0x64, 0x65, 0x52, 0x65, 0x74, 0x75,
	0x72, 0x6e, 0x65, 0x64, 0x22, 0x44, 0x0a, 0x12, 0x50, 0x69, 0x6e, 0x67, 0x53, 0x74, 0x72, 0x65,
	0x61, 0x6d, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61,
	0x6c, 0x75, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65,
	0x12, 0x18, 0x0a, 0x07, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x65, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x05, 0x52, 0x07, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x65, 0x72, 0x32, 0xc6, 0x03, 0x0a, 0x0b, 0x54,
	0x65, 0x73, 0x74, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x58, 0x0a, 0x09, 0x50, 0x69,
	0x6e, 0x67, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x12, 0x23, 0x2e, 0x74, 0x65, 0x73, 0x74, 0x69, 0x6e,
	0x67, 0x2e, 0x74, 0x65, 0x73, 0x74, 0x70, 0x62, 0x2e, 0x76, 0x31, 0x2e, 0x50, 0x69, 0x6e, 0x67,
	0x45, 0x6d, 0x70, 0x74, 0x79, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x24, 0x2e, 0x74,
	0x65, 0x73, 0x74, 0x69, 0x6e, 0x67, 0x2e, 0x74, 0x65, 0x73, 0x74, 0x70, 0x62, 0x2e, 0x76, 0x31,
	0x2e, 0x50, 0x69, 0x6e, 0x67, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x22, 0x00, 0x12, 0x49, 0x0a, 0x04, 0x50, 0x69, 0x6e, 0x67, 0x12, 0x1e, 0x2e, 0x74,
	0x65, 0x73, 0x74, 0x69, 0x6e, 0x67, 0x2e, 0x74, 0x65, 0x73, 0x74, 0x70, 0x62, 0x2e, 0x76, 0x31,
	0x2e, 0x50, 0x69, 0x6e, 0x67, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1f, 0x2e, 0x74,
	0x65, 0x73, 0x74, 0x69, 0x6e, 0x67, 0x2e, 0x74, 0x65, 0x73, 0x74, 0x70, 0x62, 0x2e, 0x76, 0x31,
	0x2e, 0x50, 0x69, 0x6e, 0x67, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12,
	0x58, 0x0a, 0x09, 0x50, 0x69, 0x6e, 0x67, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x12, 0x23, 0x2e, 0x74,
	0x65, 0x73, 0x74, 0x69, 0x6e, 0x67, 0x2e, 0x74, 0x65, 0x73, 0x74, 0x70, 0x62, 0x2e, 0x76, 0x31,
	0x2e, 0x50, 0x69, 0x6e, 0x67, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x1a, 0x24, 0x2e, 0x74, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67, 0x2e, 0x74, 0x65, 0x73, 0x74,
	0x70, 0x62, 0x2e, 0x76, 0x31, 0x2e, 0x50, 0x69, 0x6e, 0x67, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x57, 0x0a, 0x08, 0x50, 0x69, 0x6e,
	0x67, 0x4c, 0x69, 0x73, 0x74, 0x12, 0x22, 0x2e, 0x74, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67, 0x2e,
	0x74, 0x65, 0x73, 0x74, 0x70, 0x62, 0x2e, 0x76, 0x31, 0x2e, 0x50, 0x69, 0x6e, 0x67, 0x4c, 0x69,
	0x73, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x23, 0x2e, 0x74, 0x65, 0x73, 0x74,
	0x69, 0x6e, 0x67, 0x2e, 0x74, 0x65, 0x73, 0x74, 0x70, 0x62, 0x2e, 0x76, 0x31, 0x2e, 0x50, 0x69,
	0x6e, 0x67, 0x4c, 0x69, 0x73, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00,
	0x30, 0x01, 0x12, 0x5f, 0x0a, 0x0a, 0x50, 0x69, 0x6e, 0x67, 0x53, 0x74, 0x72, 0x65, 0x61, 0x6d,
	0x12, 0x24, 0x2e, 0x74, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67, 0x2e, 0x74, 0x65, 0x73, 0x74, 0x70,
	0x62, 0x2e, 0x76, 0x31, 0x2e, 0x50, 0x69, 0x6e, 0x67, 0x53, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x25, 0x2e, 0x74, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67,
	0x2e, 0x74, 0x65, 0x73, 0x74, 0x70, 0x62, 0x2e, 0x76, 0x31, 0x2e, 0x50, 0x69, 0x6e, 0x67, 0x53,
	0x74, 0x72, 0x65, 0x61, 0x6d, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x28,
	0x01, 0x30, 0x01, 0x42, 0x0a, 0x5a, 0x08, 0x2e, 0x3b, 0x74, 0x65, 0x73, 0x74, 0x70, 0x62, 0x62,
	0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_test_proto_rawDescOnce sync.Once
	file_test_proto_rawDescData = file_test_proto_rawDesc
)

func file_test_proto_rawDescGZIP() []byte {
	file_test_proto_rawDescOnce.Do(func() {
		file_test_proto_rawDescData = protoimpl.X.CompressGZIP(file_test_proto_rawDescData)
	})
	return file_test_proto_rawDescData
}

var file_test_proto_msgTypes = make([]protoimpl.MessageInfo, 10)
var file_test_proto_goTypes = []interface{}{
	(*PingEmptyRequest)(nil),   // 0: testing.testpb.v1.PingEmptyRequest
	(*PingEmptyResponse)(nil),  // 1: testing.testpb.v1.PingEmptyResponse
	(*PingRequest)(nil),        // 2: testing.testpb.v1.PingRequest
	(*PingResponse)(nil),       // 3: testing.testpb.v1.PingResponse
	(*PingErrorRequest)(nil),   // 4: testing.testpb.v1.PingErrorRequest
	(*PingErrorResponse)(nil),  // 5: testing.testpb.v1.PingErrorResponse
	(*PingListRequest)(nil),    // 6: testing.testpb.v1.PingListRequest
	(*PingListResponse)(nil),   // 7: testing.testpb.v1.PingListResponse
	(*PingStreamRequest)(nil),  // 8: testing.testpb.v1.PingStreamRequest
	(*PingStreamResponse)(nil), // 9: testing.testpb.v1.PingStreamResponse
}
var file_test_proto_depIdxs = []int32{
	0, // 0: testing.testpb.v1.TestService.PingEmpty:input_type -> testing.testpb.v1.PingEmptyRequest
	2, // 1: testing.testpb.v1.TestService.Ping:input_type -> testing.testpb.v1.PingRequest
	4, // 2: testing.testpb.v1.TestService.PingError:input_type -> testing.testpb.v1.PingErrorRequest
	6, // 3: testing.testpb.v1.TestService.PingList:input_type -> testing.testpb.v1.PingListRequest
	8, // 4: testing.testpb.v1.TestService.PingStream:input_type -> testing.testpb.v1.PingStreamRequest
	1, // 5: testing.testpb.v1.TestService.PingEmpty:output_type -> testing.testpb.v1.PingEmptyResponse
	3, // 6: testing.testpb.v1.TestService.Ping:output_type -> testing.testpb.v1.PingResponse
	5, // 7: testing.testpb.v1.TestService.PingError:output_type -> testing.testpb.v1.PingErrorResponse
	7, // 8: testing.testpb.v1.TestService.PingList:output_type -> testing.testpb.v1.PingListResponse
	9, // 9: testing.testpb.v1.TestService.PingStream:output_type -> testing.testpb.v1.PingStreamResponse
	5, // [5:10] is the sub-list for method output_type
	0, // [0:5] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_test_proto_init() }
func file_test_proto_init() {
	if File_test_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_test_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PingEmptyRequest); i {
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
		file_test_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PingEmptyResponse); i {
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
		file_test_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PingRequest); i {
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
		file_test_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PingResponse); i {
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
		file_test_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PingErrorRequest); i {
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
		file_test_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PingErrorResponse); i {
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
		file_test_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PingListRequest); i {
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
		file_test_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PingListResponse); i {
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
		file_test_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PingStreamRequest); i {
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
		file_test_proto_msgTypes[9].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PingStreamResponse); i {
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
			RawDescriptor: file_test_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   10,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_test_proto_goTypes,
		DependencyIndexes: file_test_proto_depIdxs,
		MessageInfos:      file_test_proto_msgTypes,
	}.Build()
	File_test_proto = out.File
	file_test_proto_rawDesc = nil
	file_test_proto_goTypes = nil
	file_test_proto_depIdxs = nil
}
