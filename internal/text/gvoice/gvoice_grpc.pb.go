// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package gvoice

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// GVoiceClient is the client API for GVoice service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type GVoiceClient interface {
	// Sends a message to the recipient.
	SendSMS(ctx context.Context, in *SendSMSRequest, opts ...grpc.CallOption) (*SendSMSResponse, error)
	// Gets the contact list for a given GVoice account.
	GetContactList(ctx context.Context, in *FetchContactListRequest, opts ...grpc.CallOption) (*FetchContactListResponse, error)
	// Gets all GVoice numbers that are accessible/ready to be used by the service.
	GetGVoiceNumbers(ctx context.Context, in *FetchGVoiceNumbersRequest, opts ...grpc.CallOption) (*FetchGVoiceNumbersResponse, error)
	// Fetches the text-message history between a contact on a given GVoice account.
	GetContactHistory(ctx context.Context, in *FetchContactHistoryRequest, opts ...grpc.CallOption) (*FetchContactHistoryResponse, error)
}

type gVoiceClient struct {
	cc grpc.ClientConnInterface
}

func NewGVoiceClient(cc grpc.ClientConnInterface) GVoiceClient {
	return &gVoiceClient{cc}
}

func (c *gVoiceClient) SendSMS(ctx context.Context, in *SendSMSRequest, opts ...grpc.CallOption) (*SendSMSResponse, error) {
	out := new(SendSMSResponse)
	err := c.cc.Invoke(ctx, "/GVoice/SendSMS", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *gVoiceClient) GetContactList(ctx context.Context, in *FetchContactListRequest, opts ...grpc.CallOption) (*FetchContactListResponse, error) {
	out := new(FetchContactListResponse)
	err := c.cc.Invoke(ctx, "/GVoice/GetContactList", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *gVoiceClient) GetGVoiceNumbers(ctx context.Context, in *FetchGVoiceNumbersRequest, opts ...grpc.CallOption) (*FetchGVoiceNumbersResponse, error) {
	out := new(FetchGVoiceNumbersResponse)
	err := c.cc.Invoke(ctx, "/GVoice/GetGVoiceNumbers", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *gVoiceClient) GetContactHistory(ctx context.Context, in *FetchContactHistoryRequest, opts ...grpc.CallOption) (*FetchContactHistoryResponse, error) {
	out := new(FetchContactHistoryResponse)
	err := c.cc.Invoke(ctx, "/GVoice/GetContactHistory", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// GVoiceServer is the server API for GVoice service.
// All implementations must embed UnimplementedGVoiceServer
// for forward compatibility
type GVoiceServer interface {
	// Sends a message to the recipient.
	SendSMS(context.Context, *SendSMSRequest) (*SendSMSResponse, error)
	// Gets the contact list for a given GVoice account.
	GetContactList(context.Context, *FetchContactListRequest) (*FetchContactListResponse, error)
	// Gets all GVoice numbers that are accessible/ready to be used by the service.
	GetGVoiceNumbers(context.Context, *FetchGVoiceNumbersRequest) (*FetchGVoiceNumbersResponse, error)
	// Fetches the text-message history between a contact on a given GVoice account.
	GetContactHistory(context.Context, *FetchContactHistoryRequest) (*FetchContactHistoryResponse, error)
	mustEmbedUnimplementedGVoiceServer()
}

// UnimplementedGVoiceServer must be embedded to have forward compatible implementations.
type UnimplementedGVoiceServer struct {
}

func (UnimplementedGVoiceServer) SendSMS(context.Context, *SendSMSRequest) (*SendSMSResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendSMS not implemented")
}
func (UnimplementedGVoiceServer) GetContactList(context.Context, *FetchContactListRequest) (*FetchContactListResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetContactList not implemented")
}
func (UnimplementedGVoiceServer) GetGVoiceNumbers(context.Context, *FetchGVoiceNumbersRequest) (*FetchGVoiceNumbersResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetGVoiceNumbers not implemented")
}
func (UnimplementedGVoiceServer) GetContactHistory(context.Context, *FetchContactHistoryRequest) (*FetchContactHistoryResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetContactHistory not implemented")
}
func (UnimplementedGVoiceServer) mustEmbedUnimplementedGVoiceServer() {}

// UnsafeGVoiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to GVoiceServer will
// result in compilation errors.
type UnsafeGVoiceServer interface {
	mustEmbedUnimplementedGVoiceServer()
}

func RegisterGVoiceServer(s grpc.ServiceRegistrar, srv GVoiceServer) {
	s.RegisterService(&GVoice_ServiceDesc, srv)
}

func _GVoice_SendSMS_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SendSMSRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GVoiceServer).SendSMS(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/GVoice/SendSMS",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GVoiceServer).SendSMS(ctx, req.(*SendSMSRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GVoice_GetContactList_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(FetchContactListRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GVoiceServer).GetContactList(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/GVoice/GetContactList",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GVoiceServer).GetContactList(ctx, req.(*FetchContactListRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GVoice_GetGVoiceNumbers_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(FetchGVoiceNumbersRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GVoiceServer).GetGVoiceNumbers(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/GVoice/GetGVoiceNumbers",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GVoiceServer).GetGVoiceNumbers(ctx, req.(*FetchGVoiceNumbersRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GVoice_GetContactHistory_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(FetchContactHistoryRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GVoiceServer).GetContactHistory(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/GVoice/GetContactHistory",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GVoiceServer).GetContactHistory(ctx, req.(*FetchContactHistoryRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// GVoice_ServiceDesc is the grpc.ServiceDesc for GVoice service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var GVoice_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "GVoice",
	HandlerType: (*GVoiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SendSMS",
			Handler:    _GVoice_SendSMS_Handler,
		},
		{
			MethodName: "GetContactList",
			Handler:    _GVoice_GetContactList_Handler,
		},
		{
			MethodName: "GetGVoiceNumbers",
			Handler:    _GVoice_GetGVoiceNumbers_Handler,
		},
		{
			MethodName: "GetContactHistory",
			Handler:    _GVoice_GetContactHistory_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "gvoice.proto",
}
