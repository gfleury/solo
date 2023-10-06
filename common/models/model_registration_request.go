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

func (a *RegistrationRequest) GenerateNewCode() (string, error) {
	b := make([]byte, 6)
	n, err := rand.Read(b)
	if n != 6 {
		return "", fmt.Errorf("unable to read 6 bytes from random")
	}
	if err != nil {
		return "", err
	}
	a.Code = strings.ToUpper(base32.HexEncoding.EncodeToString(b))[:6]
	return a.Code, nil
}
