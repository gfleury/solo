package noise

import (
	"fmt"
	"io"

	"github.com/flynn/noise"
)

type StreamState int

const (
	E StreamState = iota
	DHEE
	READY
)

type NoiseStream interface {
	Encrypt([]byte) ([]byte, error)
	Decrypt([]byte) ([]byte, error)
	DoHandshake([]byte) ([]byte, error)
	IsReady() bool
}

type NoiseStreamStandard struct {
	cipherSuite    noise.CipherSuite
	handshakeState *noise.HandshakeState
	state          StreamState

	initiatorCipher *noise.CipherState
	receiverCipher  *noise.CipherState

	peer io.ReadWriter
}

type NoiseStreamInitiator NoiseStreamStandard
type NoiseStreamReceiver NoiseStreamStandard

func NewNoiseStreamInitiator(myKey *noise.DHKey, peerKey, preSharedKey []byte) (*NoiseStreamInitiator, error) {
	cs := noise.NewCipherSuite(noise.DH25519, noise.CipherChaChaPoly, noise.HashBLAKE2b)

	hs, err := noise.NewHandshakeState(noise.Config{
		CipherSuite:   cs,
		Pattern:       noise.HandshakeNN,
		Initiator:     true,
		StaticKeypair: *myKey,
		PeerStatic:    peerKey,
		PresharedKey:  preSharedKey,
	})

	return &NoiseStreamInitiator{
		cipherSuite:    cs,
		handshakeState: hs,
		state:          E,
	}, err
}

func NewNoiseStreamReceiver(myKey *noise.DHKey, peerKey, preSharedKey []byte) (*NoiseStreamReceiver, error) {
	cs := noise.NewCipherSuite(noise.DH25519, noise.CipherChaChaPoly, noise.HashBLAKE2b)

	hs, err := noise.NewHandshakeState(noise.Config{
		CipherSuite:   cs,
		Pattern:       noise.HandshakeNN,
		StaticKeypair: *myKey,
		PeerStatic:    peerKey,
		PresharedKey:  preSharedKey,
	})

	return &NoiseStreamReceiver{
		cipherSuite:    cs,
		handshakeState: hs,
		state:          E,
	}, err
}

func (n *NoiseStreamInitiator) SetPeer(peer io.ReadWriter) {
	n.peer = peer
}

func (n *NoiseStreamInitiator) Encrypt(b []byte) ([]byte, error) {
	return n.initiatorCipher.Encrypt(nil, nil, b)
}

func (n *NoiseStreamInitiator) Decrypt(b []byte) ([]byte, error) {
	return n.receiverCipher.Decrypt(nil, nil, b)
}

func (n *NoiseStreamInitiator) Read(b []byte) (int, error) {
	if n.state != READY {
		return 0, fmt.Errorf("stream didn't handshake yet")
	}
	readBytes, err := n.peer.Read(b)
	if err != nil {
		return 0, err
	}
	b, err = n.receiverCipher.Decrypt(nil, nil, b[:readBytes])
	if err != nil {
		return 0, err
	}

	return len(b), nil
}

func (n *NoiseStreamInitiator) Write(b []byte) (int, error) {
	if n.state != READY {
		return 0, fmt.Errorf("stream didn't handshake yet")
	}

	msg, err := n.initiatorCipher.Encrypt(nil, nil, b)
	if err != nil {
		return 0, err
	}

	_, err = n.peer.Write(msg)
	if err != nil {
		return 0, err
	}

	return len(b), nil
}

func (n *NoiseStreamReceiver) SetPeer(peer io.ReadWriter) {
	n.peer = peer
}

func (n *NoiseStreamReceiver) Encrypt(b []byte) ([]byte, error) {
	return n.receiverCipher.Encrypt(nil, nil, b)
}

func (n *NoiseStreamReceiver) Decrypt(b []byte) ([]byte, error) {
	return n.initiatorCipher.Decrypt(nil, nil, b)
}

func (n *NoiseStreamReceiver) Read(b []byte) (int, error) {
	if n.state != READY {
		return 0, fmt.Errorf("stream didn't handshake yet")
	}
	readBytes, err := n.peer.Read(b)
	if err != nil {
		return 0, err
	}

	msg, err := n.initiatorCipher.Decrypt(nil, nil, b[:readBytes])
	if err != nil {
		return 0, err
	}

	return copy(b, msg), nil
}

func (n *NoiseStreamReceiver) Write(b []byte) (int, error) {
	if n.state != READY {
		return 0, fmt.Errorf("stream didn't handshake yet")
	}
	msg, err := n.receiverCipher.Encrypt(nil, nil, b)
	if err != nil {
		return 0, err
	}

	return n.peer.Write(msg)
}

func (n *NoiseStreamInitiator) DoHandshake(msg []byte) ([]byte, error) {
	var err error
	for err == nil {
		switch n.state {
		case E:
			replyMsg, _, _, err := n.handshakeState.WriteMessage(nil, nil)
			if err != nil {
				return nil, err
			}
			n.state = DHEE
			return replyMsg, nil
		case DHEE:
			if err != nil {
				return nil, err
			}
			var replyMsg []byte
			replyMsg, n.initiatorCipher, n.receiverCipher, err = n.handshakeState.ReadMessage(nil, msg)
			if err != nil {
				return nil, err
			}
			if len(replyMsg) != 0 {
				return nil, fmt.Errorf("handshake failed brutally initiator")
			}

			n.state = READY
		case READY:
			return nil, err
		}
	}
	return nil, err
}

func (n *NoiseStreamReceiver) DoHandshake(msg []byte) ([]byte, error) {
	var replyMsg []byte
	switch n.state {
	case E:
		_, _, _, err := n.handshakeState.ReadMessage(nil, msg)
		if err != nil {
			return nil, err
		}
		replyMsg, n.initiatorCipher, n.receiverCipher, err = n.handshakeState.WriteMessage(nil, nil)
		if err != nil {
			return nil, err
		}
		n.state = READY
	}
	return replyMsg, nil
}

func (n *NoiseStreamReceiver) IsReady() bool {
	return n.state == READY
}

func (n *NoiseStreamInitiator) IsReady() bool {
	return n.state == READY
}
