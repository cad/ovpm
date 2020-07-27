package api

import (
	"go.uber.org/thriftrw/ptr"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/cad/ovpm"
	"github.com/cad/ovpm/api/pb"
	"github.com/cad/ovpm/permset"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

type AuthService struct{}

func (s *AuthService) Status(ctx context.Context, req *pb.AuthStatusRequest) (*pb.AuthStatusResponse, error) {
	logrus.Debug("rpc call: auth status")

	username, err := GetUsernameFromContext(ctx)
	if err != nil {
		logrus.Debugln(err)
		return nil, grpc.Errorf(codes.Unauthenticated, "username not found with the provided credentials")
	}

	if username == "root" {
		userResp := pb.UserResponse_User{
			Username: username,
			IsAdmin:  true,
		}
		return &pb.AuthStatusResponse{User: &userResp, IsRoot: true}, nil
	}
	user, err := ovpm.GetUser(username)
	if err != nil {
		logrus.Debugln(err)
		return nil, grpc.Errorf(codes.Unauthenticated, "user not found with the provided credentials")
	}

	userResp := pb.UserResponse_User{
		Username: user.GetUsername(),
		IsAdmin:  user.IsAdmin(),
	}
	return &pb.AuthStatusResponse{User: &userResp}, nil
}

func (s *AuthService) Authenticate(ctx context.Context, req *pb.AuthAuthenticateRequest) (*pb.AuthAuthenticateResponse, error) {
	logrus.Debug("rpc call: auth authenticate")

	user, err := ovpm.GetUser(req.Username)
	if err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "user not found with the provided credentials")
	}
	if !user.CheckPassword(req.Password) {
		return nil, grpc.Errorf(codes.Unauthenticated, "user not found with the provided credentials")
	}

	token, err := user.RenewToken()
	if err != nil {
		return nil, grpc.Errorf(codes.Internal, "token can not be generated")
	}

	return &pb.AuthAuthenticateResponse{Token: token}, nil
}

type UserService struct{}

func (s *UserService) List(ctx context.Context, req *pb.UserListRequest) (*pb.UserResponse, error) {
	logrus.Debug("rpc call: user list")
	perms, err := permset.FromContext(ctx)
	if err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "permset not found within the context")
	}

	// Check perms.
	if !perms.Contains(ovpm.GetAnyUserPerm) {
		return nil, grpc.Errorf(codes.PermissionDenied, "ovpm.GetAnyUserPerm is required for this operation")
	}

	var ut []*pb.UserResponse_User

	users, err := ovpm.GetAllUsers()
	if err != nil {
		logrus.Errorf("users can not be fetched: %v", err)
		os.Exit(1)
		return nil, err
	}
	for _, user := range users {
		isConnected, connectedSince, bytesSent, bytesReceived := user.ConnectionStatus()
		ut = append(ut, &pb.UserResponse_User{
			ServerSerialNumber: user.GetServerSerialNumber(),
			Username:           user.GetUsername(),
			CreatedAt:          user.GetCreatedAt(),
			IpNet:              user.GetIPNet(),
			NoGw:               user.IsNoGW(),
			HostId:             user.GetHostID(),
			IsAdmin:            user.IsAdmin(),
			IsConnected:        isConnected,
			ConnectedSince:     connectedSince.UTC().Format(time.RFC3339),
			BytesSent:          bytesSent,
			BytesReceived:      bytesReceived,
			ExpiresAt:          user.ExpiresAt().UTC().Format(time.RFC3339),
			Description:        user.GetDescription(),
		})
	}

	return &pb.UserResponse{Users: ut}, nil
}

