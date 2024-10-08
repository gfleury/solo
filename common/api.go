package common

import "github.com/gfleury/solo/common/models"

type RegistrationResponse struct {
	Code string
}

type ConnectionConfigurationChallengeRequest struct {
	PeerID string
}

type ConnectionConfigurationChallengeResponse struct {
	Challenge string
}

type ConnectionConfigurationRequest struct {
	PeerID          string
	SignedChallenge string
}

type ConnectionConfigurationResponse struct {
	ConnectionConfigToken string
	InterfaceAddress      string
}

type NextIP struct {
	NextIP  string
	Network string
}

type NodeUpdateRequest struct {
	Node           models.NetworkNode
	SignedHostname string
}
