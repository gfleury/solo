package common

import (
	"crypto/ed25519"
	"fmt"
	"os"
	"time"
)

var NodeAuthenticationTokenOptions = &ed25519.Options{
	Context: "Solo_Node_Authentication",
}

func NodeAuthenticationTokenMessage(localHostname ...string) []byte {
	hostname, _ := os.Hostname()
	if len(localHostname) > 0 {
		hostname = localHostname[0]
	}
	return []byte(fmt.Sprintf("%s@%s", hostname, time.Now().UTC().Format("2006-01-02")))
}
