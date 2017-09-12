package api

import (
	"context"
	"fmt"

	gcontext "golang.org/x/net/context"
)

type apiKey int

const (
	originTypeKey apiKey = iota
	userKey
)

// OriginType indicates where the gRPC request actually came from.
//
// e.g. Is it from REST gateway? Or somewhere else?
type OriginType int

// Known origins
const (
	OriginTypeUnknown OriginType = iota
	OriginTypeREST
)

// NewOriginTypeContext creates a new ctx from the OriginType.
func NewOriginTypeContext(ctx gcontext.Context, originType OriginType) context.Context {
	return context.WithValue(ctx, originTypeKey, originType)
}

// GetOriginTypeFromContext returns the OriginType from context.
func GetOriginTypeFromContext(ctx gcontext.Context) OriginType {
	originType, ok := ctx.Value(originTypeKey).(OriginType)
	if !ok {
		return OriginTypeUnknown
	}
	return originType
}

// NewUsernameContext creates a new ctx from username and returns it.
func NewUsernameContext(ctx gcontext.Context, username string) context.Context {
	return context.WithValue(ctx, userKey, username)
}

// GetUsernameFromContext returns the Username from context.
func GetUsernameFromContext(ctx gcontext.Context) (string, error) {
	username, ok := ctx.Value(userKey).(string)
	if !ok {
		return "", fmt.Errorf("cannot get context value")
	}
	return username, nil
}
