package controller

import (
	"net"
	"net/url"
)

// Interface can be implemented by VPN servers that can be controlled via GridVPN.
type Interface interface {
	// Start MUST cause the underlying VPN server to start accepting connections
	// by using the config at hand.
	Start() error

	// Stop MUST cause the underlying VPN server to stop accepting connections and freeing
	// the resources it had allocated.
	Stop() error

	// Reload SHOULD cause the underlying VPN server to take in effect the changes made
	// on the config at hand.
	Reload() error

	// Status MUST return the appropriate StatusProvider according to the state that the server is in.
	Status() StatusProvider

	// Configure accepts a new config and SHOULD reflect the changes that are made to the config at hand,
	// but SHOULD NOT take those changes in effect.
	//
	// Invoking Configure(ConfigProvider) with a nil ConfigProvider implies a server config reset/init.
	Configure(ConfigProvider) error

	// Config MUST return the config at hand.
	Config() (ConfigProvider, error)
}

// ConfigProvider interface represents a VPN server configuration.
type ConfigProvider interface {
	Kind() string           // The kind of the server. OpenVPN, L2TP etc..
	ListenAddrs() []url.URL // Server addresses to listen at.
	NameServers() []net.IP  // DNS server addresses to push to the peers.
}

// StatusProvider interface represents the current status of the VPN server.
type StatusProvider interface {
	State() string
	// ConnectedPeers() []Peer // Currently connected peers on the VPN server.
}
