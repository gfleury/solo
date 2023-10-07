package common

import "encoding/json"

type RegistrationResponse struct {
	Code string
}

func (a *RegistrationResponse) Json() ([]byte, error) {
	return json.Marshal(a)
}

type ConnectionConfigurationRequest struct {
	PeerID                  string
	NodeAuthenticationToken string
}

type ConnectionConfigurationResponse struct {
	ConnectionConfigToken string
	InterfaceAddress      string
}

func (c *ConnectionConfigurationResponse) Json() ([]byte, error) {
	return json.Marshal(c)
}
