package noise

import (
	"bytes"
	"crypto/rand"
	"io"
	"testing"

	"github.com/flynn/noise"
	"github.com/gfleury/solo/client/utils"
	"github.com/stretchr/testify/require"
)

func TestNoiseStream(t *testing.T) {

	kI, err := noise.DH25519.GenerateKeypair(rand.Reader)
	require.NoError(t, err)

	kR, err := noise.DH25519.GenerateKeypair(rand.Reader)
	require.NoError(t, err)

	i, err := NewNoiseStreamInitiator(&kI, kR.Public, []byte("supersecretsupersecretsupersecre"))
	require.NoError(t, err)
	r, err := NewNoiseStreamReceiver(&kR, kI.Public, []byte("supersecretsupersecretsupersecre"))
	require.NoError(t, err)

	dataStream1, dataStream2 := utils.NewTestConnection()

	replyMsg, err := i.DoHandshake(nil)
	require.NoError(t, err)

	replyMsg, err = r.DoHandshake(replyMsg)
	require.NoError(t, err)

	_, err = i.DoHandshake(replyMsg)
	require.NoError(t, err)

	for i.state != READY {
	}

	i.SetPeer(dataStream1)
	r.SetPeer(dataStream2)

	msg := []byte("ALO")

	go func() {
		i.Write(msg)
	}()

	readMsg := make([]byte, 50)
	n, err := r.Read(readMsg)
	require.NoError(t, err)

	require.ElementsMatch(t, msg, readMsg[:n])

	require.Equal(t, len(msg), len(readMsg[:n]))

	msg = []byte("NEWMSGTEST")
	go func() {
		i.Write(msg)
	}()

	readMsg = make([]byte, 50)
	n, err = r.Read(readMsg)
	require.NoError(t, err)

	require.ElementsMatch(t, msg, readMsg[:n])

	require.Equal(t, len(msg), len(readMsg[:n]))

	msg = []byte("NEWMSGTEST")
	go func() {
		i.Write(msg)
	}()

	readMsg = make([]byte, 100)
	n, err = r.Read(readMsg)
	require.NoError(t, err)

	require.ElementsMatch(t, msg, readMsg[:n])

	require.Equal(t, len(msg), len(readMsg[:n]))

}

func TestNoiseWrongOrderStream(t *testing.T) {

	kI, err := noise.DH25519.GenerateKeypair(rand.Reader)
	require.NoError(t, err)

	kR, err := noise.DH25519.GenerateKeypair(rand.Reader)
	require.NoError(t, err)

	i, err := NewNoiseStreamInitiator(&kI, kR.Public, []byte("supersecretsupersecretsupersecre"))
	require.NoError(t, err)
	r, err := NewNoiseStreamReceiver(&kR, kI.Public, []byte("supersecretsupersecretsupersecre"))
	require.NoError(t, err)
	stream := bytes.NewBuffer([]byte{})

	i.SetPeer(stream)
	r.SetPeer(stream)

	msg := []byte("ALO")
	_, err = r.Write(msg)
	require.Error(t, err)

	readMsg := make([]byte, 10)
	_, err = i.Read(readMsg)
	require.Error(t, err)

}

func TestNoiseWithBidi(t *testing.T) {

	kI, err := noise.DH25519.GenerateKeypair(rand.Reader)
	require.NoError(t, err)

	kR, err := noise.DH25519.GenerateKeypair(rand.Reader)
	require.NoError(t, err)

	i, err := NewNoiseStreamInitiator(&kI, kR.Public, []byte("supersecretsupersecretsupersecre"))
	require.NoError(t, err)
	r, err := NewNoiseStreamReceiver(&kR, kI.Public, []byte("supersecretsupersecretsupersecre"))
	require.NoError(t, err)

	dataStream1, dataStream2 := utils.NewTestConnection()

	replyMsg, err := i.DoHandshake(nil)
	require.NoError(t, err)

	replyMsg, err = r.DoHandshake(replyMsg)
	require.NoError(t, err)

	_, err = i.DoHandshake(replyMsg)
	require.NoError(t, err)

	for i.state != READY {
	}

	i.SetPeer(dataStream1)
	r.SetPeer(dataStream2)

	msg := []byte("ALO")

	go func() {
		i.Write(msg)
	}()

	readMsg := make([]byte, 100)
	n, err := r.Read(readMsg)
	require.NoError(t, err)

	require.ElementsMatch(t, msg, readMsg[:n])

	require.Equal(t, len(msg), len(readMsg[:n]))

}

func TestNoiseWithBidiWrongPSK(t *testing.T) {

	kI, err := noise.DH25519.GenerateKeypair(rand.Reader)
	require.NoError(t, err)

	kR, err := noise.DH25519.GenerateKeypair(rand.Reader)
	require.NoError(t, err)

	i, err := NewNoiseStreamInitiator(&kI, kR.Public, []byte("supersecretsupersecretsupersecrt"))
	require.NoError(t, err)
	r, err := NewNoiseStreamReceiver(&kR, kI.Public, []byte("supersecretsupersecretsupersecre"))
	require.NoError(t, err)

	replyMsg, err := i.DoHandshake(nil)
	require.NoError(t, err)

	_, err = r.DoHandshake(replyMsg)
	require.Error(t, err)

}

func TestBidi(t *testing.T) {
	r1, w1 := io.Pipe()
	r2, w2 := io.Pipe()

	c1 := utils.NewBidiPipe(r1, w2)
	c2 := utils.NewBidiPipe(r2, w1)

	orig_msg := []byte("Alo")
	go c1.Write(orig_msg)

	msg := make([]byte, 128)
	n, _ := c2.Read(msg)

	require.ElementsMatch(t, orig_msg, msg[:n])
}
