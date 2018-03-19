package main

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/cad/ovpm"
	"github.com/cad/ovpm/api/pb"
	"github.com/cad/ovpm/errors"
	"github.com/olekukonko/tablewriter"
)

// userListAction lists existing VPN users on the terminal.
//
// List includes additional information about users in addition to usernames
// such as; their IP addresses, the time the user is created at etc...
func userListAction(rpcServURLStr string) error {
	// Parse RPC Server's URL.
	rpcSrvURL, err := url.Parse(rpcServURLStr)
	if err != nil {
		return errors.BadURL(rpcServURLStr, err)
	}

	// Create a gRPC connection to the server.
	rpcConn, err := grpcConnect(rpcSrvURL)
	if err != nil {
		exit(1)
		return err
	}
	defer rpcConn.Close()

	// Get services.
	var userSvc = pb.NewUserServiceClient(rpcConn)
	var vpnSvc = pb.NewVPNServiceClient(rpcConn)

	// Request vpn status and user list from the services.
	vpnStatusResp, err := vpnSvc.Status(context.Background(), &pb.VPNStatusRequest{})
	if err != nil {
		err := errors.UnknownGRPCError(err)
		exit(1)
		return err
	}
	userListResp, err := userSvc.List(context.Background(), &pb.UserListRequest{})
	if err != nil {
		err := errors.UnknownGRPCError(err)
		exit(1)
		return err
	}

	// Prepare table data.
	header := []string{"#", "username", "ip", "created at", "valid crt", "push gw"}
	rows := [][]string{}
	for i, user := range userListResp.Users {
		static := ""
		if user.HostId != 0 {
			static = "s"
		}
		username := user.Username
		if user.IsAdmin {
			username = fmt.Sprintf("%s *", username)
		}
		row := []string{
			fmt.Sprintf("%v", i+1),
			username,
			fmt.Sprintf("%s %s", user.IpNet, static),
			user.CreatedAt,
			fmt.Sprintf("%t", user.ServerSerialNumber == vpnStatusResp.SerialNumber),
			fmt.Sprintf("%t", !user.NoGw),
		}
		rows = append(rows, row)
	}

	// Draw the table on the terminal.
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(header)
	table.AppendBulk(rows)
	table.Render()

	return nil
}

// userCreateAction creates a new VPN user from the terminal.
func userCreateAction(rpcSrvURLStr string, username string, password string, ipAddr *net.IP, noGW bool, isAdmin bool) error {
	// Parse RPC Server's URL.
	rpcSrvURL, err := url.Parse(rpcSrvURLStr)
	if err != nil {
		return errors.BadURL(rpcSrvURLStr, err)
	}

	// Create a gRPC connection to the server.
	rpcConn, err := grpcConnect(rpcSrvURL)
	if err != nil {
		exit(1)
		return err
	}
	defer rpcConn.Close()

	// hostid is the integer representation of an IPv4 address.
	// OVPM uses it to send, receive and persist IP addresses instead of
	// sending dotted string representation.
	//
	// If it is 0, that means no IP address is set.
	var hostid uint32

	// Determine the hostid according to the provided ipAddr.
	if ipAddr != nil {
		hostid = ovpm.IP2HostID(ipAddr.To4())
		if hostid == 0 {
			// hostid being 0 means dynamic ip addr(no static ip address provided),
			// hence ambiguous meaning here.
			// This is perceived as an error.
			return errors.ConflictingDemands("hostid is 0, but user is trying to allocate a static ip addr")
		}
	}

	// Prepare a service caller.
	var userSvc = pb.NewUserServiceClient(rpcConn)

	// Send a user creation request to the server.
	userCreateResp, err := userSvc.Create(context.Background(), &pb.UserCreateRequest{
		Username: username,
		Password: password,
		NoGw:     noGW,
		HostId:   hostid,
		IsAdmin:  isAdmin,
	})
	if err != nil {
		err := errors.UnknownGRPCError(err)
		exit(1)
		return err
	}

	logrus.Infof("user created: %s", userCreateResp.Users[0].Username)
	return nil
}

