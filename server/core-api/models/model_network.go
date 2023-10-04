/*
 * Swagger solo - OpenAPI 3.0
 *
 * solo API
 *
 * API version: 1.0.0
 * Contact: apiteam@solo.io
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */
package models

import (
	"encoding/json"
)

type Network struct {
	Model

	Name string `json:"name"`

	CIDR string `json:"cidr,omitempty"`

	Nodes []Node `json:"nodes,omitempty"`

	// Owner User ID
	UserID uint
	User   *User `json:"user,omitempty"`

	// Users that are linked to this network
	LinkedUsers []LinkedUser `json:"linkedUsers,omitempty"`
}

type Node struct {
	Model

	NetworkID uint

	PeerID   string
	Hostname string
	OS       string
	Arch     string
	IP       string
	Version  string
}

func NewNetwork(name string) *Network {
	return &Network{
		Name: name,
	}
}

func (a *Network) Json() ([]byte, error) {
	return json.Marshal(a)
}

func (a *Network) Valid() bool {
	return true
}