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

	// Network permissions
	ListNetworksPerm
	CreateNetworkPerm
	DeleteNetworkPerm
	GetNetworkTypesPerm
	GetNetworkAssociatedUsersPerm
	AssociateNetworkUserPerm
	DissociateNetworkUserPerm
)

// AdminPerms is a collection of permissions for Admin.
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
		ListNetworksPerm,
		CreateNetworkPerm,
		DeleteNetworkPerm,
		GetNetworkTypesPerm,
		GetNetworkAssociatedUsersPerm,
		AssociateNetworkUserPerm,
		DissociateNetworkUserPerm,
	}
}

// UserPerms is a collection of permissions for User.
func UserPerms() []permset.Perm {
	return []permset.Perm{
		GetSelfPerm,
		UpdateSelfPerm,
		GenConfigSelfPerm,
	}
}
