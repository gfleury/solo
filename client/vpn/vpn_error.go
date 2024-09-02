package vpn

import (
	"fmt"
)

type ErrorType int

const (
	HostNotFound ErrorType = iota
)

type VpnError struct {
	Type    ErrorType
	Message string
}

func (e *VpnError) Error() string {
	return fmt.Sprintf("error %d: %s", e.Type, e.Message)
}

func NewNotFoundError(dst string) error {
	return &VpnError{Type: HostNotFound, Message: fmt.Sprintf("'%s' not found in the routing table", dst)}
}
