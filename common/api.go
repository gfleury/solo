package common

type RegistrationResponse struct {
	Code string
}

type ConnectionConfigurationRequest struct {
	PeerID                  string
	NodeAuthenticationToken string
}

type ConnectionConfigurationResponse struct {
	ConnectionConfigToken string
	InterfaceAddress      string
}

type NextIP struct {
	NextIP  string
	Network string
}
