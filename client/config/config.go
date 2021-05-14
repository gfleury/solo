/*
Copyright © 2021-2022 Ettore Di Giacinto <mudler@mocaccino.org>
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package config

import (
	"github.com/gfleury/solo/client/discovery"
)

// Config is the config struct for the node and the default EdgeVPN services
// It is used to generate opts for the node and the services before start.
type Config struct {
	Token             string
	InterfaceAddress  string
	InterfaceName     string
	CreateInterface   bool
	Libp2pLogLevel    string
	LogLevel          string
	DiscoveryPeers    []string
	DiscoveryInterval int
	InterfaceMTU      int
	MaxConnections    int
	HolePunch         bool
	NatMap            bool
	NatService        bool
	RandomIdentity    bool
}

func Peers2List(peers []string) discovery.AddrList {
	addrsList := discovery.AddrList{}
	for _, p := range peers {
		addrsList.Set(p)
	}
	return addrsList
}
