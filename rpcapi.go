//go:generate protoc -I pb/ pb/user.proto pb/vpn.proto --go_out=plugins=grpc:pb

package ovpm

import (
	"os"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/cad/ovpm/pb"
	"golang.org/x/net/context"
)

type UserSvc struct{}

func (s *UserSvc) List(ctx context.Context, req *pb.UserListRequest) (*pb.UserResponse, error) {
	var ut []*pb.UserResponse_User

	users, err := GetAllUsers()
	if err != nil {
		logrus.Errorf("users can not be fetched: %v", err)
		os.Exit(1)
		return nil, err
	}
	for _, user := range users {
		ut = append(ut, &pb.UserResponse_User{
			ServerSerialNumber: user.GetServerSerialNumber(),
			Username:           user.GetUsername(),
			CreatedAt:          user.GetCreatedAt(),
		})
	}

	return &pb.UserResponse{Users: ut}, nil
}

func (s *UserSvc) Create(ctx context.Context, req *pb.UserCreateRequest) (*pb.UserResponse, error) {
	var ut []*pb.UserResponse_User
	user, err := CreateNewUser(req.Username, req.Password)
	if err != nil {
		return nil, err
	}

	pbUser := pb.UserResponse_User{
		Username:           user.GetUsername(),
		ServerSerialNumber: user.GetServerSerialNumber(),
	}
	ut = append(ut, &pbUser)

	return &pb.UserResponse{Users: ut}, nil
}

func (s *UserSvc) Delete(ctx context.Context, req *pb.UserDeleteRequest) (*pb.UserResponse, error) {
	var ut []*pb.UserResponse_User
	user, err := GetUser(req.Username)
	if err != nil {
		return nil, err
	}

	pbUser := pb.UserResponse_User{
		Username:           user.GetUsername(),
		ServerSerialNumber: user.GetServerSerialNumber(),
	}
	ut = append(ut, &pbUser)

	err = user.Delete()
	if err != nil {
		return nil, err
	}

	return &pb.UserResponse{Users: ut}, nil
}

func (s *UserSvc) Renew(ctx context.Context, req *pb.UserRenewRequest) (*pb.UserResponse, error) {
	var ut []*pb.UserResponse_User
	user, err := GetUser(req.Username)
	if err != nil {
		return nil, err
	}

	pbUser := pb.UserResponse_User{
		Username:           user.GetUsername(),
		ServerSerialNumber: user.GetServerSerialNumber(),
	}
	ut = append(ut, &pbUser)

	err = user.Sign()
	if err != nil {
		return nil, err
	}

	return &pb.UserResponse{Users: ut}, nil
}

func (s *UserSvc) GenConfig(ctx context.Context, req *pb.UserGenConfigRequest) (*pb.UserGenConfigResponse, error) {
	user, err := GetUser(req.Username)
	if err != nil {
		return nil, err
	}
	configBlob, err := sDumpUserOVPNConf(user.GetUsername())
	if err != nil {
		return nil, err
	}

	return &pb.UserGenConfigResponse{ClientConfig: configBlob}, nil
}

type VPNSvc struct{}

func (s *VPNSvc) Status(ctx context.Context, req *pb.VPNStatusRequest) (*pb.VPNStatusResponse, error) {
	server, err := GetServerInstance()
	if err != nil {
		return nil, err
	}

	response := pb.VPNStatusResponse{
		Name:         server.Name,
		SerialNumber: server.SerialNumber,
		Hostname:     server.Hostname,
		Port:         server.Port,
		Cert:         server.Cert,
		CACert:       server.CACert,
		Net:          server.Net,
		Mask:         server.Mask,
		CreatedAt:    server.CreatedAt.Format(time.UnixDate),
	}
	return &response, nil
}

func (s *VPNSvc) Init(ctx context.Context, req *pb.VPNInitRequest) (*pb.VPNInitResponse, error) {
	if err := InitServer("default", req.Hostname, req.Port); err != nil {
		logrus.Errorf("server can not be created: %v", err)
	}
	return &pb.VPNInitResponse{}, nil
}

func (s *VPNSvc) Apply(ctx context.Context, req *pb.VPNApplyRequest) (*pb.VPNApplyResponse, error) {
	if err := Emit(); err != nil {
		logrus.Errorf("can not apply configuration: %v", err)
		return nil, err
	}
	logrus.Info("changes applied")
	return &pb.VPNApplyResponse{}, nil
}
