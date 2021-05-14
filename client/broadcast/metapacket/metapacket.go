package metapacket

import (
	"encoding/json"

	"github.com/gfleury/solo/client/broadcast/protocol"
	"github.com/gfleury/solo/client/broadcast/prp"
)

type MetaPacket struct {
	Type     protocol.Type
	SenderID string
	Payload  string
}

func NewMetaPacket(t protocol.Type, payload protocol.Payload) *MetaPacket {
	return &MetaPacket{t, "", payload.Payload()}
}

func NewFromPayload(payload protocol.Payload) *MetaPacket {
	return NewMetaPacket(payload.Type(), payload)
}

func (m *MetaPacket) Copy() *MetaPacket {
	copy := *m
	return &copy
}

func (m *MetaPacket) GetPayload() protocol.Payload {
	switch m.Type {
	case protocol.Type_PRP:
		p := &prp.PRPacket{}
		json.Unmarshal([]byte(m.Payload), p)
		return p
	}

	return nil
}
