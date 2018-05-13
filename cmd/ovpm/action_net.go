package main

import (
	"context"
	"fmt"
	"net/url"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/asaskevich/govalidator"
	"github.com/cad/ovpm"
	"github.com/cad/ovpm/api/pb"
	"github.com/cad/ovpm/errors"
	"github.com/olekukonko/tablewriter"
)

func netListAction(rpcServURLStr string) error {
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

	// Prepare service callable.
	var netSvc = pb.NewNetworkServiceClient(rpcConn)

	// Request vpn status and user list from the services.
	netListResp, err := netSvc.List(context.Background(), &pb.NetworkListRequest{})
	if err != nil {
		err := errors.UnknownGRPCError(err)
		exit(1)
		return err
	}

	// Render the network table.
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"#", "name", "cidr", "type", "assoc", "created at"})
	for i, network := range netListResp.Networks {
		// Create associated user list for this network.
		var usernameList string
		assocUsers, err := netSvc.GetAssociatedUsers(context.Background(), &pb.NetworkGetAssociatedUsersRequest{Name: network.Name})
		if err != nil {
			logrus.Errorf("assoc users can not be fetched: %v", err)
			exit(1)
			return err
		}

		usernames := assocUsers.Usernames
		count := len(usernames)
		for i, uname := range usernames {
			if i+1 == count {
				usernameList = usernameList + fmt.Sprintf("%s", uname)
			} else {
				usernameList = usernameList + fmt.Sprintf("%s, ", uname)
			}
		}
		var cidr = network.Cidr
		var via = network.Via
		if via == "" {
			via = "vpn-server"
		}
		if ovpm.NetworkTypeFromString(network.Type) == ovpm.ROUTE {
			cidr = fmt.Sprintf("%s via %s", network.Cidr, via)
		}
		data := []string{fmt.Sprintf("%v", i+1), network.Name, cidr, network.Type, usernameList, network.CreatedAt}
		table.Append(data)
	}
	table.Render()

	return nil
}

func netTypesAction(rpcServURLStr string) error {
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

	// Prepare service callable.
	var netSvc = pb.NewNetworkServiceClient(rpcConn)

	// Request network types from the service.
	netGetAllResp, err := netSvc.GetAllTypes(context.Background(), &pb.NetworkGetAllTypesRequest{})
	if err != nil {
		err := errors.UnknownGRPCError(err)
		exit(1)
		return err
	}

	// Render table.
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"#", "net type", "desc"})
	for i, ntype := range netGetAllResp.Types {
		data := []string{fmt.Sprintf("%v", i+1), ntype.Type, ntype.Description}
		table.Append(data)
	}
	table.Render()

	return nil
}

func netUndefAction(rpcServURLStr string, networkName string) error {
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

	// Prepare service callable.
	var netSvc = pb.NewNetworkServiceClient(rpcConn)

	// Call the service.
	netDeleteResp, err := netSvc.Delete(context.Background(), &pb.NetworkDeleteRequest{Name: networkName})
	if err != nil {
		logrus.Errorf("networks can not be deleted: %v", err)
		exit(1)
		return err
	}

	logrus.Infof("network deleted: %s (%s)", netDeleteResp.Network.Name, netDeleteResp.Network.Cidr)
	return nil
}

func netDefAction(rpcServURLStr string, netName string, netCIDR string, netType string, via *string) error {
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

	var targetVia string
	switch ovpm.NetworkTypeFromString(netType) {
	case ovpm.ROUTE:
		if via != nil {
			if !govalidator.IsIPv4(*via) {
				err := errors.NotIPv4(*via)
				exit(1)
				return err
			}
			targetVia = *via
		}
	case ovpm.SERVERNET:
		if via != nil && govalidator.IsNull(*via) {
			err := errors.ConflictingDemands("--via flag can only be used with --type ROUTE")
			exit(1)
			return err
		}
	default: // Means UNDEFINEDNET
		fmt.Printf("undefined network type %s", netType)
		fmt.Println()
		fmt.Println("Network Types:")
		fmt.Println("    ", ovpm.GetAllNetworkTypes())
		fmt.Println()
		exit(1)
		return fmt.Errorf("undefined network type")
	}

	// Prepare service callable.
	var netSvc = pb.NewNetworkServiceClient(rpcConn)

	// Call the service.
	netCreateResp, err := netSvc.Create(context.Background(), &pb.NetworkCreateRequest{Name: netName, Cidr: netCIDR, Type: netType, Via: targetVia})
	if err != nil {
		logrus.Errorf("network can not be created '%s': %v", netName, err)
		exit(1)
		return err
	}
	logrus.Infof("network created: %s (%s)", netCreateResp.Network.Name, netCreateResp.Network.Cidr)
	return nil
}

func netAssocAction(rpcServURLStr string, netName string, username string, inBulk bool) error {
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

	// Prepare service callable.
	var netSvc = pb.NewNetworkServiceClient(rpcConn)

	userNames := []string{username}
	if inBulk {
		var userSvc = pb.NewUserServiceClient(rpcConn)
		r, err := userSvc.List(context.Background(), &pb.UserListRequest{})
		if err != nil {
			errors.UnknownGRPCError(err)
			exit(1)
			return err
		}
		userNames = []string{}
		for _, u := range r.Users {
			userNames = append(userNames, u.Username)
		}
	}

	// Call the service.
	for _, userName := range userNames {
		_, err = netSvc.Associate(context.Background(), &pb.NetworkAssociateRequest{Name: netName, Username: userName})
		if err != nil {
			errors.UnknownGRPCError(err)
			//exit(1)
			//return err
		}
		logrus.Infof("network associated: user:%s <-> network:%s", userName, netName)
	}
	return nil
}

func netDissocAction(rpcServURLStr string, netName string, username string, inBulk bool) error {
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

	// Prepare service callable.
	var netSvc = pb.NewNetworkServiceClient(rpcConn)

	userNames := []string{username}
	if inBulk {
		var userSvc = pb.NewUserServiceClient(rpcConn)
		r, err := userSvc.List(context.Background(), &pb.UserListRequest{})
		if err != nil {
			errors.UnknownGRPCError(err)
			exit(1)
			return err
		}
		userNames = []string{}
		for _, u := range r.Users {
			userNames = append(userNames, u.Username)
		}
	}

	// Call the service.
	for _, userName := range userNames {
		_, err = netSvc.Dissociate(context.Background(), &pb.NetworkDissociateRequest{Name: netName, Username: userName})
		if err != nil {
			errors.UnknownGRPCError(err)
			//exit(1)
			//return err
		}

		logrus.Infof("network dissociated: user:%s <-> network:%s", userName, netName)
	}
	return nil
}
