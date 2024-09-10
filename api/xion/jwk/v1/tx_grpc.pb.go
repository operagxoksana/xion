// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             (unknown)
// source: xion/jwk/v1/tx.proto

package jwkv1

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

const (
	Msg_CreateAudienceClaim_FullMethodName = "/xion.jwk.v1.Msg/CreateAudienceClaim"
	Msg_DeleteAudienceClaim_FullMethodName = "/xion.jwk.v1.Msg/DeleteAudienceClaim"
	Msg_CreateAudience_FullMethodName      = "/xion.jwk.v1.Msg/CreateAudience"
	Msg_UpdateAudience_FullMethodName      = "/xion.jwk.v1.Msg/UpdateAudience"
	Msg_DeleteAudience_FullMethodName      = "/xion.jwk.v1.Msg/DeleteAudience"
)

// MsgClient is the client API for Msg service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type MsgClient interface {
	CreateAudienceClaim(ctx context.Context, in *MsgCreateAudienceClaim, opts ...grpc.CallOption) (*MsgCreateAudienceClaimResponse, error)
	DeleteAudienceClaim(ctx context.Context, in *MsgDeleteAudienceClaim, opts ...grpc.CallOption) (*MsgDeleteAudienceClaimResponse, error)
	CreateAudience(ctx context.Context, in *MsgCreateAudience, opts ...grpc.CallOption) (*MsgCreateAudienceResponse, error)
	UpdateAudience(ctx context.Context, in *MsgUpdateAudience, opts ...grpc.CallOption) (*MsgUpdateAudienceResponse, error)
	DeleteAudience(ctx context.Context, in *MsgDeleteAudience, opts ...grpc.CallOption) (*MsgDeleteAudienceResponse, error)
}

type msgClient struct {
	cc grpc.ClientConnInterface
}

func NewMsgClient(cc grpc.ClientConnInterface) MsgClient {
	return &msgClient{cc}
}

func (c *msgClient) CreateAudienceClaim(ctx context.Context, in *MsgCreateAudienceClaim, opts ...grpc.CallOption) (*MsgCreateAudienceClaimResponse, error) {
	out := new(MsgCreateAudienceClaimResponse)
	err := c.cc.Invoke(ctx, Msg_CreateAudienceClaim_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) DeleteAudienceClaim(ctx context.Context, in *MsgDeleteAudienceClaim, opts ...grpc.CallOption) (*MsgDeleteAudienceClaimResponse, error) {
	out := new(MsgDeleteAudienceClaimResponse)
	err := c.cc.Invoke(ctx, Msg_DeleteAudienceClaim_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) CreateAudience(ctx context.Context, in *MsgCreateAudience, opts ...grpc.CallOption) (*MsgCreateAudienceResponse, error) {
	out := new(MsgCreateAudienceResponse)
	err := c.cc.Invoke(ctx, Msg_CreateAudience_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) UpdateAudience(ctx context.Context, in *MsgUpdateAudience, opts ...grpc.CallOption) (*MsgUpdateAudienceResponse, error) {
	out := new(MsgUpdateAudienceResponse)
	err := c.cc.Invoke(ctx, Msg_UpdateAudience_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) DeleteAudience(ctx context.Context, in *MsgDeleteAudience, opts ...grpc.CallOption) (*MsgDeleteAudienceResponse, error) {
	out := new(MsgDeleteAudienceResponse)
	err := c.cc.Invoke(ctx, Msg_DeleteAudience_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MsgServer is the server API for Msg service.
// All implementations must embed UnimplementedMsgServer
// for forward compatibility
type MsgServer interface {
	CreateAudienceClaim(context.Context, *MsgCreateAudienceClaim) (*MsgCreateAudienceClaimResponse, error)
	DeleteAudienceClaim(context.Context, *MsgDeleteAudienceClaim) (*MsgDeleteAudienceClaimResponse, error)
	CreateAudience(context.Context, *MsgCreateAudience) (*MsgCreateAudienceResponse, error)
	UpdateAudience(context.Context, *MsgUpdateAudience) (*MsgUpdateAudienceResponse, error)
	DeleteAudience(context.Context, *MsgDeleteAudience) (*MsgDeleteAudienceResponse, error)
	mustEmbedUnimplementedMsgServer()
}

// UnimplementedMsgServer must be embedded to have forward compatible implementations.
type UnimplementedMsgServer struct {
}

func (UnimplementedMsgServer) CreateAudienceClaim(context.Context, *MsgCreateAudienceClaim) (*MsgCreateAudienceClaimResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateAudienceClaim not implemented")
}
func (UnimplementedMsgServer) DeleteAudienceClaim(context.Context, *MsgDeleteAudienceClaim) (*MsgDeleteAudienceClaimResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteAudienceClaim not implemented")
}
func (UnimplementedMsgServer) CreateAudience(context.Context, *MsgCreateAudience) (*MsgCreateAudienceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateAudience not implemented")
}
func (UnimplementedMsgServer) UpdateAudience(context.Context, *MsgUpdateAudience) (*MsgUpdateAudienceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateAudience not implemented")
}
func (UnimplementedMsgServer) DeleteAudience(context.Context, *MsgDeleteAudience) (*MsgDeleteAudienceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteAudience not implemented")
}
func (UnimplementedMsgServer) mustEmbedUnimplementedMsgServer() {}

// UnsafeMsgServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to MsgServer will
// result in compilation errors.
type UnsafeMsgServer interface {
	mustEmbedUnimplementedMsgServer()
}

func RegisterMsgServer(s grpc.ServiceRegistrar, srv MsgServer) {
	s.RegisterService(&Msg_ServiceDesc, srv)
}

func _Msg_CreateAudienceClaim_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgCreateAudienceClaim)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).CreateAudienceClaim(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Msg_CreateAudienceClaim_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).CreateAudienceClaim(ctx, req.(*MsgCreateAudienceClaim))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_DeleteAudienceClaim_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgDeleteAudienceClaim)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).DeleteAudienceClaim(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Msg_DeleteAudienceClaim_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).DeleteAudienceClaim(ctx, req.(*MsgDeleteAudienceClaim))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_CreateAudience_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgCreateAudience)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).CreateAudience(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Msg_CreateAudience_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).CreateAudience(ctx, req.(*MsgCreateAudience))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_UpdateAudience_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgUpdateAudience)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).UpdateAudience(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Msg_UpdateAudience_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).UpdateAudience(ctx, req.(*MsgUpdateAudience))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_DeleteAudience_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgDeleteAudience)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).DeleteAudience(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Msg_DeleteAudience_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).DeleteAudience(ctx, req.(*MsgDeleteAudience))
	}
	return interceptor(ctx, in, info, handler)
}

// Msg_ServiceDesc is the grpc.ServiceDesc for Msg service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Msg_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "xion.jwk.v1.Msg",
	HandlerType: (*MsgServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateAudienceClaim",
			Handler:    _Msg_CreateAudienceClaim_Handler,
		},
		{
			MethodName: "DeleteAudienceClaim",
			Handler:    _Msg_DeleteAudienceClaim_Handler,
		},
		{
			MethodName: "CreateAudience",
			Handler:    _Msg_CreateAudience_Handler,
		},
		{
			MethodName: "UpdateAudience",
			Handler:    _Msg_UpdateAudience_Handler,
		},
		{
			MethodName: "DeleteAudience",
			Handler:    _Msg_DeleteAudience_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "xion/jwk/v1/tx.proto",
}
