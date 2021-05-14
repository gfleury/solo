package protocol

import "github.com/ipfs/go-log"

type Type int

const (
	Type_PRP = iota
)

type Payload interface {
	Type() Type
	Payload() string
	Process(logger log.StandardLogger, table interface{}) (Payload, error)
}
