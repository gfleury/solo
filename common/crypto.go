package common

import (
	"crypto/ed25519"
)

var NodeAuthenticationTokenOptions = &ed25519.Options{
	Context: "Solo_Node_Authentication",
}