func (s *UserService) Create(ctx context.Context, req *pb.UserCreateRequest) (*pb.UserResponse, error) {
	logrus.Debugf("rpc call: user create: %s", req.Username)
	perms, err := permset.FromContext(ctx)
	if err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "permset not found within the context")
	}

	// Check perms.
	if !perms.Contains(ovpm.CreateUserPerm) {
		return nil, grpc.Errorf(codes.PermissionDenied, "ovpm.CreateUserPerm is required for this operation")
	}

	var ut []*pb.UserResponse_User
	user, err := ovpm.CreateNewUser(req.Username, req.Password, req.NoGw, req.HostId, req.IsAdmin, req.Description)
	if err != nil {
		return nil, err
	}

	pbUser := pb.UserResponse_User{
		Username:           user.GetUsername(),
		ServerSerialNumber: user.GetServerSerialNumber(),
		NoGw:               user.IsNoGW(),
		HostId:             user.GetHostID(),
		IsAdmin:            user.IsAdmin(),
		Description:        user.GetDescription(),
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

	var admin bool

	switch req.AdminPref {
	case pb.UserUpdateRequest_ADMIN:
		admin = true
	case pb.UserUpdateRequest_NOADMIN:
		admin = false
	case pb.UserUpdateRequest_NOPREFADMIN:
		admin = user.IsAdmin()
	}

	perms, err := permset.FromContext(ctx)
	if err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "permset not found within the context")
	}

	username, err := GetUsernameFromContext(ctx)
	if err != nil {
		logrus.Debugln(err)
		return nil, grpc.Errorf(codes.Unauthenticated, "username not found with the provided credentials")
	}

	// User has admin perms?
	if perms.Contains(ovpm.UpdateAnyUserPerm) {
		err = user.Update(req.Password, noGW, req.HostId, admin, req.Description)
		if err != nil {
			return nil, err
		}
		ut = append(ut, &pb.UserResponse_User{
			Username:           user.GetUsername(),
			ServerSerialNumber: user.GetServerSerialNumber(),
			NoGw:               user.IsNoGW(),
			HostId:             user.GetHostID(),
			IsAdmin:            user.IsAdmin(),
			Description:        user.GetDescription(),
		})
		return &pb.UserResponse{Users: ut}, nil
	}

	// User has self update perms?
	if perms.Contains(ovpm.UpdateSelfPerm) {
		if user.GetUsername() != username {
			return nil, grpc.Errorf(codes.PermissionDenied, "Caller can only update their user with ovpm.UpdateSelfPerm")
		}

		err = user.Update(req.Password, noGW, req.HostId, admin, req.Description)
		if err != nil {
			return nil, err
		}
		ut = append(ut, &pb.UserResponse_User{
			Username:           user.GetUsername(),
			ServerSerialNumber: user.GetServerSerialNumber(),
			NoGw:               user.IsNoGW(),
			HostId:             user.GetHostID(),
			IsAdmin:            user.IsAdmin(),
			Description:        user.GetDescription(),
		})
		return &pb.UserResponse{Users: ut}, nil
	}
	return nil, grpc.Errorf(codes.PermissionDenied, "Permissions are required for this operation.")
}

func (s *UserService) Delete(ctx context.Context, req *pb.UserDeleteRequest) (*pb.UserResponse, error) {
	logrus.Debugf("rpc call: user delete: %s", req.Username)
	var ut []*pb.UserResponse_User
	user, err := ovpm.GetUser(req.Username)
	if err != nil {
		return nil, err
	}

	perms, err := permset.FromContext(ctx)
	if err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "Can't get permset from context")
	}

	if !perms.Contains(ovpm.DeleteAnyUserPerm) {
		return nil, grpc.Errorf(codes.PermissionDenied, "ovpm.DeleteAnyUserPerm is required for this operation.")
	}

	pbUser := pb.UserResponse_User{
		Username:           user.GetUsername(),
		ServerSerialNumber: user.GetServerSerialNumber(),
		HostId:             user.GetHostID(),
		IsAdmin:            user.IsAdmin(),
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
		HostId:             user.GetHostID(),
		IsAdmin:            user.IsAdmin(),
	}
	ut = append(ut, &pbUser)

	perms, err := permset.FromContext(ctx)
	if err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "Can't get permset from context")
	}

	if !perms.Contains(ovpm.RenewAnyUserPerm) {
		return nil, grpc.Errorf(codes.PermissionDenied, "ovpm.RenewAnyUserPerm is required for this operation.")
	}

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
	username, err := GetUsernameFromContext(ctx)
	if err != nil {
		logrus.Debugln(err)
		return nil, grpc.Errorf(codes.Unauthenticated, "username not found with the provided credentials")
	}

	perms, err := permset.FromContext(ctx)
	if err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "Can't get permset from context")
	}

	if perms.Contains(ovpm.GenConfigAnyUserPerm) {
		configBlob, err := ovpm.TheServer().DumpsClientConfig(user.GetUsername())
		if err != nil {
			return nil, err
		}
		return &pb.UserGenConfigResponse{ClientConfig: configBlob}, nil
	}

	if perms.Contains(ovpm.GenConfigSelfPerm) {
		if user.GetUsername() != username {
			return nil, grpc.Errorf(codes.PermissionDenied, "Caller can only genconfig for their user.")
		}
		configBlob, err := ovpm.TheServer().DumpsClientConfig(user.GetUsername())
		if err != nil {
			return nil, err
		}
		return &pb.UserGenConfigResponse{ClientConfig: configBlob}, nil
	}

	return nil, grpc.Errorf(codes.PermissionDenied, "Permissions are required for this operation.")
}

