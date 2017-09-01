package api

import (
	"os"
	"time"

	"google.golang.org/grpc"

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
			HostID:             user.GetHostID(),
		})
	}

	return &pb.UserResponse{Users: ut}, nil
}

func (s *UserService) Create(ctx context.Context, req *pb.UserCreateRequest) (*pb.UserResponse, error) {
	logrus.Debugf("rpc call: user create: %s", req.Username)
	var ut []*pb.UserResponse_User
	user, err := ovpm.CreateNewUser(req.Username, req.Password, req.NoGW, req.HostID)
	if err != nil {
		return nil, err
	}

	pbUser := pb.UserResponse_User{
		Username:           user.GetUsername(),
		ServerSerialNumber: user.GetServerSerialNumber(),
		NoGW:               user.IsNoGW(),
		HostID:             user.GetHostID(),
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
		noGW = true
	case pb.UserUpdateRequest_GW:
		noGW = false
	default:
		noGW = user.NoGW

	}

	err = user.Update(req.Password, noGW, req.HostID)
	if err != nil {
		return nil, err
	}
	pbUser := pb.UserResponse_User{
		Username:           user.GetUsername(),
		ServerSerialNumber: user.GetServerSerialNumber(),
		NoGW:               user.IsNoGW(),
		HostID:             user.GetHostID(),
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
		HostID:             user.GetHostID(),
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
		HostID:             user.GetHostID(),
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
		Proto:        server.GetProto(),
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
	var proto string
	switch req.Protopref {
	case pb.VPNProto_TCP:
		proto = ovpm.TCPProto
	case pb.VPNProto_UDP:
		proto = ovpm.UDPProto
	case pb.VPNProto_NOPREF:
		proto = ovpm.UDPProto
	}

	if err := ovpm.Init(req.Hostname, req.Port, proto, req.IPBlock); err != nil {
		logrus.Errorf("server can not be created: %v", err)
	}
	return &pb.VPNInitResponse{}, nil
}

type NetworkService struct{}

func (s *NetworkService) List(ctx context.Context, req *pb.NetworkListRequest) (*pb.NetworkListResponse, error) {
	logrus.Debug("rpc call: network list")
	var nt []*pb.Network

	networks := ovpm.GetAllNetworks()
	for _, network := range networks {
		nt = append(nt, &pb.Network{
			Name:                network.GetName(),
			CIDR:                network.GetCIDR(),
			Type:                network.GetType().String(),
			CreatedAt:           network.GetCreatedAt(),
			AssociatedUsernames: network.GetAssociatedUsernames(),
			Via:                 network.GetVia(),
		})
	}

	return &pb.NetworkListResponse{Networks: nt}, nil
}

func (s *NetworkService) Create(ctx context.Context, req *pb.NetworkCreateRequest) (*pb.NetworkCreateResponse, error) {
	logrus.Debugf("rpc call: network create: %s", req.Name)
	network, err := ovpm.CreateNewNetwork(req.Name, req.CIDR, ovpm.NetworkTypeFromString(req.Type), req.Via)
	if err != nil {
		return nil, err
	}

	n := pb.Network{
		Name:                network.GetName(),
		CIDR:                network.GetCIDR(),
		Type:                network.GetType().String(),
		CreatedAt:           network.GetCreatedAt(),
		AssociatedUsernames: network.GetAssociatedUsernames(),
		Via:                 network.GetVia(),
	}

	return &pb.NetworkCreateResponse{Network: &n}, nil
}

func (s *NetworkService) Delete(ctx context.Context, req *pb.NetworkDeleteRequest) (*pb.NetworkDeleteResponse, error) {
	logrus.Debugf("rpc call: network delete: %s", req.Name)
	network, err := ovpm.GetNetwork(req.Name)
	if err != nil {
		return nil, err
	}

	err = network.Delete()
	if err != nil {
		return nil, err
	}

	n := pb.Network{
		Name:                network.GetName(),
		CIDR:                network.GetCIDR(),
		Type:                network.GetType().String(),
		CreatedAt:           network.GetCreatedAt(),
		AssociatedUsernames: network.GetAssociatedUsernames(),
		Via:                 network.GetVia(),
	}

	return &pb.NetworkDeleteResponse{Network: &n}, nil
}

func (s *NetworkService) GetAllTypes(ctx context.Context, req *pb.NetworkGetAllTypesRequest) (*pb.NetworkGetAllTypesResponse, error) {
	logrus.Debugf("rpc call: network get-types")
	var networkTypes []*pb.NetworkType
	for _, nt := range ovpm.GetAllNetworkTypes() {
		if nt == ovpm.UNDEFINEDNET {
			continue
		}
		networkTypes = append(networkTypes, &pb.NetworkType{Type: nt.String(), Description: nt.Description()})
	}

	return &pb.NetworkGetAllTypesResponse{Types: networkTypes}, nil
}

func (s *NetworkService) GetAssociatedUsers(ctx context.Context, req *pb.NetworkGetAssociatedUsersRequest) (*pb.NetworkGetAssociatedUsersResponse, error) {
	logrus.Debugf("rpc call: network get-associated-users")
	network, err := ovpm.GetNetwork(req.Name)
	if err != nil {
		return nil, err
	}
	usernames := network.GetAssociatedUsernames()
	return &pb.NetworkGetAssociatedUsersResponse{Usernames: usernames}, nil
}

func (s *NetworkService) Associate(ctx context.Context, req *pb.NetworkAssociateRequest) (*pb.NetworkAssociateResponse, error) {
	logrus.Debugf("rpc call: network associate")

	network, err := ovpm.GetNetwork(req.Name)
	if err != nil {
		return nil, err
	}
	err = network.Associate(req.Username)
	if err != nil {
		return nil, err
	}

	return &pb.NetworkAssociateResponse{}, nil
}

func (s *NetworkService) Dissociate(ctx context.Context, req *pb.NetworkDissociateRequest) (*pb.NetworkDissociateResponse, error) {
	logrus.Debugf("rpc call: network dissociate")

	network, err := ovpm.GetNetwork(req.Name)
	if err != nil {
		return nil, err
	}

	err = network.Dissociate(req.Username)
	if err != nil {
		return nil, err
	}

	return &pb.NetworkDissociateResponse{}, nil
}

// NewRPCServer returns a new gRPC server.
func NewRPCServer() *grpc.Server {
	s := grpc.NewServer()
	pb.RegisterUserServiceServer(s, &UserService{})
	pb.RegisterVPNServiceServer(s, &VPNService{})
	pb.RegisterNetworkServiceServer(s, &NetworkService{})
	return s
}
