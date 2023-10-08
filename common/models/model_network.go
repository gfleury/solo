/*
 *
 * solo Server API
 *
 */
package models

import (
	"crypto/ed25519"
	"encoding/json"
	"fmt"
	"net"
	"net/netip"
	"os"
	"runtime"
	"sort"
	"strings"

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
	Model
	NetworkID *uint    `json:"-"`
	Network   *Network `json:"-"`
	Actived   bool     `json:"-"`

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

func (a *Network) NextFreeIP() string {
	_, cidr, err := net.ParseCIDR(a.CIDR)
	if err != nil {
		return ""
	}
	if len(a.Nodes) == 0 {

		ip := cidr.IP
		ip[3] = uint8(1)

		nextIpWithMask := net.IPNet{IP: ip, Mask: cidr.Mask}
		return nextIpWithMask.String()
	}

	sort.Slice(a.Nodes, func(i, j int) bool {
		iIP, _, err := net.ParseCIDR(a.Nodes[i].IP)
		if err != nil {
			return false
		}
		jIP, _, err := net.ParseCIDR(a.Nodes[j].IP)
		if err != nil {
			return false
		}
		return uint8(iIP[3]) > uint8(jIP[3])
	})

	lastNode := a.Nodes[len(a.Nodes)-1]

	addr, err := netip.ParseAddr(strings.Split(lastNode.IP, "/")[0])
	if err != nil {
		return ""
	}

	ipAddr := addr.As4()
	ip := net.IP(ipAddr[:])
	ip[3] = uint8(ip[3]) + uint8(1)

	nextIpWithMask := net.IPNet{IP: ip, Mask: cidr.Mask}

	return nextIpWithMask.String()
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
	if a.Arch == "" || a.IP != "" || a.OS == "" ||
		a.PeerID == "" || a.Hostname == "" || a.Version == "" {
		return fmt.Errorf("node is invalid")
	}
	return nil
}

func NewLocalNode(host host.Host, IP string) NetworkNode {
	hostname, _ := os.Hostname()

	// Extract PubKey from private Key
	rawPrivKey, _ := host.Peerstore().PrivKey(host.ID()).Raw()
	privKey := ed25519.PrivateKey(rawPrivKey)
	pubKey := privKey.Public().(ed25519.PublicKey)

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
