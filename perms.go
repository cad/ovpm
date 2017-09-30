package ovpm

import "github.com/cad/ovpm/permset"

// OVPM defined permissions.
const (
	CreateUserPerm permset.Perm = iota
	GetAnyUserPerm
	GetSelfPerm
	UpdateAnyUserPerm
	UpdateSelfPerm
	DeleteAnyUserPerm
)
