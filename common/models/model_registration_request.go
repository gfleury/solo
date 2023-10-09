package models

import (
	"crypto/rand"
	"encoding/base32"
	"encoding/json"
	"fmt"
	"strings"
)

type RegistrationRequest struct {
	Code   string `gorm:"primarykey"`
	NodeID uint
	Node   NetworkNode
}

func (a *RegistrationRequest) Json() ([]byte, error) {
	return json.Marshal(a)
}

func GenerateNewRandomCode(size int) (string, error) {
	b := make([]byte, size)
	n, err := rand.Read(b)
	if n != size {
		return "", fmt.Errorf("unable to read %d bytes from random", size)
	}
	if err != nil {
		return "", err
	}
	code := strings.ToUpper(base32.HexEncoding.EncodeToString(b))[:size]
	return code, nil
}
