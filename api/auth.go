package api

import (
	"fmt"
	"strings"

	"github.com/cad/ovpm"
	"github.com/cad/ovpm/permset"
	"github.com/sirupsen/logrus"
	gcontext "golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
)

func authRequired(ctx gcontext.Context, req interface{}, handler grpc.UnaryHandler) (resp interface{}, err error) {
	logrus.Debugln("rpc: auth applied")
	token, err := authzTokenFromContext(ctx)
	if err != nil {
		logrus.Debugln("rpc: auth denied because token can not be gathered from header contest")
		return nil, grpc.Errorf(codes.Unauthenticated, err.Error())
	}
	user, err := ovpm.GetUserByToken(token)
	if err != nil {
		logrus.Debugln("rpc: auth denied because user with this token can not be found")
		return nil, grpc.Errorf(codes.Unauthenticated, "access denied")
	}

	// Set user's permissions according to it's criteria.
	var permissions permset.Permset
	if user.IsAdmin() {
		permissions = permset.New(ovpm.AdminPerms()...)
	} else {
		permissions = permset.New(ovpm.UserPerms()...)
	}

	newCtx := NewUsernameContext(ctx, user.GetUsername())
	newCtx = permset.NewContext(newCtx, permissions)
	return handler(newCtx, req)
}

func authzTokenFromContext(ctx gcontext.Context) (string, error) {
	// retrieve metadata from context
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", fmt.Errorf("authentication required")
	}
	if len(md["authorization"]) != 1 {
		return "", fmt.Errorf("authentication required (length)")
	}

	authHeader := md["authorization"][0]

	// split authorization header into two
	splitToken := strings.Split(authHeader, "Bearer")
	if len(splitToken) != 2 {
		return "", fmt.Errorf("invalid Authorization header. it should be in the form of 'Bearer <token>': %s", authHeader)
	}
	// get token
	token := splitToken[1]
	token = strings.TrimSpace(token)
	return token, nil
}
