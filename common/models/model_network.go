/*
 *
 * solo Server API
 *
 */
package models

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"runtime"

	"github.com/libp2p/go-libp2p/core/host"
)

type Network struct {
	Model

	Name string `json:"name"`

	CIDR string `json:"cidr,omitempty"`

	ConnectionConfigToken string `json:"connection_config,omitempty"`

	Nodes []NetworkNode `json:"nodes,omitempty"`

	// Owner User ID
	UserID uint
	User   *User `json:"user,omitempty"`

	// Users that are linked to this network
	LinkedUsers []LinkedUser `json:"linkedUsers,omitempty"`
}

type NetworkNode struct {
	Model `json:"-"`

	NetworkID uint `json:"-"`

	PeerID    string `gorm:"uniqueIndex"`
	Hostname  string
	OS        string
	Arch      string
	IP        string
	Version   string
	PublicKey []byte
}

func NewNetwork(name, CIDR string) *Network {
	return &Network{
		Name: name,
		CIDR: CIDR,
	}
}

func (a *Network) Json() ([]byte, error) {
	return json.Marshal(a)
}

func (a *Network) Valid() error {
	if a.Name == "" {
		return fmt.Errorf("name can't be empty")
	}
	_, _, err := net.ParseCIDR(a.CIDR)
	return err
}

func (a *NetworkNode) Json() ([]byte, error) {
	return json.Marshal(a)
}

func (a *NetworkNode) Valid() error {
	if a.Arch == "" || a.IP == "" || a.OS == "" ||
		a.PeerID == "" || a.Hostname == "" || a.Version == "" {
		return fmt.Errorf("node is invalid")
	}
	return nil
}

func NewLocalNode(host host.Host, IP string) NetworkNode {
	hostname, _ := os.Hostname()

	pubKey, _ := host.Peerstore().PubKey(host.ID()).Raw()
	return NetworkNode{
		PeerID:    host.ID().String(),
		Hostname:  hostname,
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
		Version:   "0.0.1",
		IP:        IP,
		PublicKey: pubKey,
	}
}