type VPNService struct{}

func (s *VPNService) Status(ctx context.Context, req *pb.VPNStatusRequest) (*pb.VPNStatusResponse, error) {
	logrus.Debugf("rpc call: vpn status")
	server := ovpm.TheServer()

	perms, err := permset.FromContext(ctx)
	if err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "Can't get permset from context")
	}

	if !perms.Contains(ovpm.GetVPNStatusPerm) {
		return nil, grpc.Errorf(codes.PermissionDenied, "ovpm.GetVPNStatusPerm is required for this operation.")
	}

	response := pb.VPNStatusResponse{
		Name:         server.GetServerName(),
		SerialNumber: server.GetSerialNumber(),
		Hostname:     server.GetHostname(),
		Port:         server.GetPort(),
		Proto:        server.GetProto(),
		Cert:         server.Cert,
		CaCert:       server.GetCACert(),
		Net:          server.GetNet(),
		Mask:         server.GetMask(),
		CreatedAt:    server.GetCreatedAt(),
		Dns:          server.GetDNS(),
		ExpiresAt:    server.ExpiresAt().UTC().Format(time.RFC3339),
		CaExpiresAt:  server.CAExpiresAt().UTC().Format(time.RFC3339),
		UseLzo:       server.IsUseLZO(),
	}
	return &response, nil
}

func (s *VPNService) Init(ctx context.Context, req *pb.VPNInitRequest) (*pb.VPNInitResponse, error) {
	logrus.Debugf("rpc call: vpn init")
	var proto string
	switch req.ProtoPref {
	case pb.VPNProto_TCP:
		proto = ovpm.TCPProto
	case pb.VPNProto_UDP:
		proto = ovpm.UDPProto
	case pb.VPNProto_NOPREF:
		proto = ovpm.UDPProto
	}

	perms, err := permset.FromContext(ctx)
	if err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "Can't get permset from context")
	}

	if !perms.Contains(ovpm.InitVPNPerm) {
		return nil, grpc.Errorf(codes.PermissionDenied, "ovpm.InitVPNPerm is required for this operation.")
	}

	if err := ovpm.TheServer().Init(req.Hostname, req.Port, proto, req.IpBlock, req.Dns, req.KeepalivePeriod, req.KeepaliveTimeout, req.UseLzo); err != nil {
		logrus.Errorf("server can not be created: %v", err)
	}
	return &pb.VPNInitResponse{}, nil
}

func (s *VPNService) Update(ctx context.Context, req *pb.VPNUpdateRequest) (*pb.VPNUpdateResponse, error) {
	logrus.Debugf("rpc call: vpn update")
	perms, err := permset.FromContext(ctx)
	if err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "Can't get permset from context")
	}

	if !perms.Contains(ovpm.UpdateVPNPerm) {
		return nil, grpc.Errorf(codes.PermissionDenied, "ovpm.UpdateVPNPerm is required for this operation.")
	}

	var useLzo *bool
	switch req.LzoPref {
	case pb.VPNLZOPref_USE_LZO_ENABLE:
		useLzo = ptr.Bool(true)
	case pb.VPNLZOPref_USE_LZO_DISABLE:
		useLzo = ptr.Bool(false)
	}
	if err := ovpm.TheServer().Update(req.IpBlock, req.Dns, useLzo); err != nil {
		logrus.Errorf("server can not be updated: %v", err)
	}
	return &pb.VPNUpdateResponse{}, nil
}

func (s *VPNService) Restart(ctx context.Context, req *pb.VPNRestartRequest) (*pb.VPNRestartResponse, error) {
	logrus.Debugf("rpc call: vpn restart")
	perms, err := permset.FromContext(ctx)
	if err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "Can't get permset from context")
	}

	if !perms.Contains(ovpm.RestartVPNPerm) {
		return nil, grpc.Errorf(codes.PermissionDenied, "ovpm.UpdateVPNPerm is required for this operation.")
	}

	ovpm.TheServer().RestartVPNProc()
	return &pb.VPNRestartResponse{}, nil
}

