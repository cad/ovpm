// Code generated by protoc-gen-go. DO NOT EDIT.
// source: user.proto

/*
Package pb is a generated protocol buffer package.

It is generated from these files:
	user.proto
	vpn.proto

It has these top-level messages:
	UserListRequest
	UserCreateRequest
	UserDeleteRequest
	UserRenewRequest
	UserGenConfigRequest
	UserResponse
	UserGenConfigResponse
	VPNStatusRequest
	VPNInitRequest
	VPNApplyRequest
	VPNStatusResponse
	VPNInitResponse
	VPNApplyResponse
*/
package pb

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

type UserListRequest struct {
}

func (m *UserListRequest) Reset()                    { *m = UserListRequest{} }
func (m *UserListRequest) String() string            { return proto.CompactTextString(m) }
func (*UserListRequest) ProtoMessage()               {}
func (*UserListRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

type UserCreateRequest struct {
	Username string `protobuf:"bytes,1,opt,name=Username" json:"Username,omitempty"`
	Password string `protobuf:"bytes,2,opt,name=Password" json:"Password,omitempty"`
}

func (m *UserCreateRequest) Reset()                    { *m = UserCreateRequest{} }
func (m *UserCreateRequest) String() string            { return proto.CompactTextString(m) }
func (*UserCreateRequest) ProtoMessage()               {}
func (*UserCreateRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *UserCreateRequest) GetUsername() string {
	if m != nil {
		return m.Username
	}
	return ""
}

func (m *UserCreateRequest) GetPassword() string {
	if m != nil {
		return m.Password
	}
	return ""
}

type UserDeleteRequest struct {
	Username string `protobuf:"bytes,1,opt,name=Username" json:"Username,omitempty"`
}

func (m *UserDeleteRequest) Reset()                    { *m = UserDeleteRequest{} }
func (m *UserDeleteRequest) String() string            { return proto.CompactTextString(m) }
func (*UserDeleteRequest) ProtoMessage()               {}
func (*UserDeleteRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *UserDeleteRequest) GetUsername() string {
	if m != nil {
		return m.Username
	}
	return ""
}

type UserRenewRequest struct {
	Username string `protobuf:"bytes,1,opt,name=Username" json:"Username,omitempty"`
}

func (m *UserRenewRequest) Reset()                    { *m = UserRenewRequest{} }
func (m *UserRenewRequest) String() string            { return proto.CompactTextString(m) }
func (*UserRenewRequest) ProtoMessage()               {}
func (*UserRenewRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *UserRenewRequest) GetUsername() string {
	if m != nil {
		return m.Username
	}
	return ""
}

type UserGenConfigRequest struct {
	Username string `protobuf:"bytes,1,opt,name=Username" json:"Username,omitempty"`
}

func (m *UserGenConfigRequest) Reset()                    { *m = UserGenConfigRequest{} }
func (m *UserGenConfigRequest) String() string            { return proto.CompactTextString(m) }
func (*UserGenConfigRequest) ProtoMessage()               {}
func (*UserGenConfigRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

func (m *UserGenConfigRequest) GetUsername() string {
	if m != nil {
		return m.Username
	}
	return ""
}

type UserResponse struct {
	Users []*UserResponse_User `protobuf:"bytes,1,rep,name=users" json:"users,omitempty"`
}

func (m *UserResponse) Reset()                    { *m = UserResponse{} }
func (m *UserResponse) String() string            { return proto.CompactTextString(m) }
func (*UserResponse) ProtoMessage()               {}
func (*UserResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

func (m *UserResponse) GetUsers() []*UserResponse_User {
	if m != nil {
		return m.Users
	}
	return nil
}

type UserResponse_User struct {
	Username           string `protobuf:"bytes,1,opt,name=Username" json:"Username,omitempty"`
	ServerSerialNumber string `protobuf:"bytes,2,opt,name=ServerSerialNumber" json:"ServerSerialNumber,omitempty"`
	Cert               string `protobuf:"bytes,3,opt,name=Cert" json:"Cert,omitempty"`
	CreatedAt          string `protobuf:"bytes,4,opt,name=CreatedAt" json:"CreatedAt,omitempty"`
}

func (m *UserResponse_User) Reset()                    { *m = UserResponse_User{} }
func (m *UserResponse_User) String() string            { return proto.CompactTextString(m) }
func (*UserResponse_User) ProtoMessage()               {}
func (*UserResponse_User) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5, 0} }

func (m *UserResponse_User) GetUsername() string {
	if m != nil {
		return m.Username
	}
	return ""
}

func (m *UserResponse_User) GetServerSerialNumber() string {
	if m != nil {
		return m.ServerSerialNumber
	}
	return ""
}

func (m *UserResponse_User) GetCert() string {
	if m != nil {
		return m.Cert
	}
	return ""
}

func (m *UserResponse_User) GetCreatedAt() string {
	if m != nil {
		return m.CreatedAt
	}
	return ""
}

type UserGenConfigResponse struct {
	ClientConfig string `protobuf:"bytes,1,opt,name=ClientConfig" json:"ClientConfig,omitempty"`
}

func (m *UserGenConfigResponse) Reset()                    { *m = UserGenConfigResponse{} }
func (m *UserGenConfigResponse) String() string            { return proto.CompactTextString(m) }
func (*UserGenConfigResponse) ProtoMessage()               {}
func (*UserGenConfigResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{6} }

func (m *UserGenConfigResponse) GetClientConfig() string {
	if m != nil {
		return m.ClientConfig
	}
	return ""
}

func init() {
	proto.RegisterType((*UserListRequest)(nil), "pb.UserListRequest")
	proto.RegisterType((*UserCreateRequest)(nil), "pb.UserCreateRequest")
	proto.RegisterType((*UserDeleteRequest)(nil), "pb.UserDeleteRequest")
	proto.RegisterType((*UserRenewRequest)(nil), "pb.UserRenewRequest")
	proto.RegisterType((*UserGenConfigRequest)(nil), "pb.UserGenConfigRequest")
	proto.RegisterType((*UserResponse)(nil), "pb.UserResponse")
	proto.RegisterType((*UserResponse_User)(nil), "pb.UserResponse.User")
	proto.RegisterType((*UserGenConfigResponse)(nil), "pb.UserGenConfigResponse")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for UserService service

type UserServiceClient interface {
	List(ctx context.Context, in *UserListRequest, opts ...grpc.CallOption) (*UserResponse, error)
	Create(ctx context.Context, in *UserCreateRequest, opts ...grpc.CallOption) (*UserResponse, error)
	Delete(ctx context.Context, in *UserDeleteRequest, opts ...grpc.CallOption) (*UserResponse, error)
	Renew(ctx context.Context, in *UserRenewRequest, opts ...grpc.CallOption) (*UserResponse, error)
	GenConfig(ctx context.Context, in *UserGenConfigRequest, opts ...grpc.CallOption) (*UserGenConfigResponse, error)
}

type userServiceClient struct {
	cc *grpc.ClientConn
}

func NewUserServiceClient(cc *grpc.ClientConn) UserServiceClient {
	return &userServiceClient{cc}
}

func (c *userServiceClient) List(ctx context.Context, in *UserListRequest, opts ...grpc.CallOption) (*UserResponse, error) {
	out := new(UserResponse)
	err := grpc.Invoke(ctx, "/pb.UserService/List", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userServiceClient) Create(ctx context.Context, in *UserCreateRequest, opts ...grpc.CallOption) (*UserResponse, error) {
	out := new(UserResponse)
	err := grpc.Invoke(ctx, "/pb.UserService/Create", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userServiceClient) Delete(ctx context.Context, in *UserDeleteRequest, opts ...grpc.CallOption) (*UserResponse, error) {
	out := new(UserResponse)
	err := grpc.Invoke(ctx, "/pb.UserService/Delete", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userServiceClient) Renew(ctx context.Context, in *UserRenewRequest, opts ...grpc.CallOption) (*UserResponse, error) {
	out := new(UserResponse)
	err := grpc.Invoke(ctx, "/pb.UserService/Renew", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userServiceClient) GenConfig(ctx context.Context, in *UserGenConfigRequest, opts ...grpc.CallOption) (*UserGenConfigResponse, error) {
	out := new(UserGenConfigResponse)
	err := grpc.Invoke(ctx, "/pb.UserService/GenConfig", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for UserService service

type UserServiceServer interface {
	List(context.Context, *UserListRequest) (*UserResponse, error)
	Create(context.Context, *UserCreateRequest) (*UserResponse, error)
	Delete(context.Context, *UserDeleteRequest) (*UserResponse, error)
	Renew(context.Context, *UserRenewRequest) (*UserResponse, error)
	GenConfig(context.Context, *UserGenConfigRequest) (*UserGenConfigResponse, error)
}

func RegisterUserServiceServer(s *grpc.Server, srv UserServiceServer) {
	s.RegisterService(&_UserService_serviceDesc, srv)
}

func _UserService_List_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UserListRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserServiceServer).List(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.UserService/List",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserServiceServer).List(ctx, req.(*UserListRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _UserService_Create_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UserCreateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserServiceServer).Create(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.UserService/Create",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserServiceServer).Create(ctx, req.(*UserCreateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _UserService_Delete_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UserDeleteRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserServiceServer).Delete(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.UserService/Delete",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserServiceServer).Delete(ctx, req.(*UserDeleteRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _UserService_Renew_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UserRenewRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserServiceServer).Renew(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.UserService/Renew",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserServiceServer).Renew(ctx, req.(*UserRenewRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _UserService_GenConfig_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UserGenConfigRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserServiceServer).GenConfig(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.UserService/GenConfig",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserServiceServer).GenConfig(ctx, req.(*UserGenConfigRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _UserService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "pb.UserService",
	HandlerType: (*UserServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "List",
			Handler:    _UserService_List_Handler,
		},
		{
			MethodName: "Create",
			Handler:    _UserService_Create_Handler,
		},
		{
			MethodName: "Delete",
			Handler:    _UserService_Delete_Handler,
		},
		{
			MethodName: "Renew",
			Handler:    _UserService_Renew_Handler,
		},
		{
			MethodName: "GenConfig",
			Handler:    _UserService_GenConfig_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "user.proto",
}

func init() { proto.RegisterFile("user.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 351 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x53, 0xd1, 0x4e, 0xea, 0x40,
	0x10, 0xbd, 0x85, 0x42, 0x2e, 0x03, 0x89, 0x30, 0x42, 0xb2, 0x36, 0x3e, 0x90, 0x7d, 0x22, 0x31,
	0x29, 0x11, 0x1e, 0x7d, 0xd2, 0x9a, 0xf8, 0xa0, 0x31, 0x06, 0xe2, 0x07, 0xb4, 0x32, 0x9a, 0x26,
	0xb0, 0xad, 0xbb, 0x8b, 0xfc, 0x80, 0xff, 0xe1, 0xbf, 0xf8, 0x65, 0x66, 0xbb, 0x2d, 0x05, 0x52,
	0x0d, 0x6f, 0x3b, 0xe7, 0xcc, 0x19, 0x66, 0x0e, 0xa7, 0x00, 0x6b, 0x45, 0xd2, 0x4f, 0x65, 0xa2,
	0x13, 0xac, 0xa5, 0x11, 0xef, 0xc1, 0xc9, 0xb3, 0x22, 0xf9, 0x10, 0x2b, 0x3d, 0xa3, 0xf7, 0x35,
	0x29, 0xcd, 0xef, 0xa1, 0x67, 0xa0, 0x40, 0x52, 0xa8, 0x29, 0x07, 0xd1, 0x83, 0xff, 0x06, 0x14,
	0xe1, 0x8a, 0x98, 0x33, 0x74, 0x46, 0xad, 0xd9, 0xb6, 0x36, 0xdc, 0x53, 0xa8, 0xd4, 0x26, 0x91,
	0x0b, 0x56, 0xb3, 0x5c, 0x51, 0xf3, 0xb1, 0x1d, 0x76, 0x4b, 0x4b, 0x3a, 0x6a, 0x18, 0xf7, 0xa1,
	0x6b, 0xde, 0x33, 0x12, 0xb4, 0x39, 0xa6, 0x7f, 0x02, 0x7d, 0xf3, 0xbe, 0x23, 0x11, 0x24, 0xe2,
	0x35, 0x7e, 0x3b, 0x46, 0xf3, 0xed, 0x40, 0xc7, 0xfe, 0x88, 0x4a, 0x13, 0xa1, 0x08, 0x2f, 0xa0,
	0x61, 0x7c, 0x51, 0xcc, 0x19, 0xd6, 0x47, 0xed, 0xc9, 0xc0, 0x4f, 0x23, 0x7f, 0xb7, 0xc1, 0x16,
	0xb6, 0xc7, 0xfb, 0x74, 0xc0, 0x35, 0xf5, 0x9f, 0x9e, 0xf8, 0x80, 0x73, 0x92, 0x1f, 0x24, 0xe7,
	0x24, 0xe3, 0x70, 0xf9, 0xb8, 0x5e, 0x45, 0x24, 0x73, 0x77, 0x2a, 0x18, 0x44, 0x70, 0x03, 0x92,
	0x9a, 0xd5, 0xb3, 0x8e, 0xec, 0x8d, 0xe7, 0xd0, 0xb2, 0x7f, 0xc2, 0xe2, 0x5a, 0x33, 0x37, 0x23,
	0x4a, 0x80, 0x5f, 0xc1, 0xe0, 0xe0, 0xf0, 0xfc, 0x18, 0x0e, 0x9d, 0x60, 0x19, 0x93, 0xd0, 0x16,
	0xcf, 0x57, 0xdb, 0xc3, 0x26, 0x5f, 0x35, 0x68, 0x1b, 0xb5, 0xd9, 0x24, 0x7e, 0x21, 0x1c, 0x83,
	0x6b, 0x22, 0x80, 0xa7, 0xc5, 0xe5, 0x3b, 0x81, 0xf0, 0xba, 0x87, 0x76, 0xf0, 0x7f, 0x38, 0x85,
	0xa6, 0x5d, 0x05, 0xb7, 0x66, 0xed, 0x05, 0xe6, 0x37, 0x91, 0x0d, 0x42, 0x29, 0xda, 0x0b, 0x46,
	0xa5, 0xe8, 0x12, 0x1a, 0x59, 0x18, 0xb0, 0x5f, 0x92, 0x65, 0x36, 0x2a, 0x25, 0x37, 0xd0, 0xda,
	0xda, 0x82, 0xac, 0x68, 0x38, 0x8c, 0x88, 0x77, 0x56, 0xc1, 0x14, 0x33, 0xa2, 0x66, 0xf6, 0x8d,
	0x4c, 0x7f, 0x02, 0x00, 0x00, 0xff, 0xff, 0x31, 0xff, 0x3d, 0x2c, 0x31, 0x03, 0x00, 0x00,
}
