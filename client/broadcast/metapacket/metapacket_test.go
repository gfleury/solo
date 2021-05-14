package metapacket

import (
	"encoding/json"
	"testing"

	"github.com/gfleury/solo/client/broadcast/prp"
)

func TestMetapacket(t *testing.T) {
	var expected MetaPacket
	actual := NewFromPayload(prp.NewPRPRequestPacket("10.2.3.1"))

	a, err := json.Marshal(actual)
	if err != nil {
		t.Errorf(err.Error())
	}

	err = json.Unmarshal(a, &expected)
	if err != nil {
		t.Errorf(err.Error())
	}

	ap := actual.GetPayload().(*prp.PRPacket)
	ep := expected.GetPayload().(*prp.PRPacket)

	if ap.IP != ep.IP ||
		ap.Machine.IP != ep.Machine.IP ||
		ep.Machine.PeerID != ap.Machine.PeerID {
		t.Errorf("ap != ep ( %v != %v )", ap, ep)
	}
}