type NetworkService struct{}

func (s *NetworkService) List(ctx context.Context, req *pb.NetworkListRequest) (*pb.NetworkListResponse, error) {
	logrus.Debug("rpc call: network list")
	var nt []*pb.Network

	perms, err := permset.FromContext(ctx)
	if err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "Can't get permset from context")
	}

	if !perms.Contains(ovpm.ListNetworksPerm) {
		return nil, grpc.Errorf(codes.PermissionDenied, "ovpm.ListNetworksPerm is required for this operation.")
	}

	networks := ovpm.GetAllNetworks()
	for _, network := range networks {
		nt = append(nt, &pb.Network{
			Name:                network.GetName(),
			Cidr:                network.GetCIDR(),
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
	perms, err := permset.FromContext(ctx)
	if err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "Can't get permset from context")
	}

	if !perms.Contains(ovpm.CreateNetworkPerm) {
		return nil, grpc.Errorf(codes.PermissionDenied, "ovpm.CreateNetworkPerm is required for this operation.")
	}

	network, err := ovpm.CreateNewNetwork(req.Name, req.Cidr, ovpm.NetworkTypeFromString(req.Type), req.Via)
	if err != nil {
		return nil, err
	}

	n := pb.Network{
		Name:                network.GetName(),
		Cidr:                network.GetCIDR(),
		Type:                network.GetType().String(),
		CreatedAt:           network.GetCreatedAt(),
		AssociatedUsernames: network.GetAssociatedUsernames(),
		Via:                 network.GetVia(),
	}

	return &pb.NetworkCreateResponse{Network: &n}, nil
}

func (s *NetworkService) Delete(ctx context.Context, req *pb.NetworkDeleteRequest) (*pb.NetworkDeleteResponse, error) {
	logrus.Debugf("rpc call: network delete: %s", req.Name)
	perms, err := permset.FromContext(ctx)
	if err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "Can't get permset from context")
	}

	if !perms.Contains(ovpm.DeleteNetworkPerm) {
		return nil, grpc.Errorf(codes.PermissionDenied, "ovpm.DeleteNetworkPerm is required for this operation.")
	}

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
		Cidr:                network.GetCIDR(),
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

	perms, err := permset.FromContext(ctx)
	if err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "Can't get permset from context")
	}

	if !perms.Contains(ovpm.GetNetworkTypesPerm) {
		return nil, grpc.Errorf(codes.PermissionDenied, "ovpm.GetNetworkTypesPerm is required for this operation.")
	}

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
	perms, err := permset.FromContext(ctx)
	if err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "Can't get permset from context")
	}

	if !perms.Contains(ovpm.GetNetworkAssociatedUsersPerm) {
		return nil, grpc.Errorf(codes.PermissionDenied, "ovpm.GetNetworkAssociatedUsersPerm is required for this operation.")
	}

	network, err := ovpm.GetNetwork(req.Name)
	if err != nil {
		return nil, err
	}
	usernames := network.GetAssociatedUsernames()
	return &pb.NetworkGetAssociatedUsersResponse{Usernames: usernames}, nil
}

func (s *NetworkService) Associate(ctx context.Context, req *pb.NetworkAssociateRequest) (*pb.NetworkAssociateResponse, error) {
	logrus.Debugf("rpc call: network associate")
	perms, err := permset.FromContext(ctx)
	if err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "Can't get permset from context")
	}

	if !perms.Contains(ovpm.AssociateNetworkUserPerm) {
		return nil, grpc.Errorf(codes.PermissionDenied, "ovpm.AssociateNetworkUserPerm is required for this operation.")
	}

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
	perms, err := permset.FromContext(ctx)
	if err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "Can't get permset from context")
	}

	if !perms.Contains(ovpm.DissociateNetworkUserPerm) {
		return nil, grpc.Errorf(codes.PermissionDenied, "ovpm.DissociateNetworkUserPerm is required for this operation.")
	}

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
	var opts []grpc.ServerOption
	opts = append(opts, grpc.UnaryInterceptor(AuthUnaryInterceptor))
	s := grpc.NewServer(opts...)
	//s := grpc.NewServer()
	pb.RegisterUserServiceServer(s, &UserService{})
	pb.RegisterVPNServiceServer(s, &VPNService{})
	pb.RegisterNetworkServiceServer(s, &NetworkService{})
	pb.RegisterAuthServiceServer(s, &AuthService{})
	return s
}
