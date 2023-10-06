package models

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetNextFreeIP(t *testing.T) {
	network := Network{CIDR: "10.1.0.0/24"}

	require.Equal(t, "10.1.0.1/24", network.NextFreeIP())

	network.Nodes = append(network.Nodes, NetworkNode{IP: network.NextFreeIP()})

	require.Equal(t, "10.1.0.2/24", network.NextFreeIP())

	network.Nodes = append(network.Nodes, NetworkNode{IP: network.NextFreeIP()})
	network.Nodes = append(network.Nodes, NetworkNode{IP: network.NextFreeIP()})
	network.Nodes = append(network.Nodes, NetworkNode{IP: network.NextFreeIP()})
	network.Nodes = append(network.Nodes, NetworkNode{IP: network.NextFreeIP()})
	network.Nodes = append(network.Nodes, NetworkNode{IP: network.NextFreeIP()})
	network.Nodes = append(network.Nodes, NetworkNode{IP: network.NextFreeIP()})
	network.Nodes = append(network.Nodes, NetworkNode{IP: network.NextFreeIP()})
	network.Nodes = append(network.Nodes, NetworkNode{IP: network.NextFreeIP()})

	require.Equal(t, "10.1.0.10/24", network.NextFreeIP())
}
