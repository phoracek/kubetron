// Code generated by protoc-gen-go. DO NOT EDIT.
// source: pkg/cniplugin/server.proto

/*
Package cniplugin is a generated protocol buffer package.

It is generated from these files:
	pkg/cniplugin/server.proto

It has these top-level messages:
	PluginParams
	PluginResult
*/
package cniplugin

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type PluginParams struct {
	ContainerName    *string `protobuf:"bytes,1,req,name=containerName" json:"containerName,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *PluginParams) Reset()                    { *m = PluginParams{} }
func (m *PluginParams) String() string            { return proto.CompactTextString(m) }
func (*PluginParams) ProtoMessage()               {}
func (*PluginParams) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *PluginParams) GetContainerName() string {
	if m != nil && m.ContainerName != nil {
		return *m.ContainerName
	}
	return ""
}

type PluginResult struct {
	Msg              *string `protobuf:"bytes,2,opt,name=msg" json:"msg,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *PluginResult) Reset()                    { *m = PluginResult{} }
func (m *PluginResult) String() string            { return proto.CompactTextString(m) }
func (*PluginResult) ProtoMessage()               {}
func (*PluginResult) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *PluginResult) GetMsg() string {
	if m != nil && m.Msg != nil {
		return *m.Msg
	}
	return ""
}

func init() {
	proto.RegisterType((*PluginParams)(nil), "cniplugin.PluginParams")
	proto.RegisterType((*PluginResult)(nil), "cniplugin.PluginResult")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for Plugin service

type PluginClient interface {
	AddNetwork(ctx context.Context, in *PluginParams, opts ...grpc.CallOption) (*PluginResult, error)
	DelNetwork(ctx context.Context, in *PluginParams, opts ...grpc.CallOption) (*PluginResult, error)
}

type pluginClient struct {
	cc *grpc.ClientConn
}

func NewPluginClient(cc *grpc.ClientConn) PluginClient {
	return &pluginClient{cc}
}

func (c *pluginClient) AddNetwork(ctx context.Context, in *PluginParams, opts ...grpc.CallOption) (*PluginResult, error) {
	out := new(PluginResult)
	err := grpc.Invoke(ctx, "/cniplugin.Plugin/AddNetwork", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *pluginClient) DelNetwork(ctx context.Context, in *PluginParams, opts ...grpc.CallOption) (*PluginResult, error) {
	out := new(PluginResult)
	err := grpc.Invoke(ctx, "/cniplugin.Plugin/DelNetwork", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for Plugin service

type PluginServer interface {
	AddNetwork(context.Context, *PluginParams) (*PluginResult, error)
	DelNetwork(context.Context, *PluginParams) (*PluginResult, error)
}

func RegisterPluginServer(s *grpc.Server, srv PluginServer) {
	s.RegisterService(&_Plugin_serviceDesc, srv)
}

func _Plugin_AddNetwork_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PluginParams)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PluginServer).AddNetwork(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/cniplugin.Plugin/AddNetwork",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PluginServer).AddNetwork(ctx, req.(*PluginParams))
	}
	return interceptor(ctx, in, info, handler)
}

func _Plugin_DelNetwork_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PluginParams)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PluginServer).DelNetwork(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/cniplugin.Plugin/DelNetwork",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PluginServer).DelNetwork(ctx, req.(*PluginParams))
	}
	return interceptor(ctx, in, info, handler)
}

var _Plugin_serviceDesc = grpc.ServiceDesc{
	ServiceName: "cniplugin.Plugin",
	HandlerType: (*PluginServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "AddNetwork",
			Handler:    _Plugin_AddNetwork_Handler,
		},
		{
			MethodName: "DelNetwork",
			Handler:    _Plugin_DelNetwork_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "pkg/cniplugin/server.proto",
}

func init() { proto.RegisterFile("pkg/cniplugin/server.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 165 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x92, 0x2a, 0xc8, 0x4e, 0xd7,
	0x4f, 0xce, 0xcb, 0x2c, 0xc8, 0x29, 0x4d, 0xcf, 0xcc, 0xd3, 0x2f, 0x4e, 0x2d, 0x2a, 0x4b, 0x2d,
	0xd2, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0xe2, 0x84, 0x8b, 0x2b, 0x99, 0x70, 0xf1, 0x04, 0x80,
	0x59, 0x01, 0x89, 0x45, 0x89, 0xb9, 0xc5, 0x42, 0x2a, 0x5c, 0xbc, 0xc9, 0xf9, 0x79, 0x25, 0x89,
	0x99, 0x79, 0xa9, 0x45, 0x7e, 0x89, 0xb9, 0xa9, 0x12, 0x8c, 0x0a, 0x4c, 0x1a, 0x9c, 0x41, 0xa8,
	0x82, 0x4a, 0x0a, 0x30, 0x5d, 0x41, 0xa9, 0xc5, 0xa5, 0x39, 0x25, 0x42, 0x02, 0x5c, 0xcc, 0xb9,
	0xc5, 0xe9, 0x12, 0x4c, 0x0a, 0x8c, 0x1a, 0x9c, 0x41, 0x20, 0xa6, 0x51, 0x0f, 0x23, 0x17, 0x1b,
	0x44, 0x89, 0x90, 0x03, 0x17, 0x97, 0x63, 0x4a, 0x8a, 0x5f, 0x6a, 0x49, 0x79, 0x7e, 0x51, 0xb6,
	0x90, 0xb8, 0x1e, 0xdc, 0x72, 0x3d, 0x64, 0x9b, 0xa5, 0x30, 0x25, 0x20, 0x86, 0x2b, 0x31, 0x80,
	0x4c, 0x70, 0x49, 0xcd, 0xa1, 0xc0, 0x04, 0x40, 0x00, 0x00, 0x00, 0xff, 0xff, 0x26, 0x06, 0x93,
	0x0f, 0x0e, 0x01, 0x00, 0x00,
}