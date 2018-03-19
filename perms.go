package ovpm

import "github.com/cad/ovpm/permset"

// OVPM available permissions.
const (
	// User permissions
	CreateUserPerm permset.Perm = iota
	GetAnyUserPerm
	GetSelfPerm
	UpdateAnyUserPerm
	UpdateSelfPerm
	DeleteAnyUserPerm
	RenewAnyUserPerm
	GenConfigAnyUserPerm
	GenConfigSelfPerm

	// VPN permissions
	GetVPNStatusPerm
	InitVPNPerm
	UpdateVPNPerm
	RestartVPNPerm

	// Network permissions
	ListNetworksPerm
	CreateNetworkPerm
	DeleteNetworkPerm
	GetNetworkTypesPerm
	GetNetworkAssociatedUsersPerm
	AssociateNetworkUserPerm
	DissociateNetworkUserPerm
)

// AdminPerms returns the list of permissions that admin type user has.
func AdminPerms() []permset.Perm {
	return []permset.Perm{
		CreateUserPerm,
		GetAnyUserPerm,
		GetSelfPerm,
		UpdateAnyUserPerm,
		UpdateSelfPerm,
		DeleteAnyUserPerm,
		RenewAnyUserPerm,
		GenConfigAnyUserPerm,
		GenConfigSelfPerm,
		GetVPNStatusPerm,
		InitVPNPerm,
		UpdateVPNPerm,
		RestartVPNPerm,
		ListNetworksPerm,
		CreateNetworkPerm,
		DeleteNetworkPerm,
		GetNetworkTypesPerm,
		GetNetworkAssociatedUsersPerm,
		AssociateNetworkUserPerm,
		DissociateNetworkUserPerm,
	}
}

// UserPerms returns the collection of permissions that the regular users have.
func UserPerms() []permset.Perm {
	return []permset.Perm{
		GetSelfPerm,
		UpdateSelfPerm,
		GenConfigSelfPerm,
	}
}