// userUpdateAction creates a new VPN user from the terminal.
func userUpdateAction(rpcSrvURLStr string, username string, password *string, ipAddr *net.IP, isStatic *bool, noGW *bool, isAdmin *bool) error {
	// Parse RPC Server's URL.
	rpcSrvURL, err := url.Parse(rpcSrvURLStr)
	if err != nil {
		return errors.BadURL(rpcSrvURLStr, err)
	}

	// Create a gRPC connection to the server.
	rpcConn, err := grpcConnect(rpcSrvURL)
	if err != nil {
		exit(1)
		return err
	}
	defer rpcConn.Close()

	// Set target password.
	targetPassword := ""
	if password != nil {
		targetPassword = *password
	}

	// Set targeted static IP addr.
	targetHostid := uint32(0)
	targetStaticPref := pb.UserUpdateRequest_NOPREFSTATIC
	if isStatic != nil {
		if *isStatic {
			targetHostid = ovpm.IP2HostID(ipAddr.To4())
			if targetHostid == 0 {
				// hostid being 0 means dynamic ip addr(no static ip address provided),
				// hence ambiguous meaning here.
				// This is perceived as an error.
				return errors.ConflictingDemands("hostid is 0, but user is trying to allocate a static ip addr")
			}
			targetStaticPref = pb.UserUpdateRequest_STATIC
		} else {
			targetStaticPref = pb.UserUpdateRequest_NOSTATIC
		}
	}

	// Set targeted gwPref.
	targetGWPref := pb.UserUpdateRequest_NOPREF
	if noGW != nil {
		if *noGW {
			targetGWPref = pb.UserUpdateRequest_NOGW
		} else {
			targetGWPref = pb.UserUpdateRequest_GW
		}
	}

	// Set targeted adminPref.
	targetAdminPref := pb.UserUpdateRequest_NOPREFADMIN
	if isAdmin != nil {
		if *isAdmin {
			targetAdminPref = pb.UserUpdateRequest_ADMIN
		} else {
			targetAdminPref = pb.UserUpdateRequest_NOADMIN
		}
	}

	// Prepare a service caller.
	var userSvc = pb.NewUserServiceClient(rpcConn)

	// Send a user update request to the server.
	userUpdateResp, err := userSvc.Update(context.Background(), &pb.UserUpdateRequest{
		Username:   username,
		Password:   targetPassword,
		Gwpref:     targetGWPref,
		StaticPref: targetStaticPref,
		HostId:     targetHostid,
		AdminPref:  targetAdminPref,
	})
	if err != nil {
		err := errors.UnknownGRPCError(err)
		exit(1)
		return err
	}

	logrus.Infof("user updated: %s", userUpdateResp.Users[0].Username)
	return nil
}

// userDeleteAction deletes a VPN user.
func userDeleteAction(rpcSrvURLStr string, username string) error {
	// Parse RPC Server's URL.
	rpcSrvURL, err := url.Parse(rpcSrvURLStr)
	if err != nil {
		return errors.BadURL(rpcSrvURLStr, err)
	}

	// Create a gRPC connection to the server.
	rpcConn, err := grpcConnect(rpcSrvURL)
	if err != nil {
		exit(1)
		return err
	}
	defer rpcConn.Close()

	// Prepare a service caller.
	var userSvc = pb.NewUserServiceClient(rpcConn)

	// Send a user delete request to the server.
	userDeleteRequest, err := userSvc.Delete(context.Background(), &pb.UserDeleteRequest{Username: username})
	if err != nil {
		err := errors.UnknownGRPCError(err)
		exit(1)
		return err
	}

	logrus.Infof("user deleted: %s", userDeleteRequest.Users[0].Username)
	return nil
}

// userRenewAction renews a VPN user.
func userRenewAction(rpcSrvURLStr string, username string) error {
	// Parse RPC Server's URL.
	rpcSrvURL, err := url.Parse(rpcSrvURLStr)
	if err != nil {
		return errors.BadURL(rpcSrvURLStr, err)
	}

	// Create a gRPC connection to the server.
	rpcConn, err := grpcConnect(rpcSrvURL)
	if err != nil {
		exit(1)
		return err
	}
	defer rpcConn.Close()

	// Prepare a service caller.
	var userSvc = pb.NewUserServiceClient(rpcConn)

	// Send a user renew request to the server.
	userRenewResp, err := userSvc.Renew(context.Background(), &pb.UserRenewRequest{Username: username})
	if err != nil {
		err := errors.UnknownGRPCError(err)
		exit(1)
		return err
	}

	logrus.Infof("user renewed: %s", userRenewResp.Users[0].Username)
	return nil
}

// userGenconfigAction generates ovpn configs for a VPN user.
func userGenconfigAction(rpcSrvURLStr string, username string, outPath *string) error {
	// Parse RPC Server's URL.
	rpcSrvURL, err := url.Parse(rpcSrvURLStr)
	if err != nil {
		return errors.BadURL(rpcSrvURLStr, err)
	}

	// Create a gRPC connection to the server.
	rpcConn, err := grpcConnect(rpcSrvURL)
	if err != nil {
		exit(1)
		return err
	}
	defer rpcConn.Close()

	// If no outPath is provided, then use the default one with
	// the username.
	if outPath == nil {
		tmp := fmt.Sprintf("%s.ovpn", username)
		outPath = &tmp
	}

	// Prepare a service caller.
	var userSvc = pb.NewUserServiceClient(rpcConn)

	// Send a user genconfig request to the server.
	userGenconfigResp, err := userSvc.GenConfig(context.Background(), &pb.UserGenConfigRequest{Username: username})
	if err != nil {
		err := errors.UnknownGRPCError(err)
		exit(1)
		return err
	}

	// Write out the contents of the vpn profile
	// to the filesystem.
	if err := emitToFile(*outPath, userGenconfigResp.ClientConfig, 0); err != nil {
		err := errors.UnknownFileIOError(err)
		exit(1)
		return err
	}

	logrus.Infof("exported to %s", *outPath)
	return nil
}
