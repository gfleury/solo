package prp

import (
	"encoding/json"
	"net"

	"github.com/gfleury/solo/client/broadcast/protocol"
	"github.com/gfleury/solo/common/models"
	"github.com/ipfs/go-log"
)

type PRPPacketType int

const (
	PRPRequest PRPPacketType = iota
	PRPReply
)

type PRPacket struct {
	PRPType PRPPacketType
	Machine models.NetworkNode
	IP      string
}

func (p *PRPacket) Type() protocol.Type {
	return protocol.Type_PRP
}

func (p *PRPacket) Payload() string {
	bytesPayload, _ := json.Marshal(p)
	return string(bytesPayload)
}

func (p *PRPacket) Process(logger log.StandardLogger, table interface{}) (protocol.Payload, error) {

	PRPTable := table.(*PRPTableType)

	switch p.PRPType {
	case PRPReply:
		logger.Infof("PRPReply IP: %s Machine: %v", p.IP, p.Machine)
		PRPTable.insertEntry(p.IP, &p.Machine)
	case PRPRequest:
		logger.Infof("PRPRequest IP: who's %s?  I'm %s", p.IP, PRPTable.localIP)
		if p.IP == PRPTable.localIP {
			return PRPTable.PRPReplyMyself(false), nil
		} else {
			_, mySelf := PRPTable.Myself()
			for _, route := range mySelf.LocalRoutes {
				_, ipnet, err := net.ParseCIDR(route)
				if err != nil {
					logger.Errorf("Failed to query route: %s with error %s", route, err)
					continue
				}
				if ipnet.Contains(net.IP(p.IP)) {
					logger.Infof("Ip %s belongs to network %s", p.IP, ipnet)
					return PRPTable.PRPReplyMyself(false), nil
				}
			}
		}
	}
	return nil, nil
}

func NewPRPRequestPacket(ip string) *PRPacket {
	return &PRPacket{
		PRPType: PRPRequest,
		IP:      ip,
	}
}
