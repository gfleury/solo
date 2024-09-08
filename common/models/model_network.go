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

	"github.com/gfleury/solo/client/utils"

	"github.com/lib/pq"
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

	PeerID      string `gorm:"embedded,uniqueIndex"`
	Hostname    string
	OS          string
	Arch        string
	IP          string
	Version     string
	PublicKey   []byte
	LocalRoutes pq.StringArray `gorm:"type:text[]"`
}

func NewNetwork(name, CIDR string) *Network {
	return &Network{
		Name: name,
		CIDR: CIDR,
	}
}

func (n *Network) NextFreeIP() string {
	_, cidr, err := net.ParseCIDR(n.CIDR)
	if err != nil {
		return ""
	}
	if len(n.Nodes) == 0 {
		ip := cidr.IP
		ip[3] = uint8(1)

		nextIpWithMask := net.IPNet{IP: ip, Mask: cidr.Mask}
		return nextIpWithMask.String()
	}

	sort.Slice(n.Nodes, func(i, j int) bool {
		iIP, _, err := net.ParseCIDR(n.Nodes[i].IP)
		if err != nil {
			return false
		}
		jIP, _, err := net.ParseCIDR(n.Nodes[j].IP)
		if err != nil {
			return false
		}
		return uint8(iIP[3]) > uint8(jIP[3])
	})

	lastNode := n.Nodes[len(n.Nodes)-1]

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

func (n *Network) Json() ([]byte, error) {
	return json.Marshal(n)
}

func (n *Network) Valid() error {
	if n.Name == "" {
		return fmt.Errorf("name can't be empty")
	}
	_, _, err := net.ParseCIDR(n.CIDR)
	return err
}

func (n *NetworkNode) Json() ([]byte, error) {
	return json.Marshal(n)
}

func (n *NetworkNode) Valid() error {
	if n.Arch == "" || n.OS == "" ||
		n.PeerID == "" || n.Hostname == "" || n.Version == "" {
		return fmt.Errorf("node is invalid")
	}
	return nil
}

func NewLocalNode(host host.Host, IP string) NetworkNode {
	return NewLocalNodeWithRoutes(host, IP, false)
}

func NewLocalNodeWithRoutes(host host.Host, IP string, fetchLocalRoutes bool) NetworkNode {
	hostname, _ := os.Hostname()

	// Extract PubKey from private Key
	rawPrivKey, _ := host.Peerstore().PrivKey(host.ID()).Raw()
	privKey := ed25519.PrivateKey(rawPrivKey)
	pubKey := privKey.Public().(ed25519.PublicKey)

	var networks []string
	var err error
	if fetchLocalRoutes {
		networks, err = utils.FetchLocalRoutes(IP)
		if err != nil {
			fmt.Println(err)
		}
	}

	return NetworkNode{
		PeerID:      host.ID().String(),
		Hostname:    hostname,
		OS:          runtime.GOOS,
		Arch:        runtime.GOARCH,
		Version:     "0.0.1",
		IP:          IP,
		PublicKey:   pubKey,
		LocalRoutes: networks,
	}
}
