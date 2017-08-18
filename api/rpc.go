package api

import (
	"os"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/cad/ovpm"
	"github.com/cad/ovpm/pb"
	"golang.org/x/net/context"
)

type UserService struct{}

func (s *UserService) List(ctx context.Context, req *pb.UserListRequest) (*pb.UserResponse, error) {
	logrus.Debug("rpc call: user list")
	var ut []*pb.UserResponse_User

	users, err := ovpm.GetAllUsers()
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
			IPNet:              user.GetIPNet(),
			NoGW:               user.IsNoGW(),
		})
	}

	return &pb.UserResponse{Users: ut}, nil
}

func (s *UserService) Create(ctx context.Context, req *pb.UserCreateRequest) (*pb.UserResponse, error) {
	logrus.Debugf("rpc call: user create: %s", req.Username)
	var ut []*pb.UserResponse_User
	user, err := ovpm.CreateNewUser(req.Username, req.Password, req.NoGW)
	if err != nil {
		return nil, err
	}

	pbUser := pb.UserResponse_User{
		Username:           user.GetUsername(),
		ServerSerialNumber: user.GetServerSerialNumber(),
		NoGW:               user.IsNoGW(),
	}
	ut = append(ut, &pbUser)

	return &pb.UserResponse{Users: ut}, nil
}

func (s *UserService) Update(ctx context.Context, req *pb.UserUpdateRequest) (*pb.UserResponse, error) {
	logrus.Debugf("rpc call: user update: %s", req.Username)
	var ut []*pb.UserResponse_User
	user, err := ovpm.GetUser(req.Username)
	if err != nil {
		return nil, err
	}
	var noGW bool

	switch req.Gwpref {
	case pb.UserUpdateRequest_NOGW:
		noGW = false
	case pb.UserUpdateRequest_GW:
		noGW = true
	default:
		noGW = user.NoGW

	}

	user.Update(req.Password, noGW)
	pbUser := pb.UserResponse_User{
		Username:           user.GetUsername(),
		ServerSerialNumber: user.GetServerSerialNumber(),
		NoGW:               user.IsNoGW(),
	}

	ut = append(ut, &pbUser)

	return &pb.UserResponse{Users: ut}, nil
}

func (s *UserService) Delete(ctx context.Context, req *pb.UserDeleteRequest) (*pb.UserResponse, error) {
	logrus.Debugf("rpc call: user delete: %s", req.Username)
	var ut []*pb.UserResponse_User
	user, err := ovpm.GetUser(req.Username)
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

func (s *UserService) Renew(ctx context.Context, req *pb.UserRenewRequest) (*pb.UserResponse, error) {
	logrus.Debugf("rpc call: user renew cert: %s", req.Username)
	var ut []*pb.UserResponse_User
	user, err := ovpm.GetUser(req.Username)
	if err != nil {
		return nil, err
	}

	pbUser := pb.UserResponse_User{
		Username:           user.GetUsername(),
		ServerSerialNumber: user.GetServerSerialNumber(),
	}
	ut = append(ut, &pbUser)

	err = user.Renew()
	if err != nil {
		return nil, err
	}

	return &pb.UserResponse{Users: ut}, nil
}

func (s *UserService) GenConfig(ctx context.Context, req *pb.UserGenConfigRequest) (*pb.UserGenConfigResponse, error) {
	logrus.Debugf("rpc call: user genconfig: %s", req.Username)
	user, err := ovpm.GetUser(req.Username)
	if err != nil {
		return nil, err
	}
	configBlob, err := ovpm.DumpsClientConfig(user.GetUsername())
	if err != nil {
		return nil, err
	}

	return &pb.UserGenConfigResponse{ClientConfig: configBlob}, nil
}

type VPNService struct{}

func (s *VPNService) Status(ctx context.Context, req *pb.VPNStatusRequest) (*pb.VPNStatusResponse, error) {
	logrus.Debugf("rpc call: vpn status")
	server, err := ovpm.GetServerInstance()
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

func (s *VPNService) Init(ctx context.Context, req *pb.VPNInitRequest) (*pb.VPNInitResponse, error) {
	logrus.Debugf("rpc call: vpn init")
	if err := ovpm.Init(req.Hostname, req.Port); err != nil {
		logrus.Errorf("server can not be created: %v", err)
	}
	return &pb.VPNInitResponse{}, nil
}
