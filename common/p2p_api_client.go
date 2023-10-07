package common

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gfleury/solo/common/models"
	p2phttp "github.com/libp2p/go-libp2p-http"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
)

var SoloAPIP2PProtocol protocol.ID = "/solo-api-http"

type SoloAPIP2PClient struct {
	client  *http.Client
	address string
	host    host.Host
}

func GetSoloAPIP2PClient(serverID peer.ID, clientHost host.Host) *SoloAPIP2PClient {
	tr := &http.Transport{}
	tr.RegisterProtocol("libp2p", p2phttp.NewTransport(clientHost, p2phttp.ProtocolOption(SoloAPIP2PProtocol)))
	client := &http.Client{Transport: tr}
	return &SoloAPIP2PClient{client: client, address: fmt.Sprintf("libp2p://%s", serverID), host: clientHost}
}

func (s *SoloAPIP2PClient) RegisterNode(node models.NetworkNode) (string, error) {
	b, err := json.Marshal(&node)
	if err != nil {
		return "", err
	}
	resp, err := s.client.Post(fmt.Sprintf("%s/api/v1/node/register", s.address), "application/json", bytes.NewReader(b))
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	if resp.StatusCode > 399 {
		return "", fmt.Errorf("HTTP Error: %s", resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	r := &RegistrationResponse{}

	err = json.Unmarshal(body, r)

	return r.Code, err
}

func (s *SoloAPIP2PClient) GetNodeNetworkConfiguration() (*ConnectionConfigurationResponse, int, error) {
	// Get host private key
	rawPrivKey, err := s.host.Peerstore().PrivKey(s.host.ID()).Raw()
	if err != nil {
		return nil, 0, err
	}
	privKey := ed25519.PrivateKey(rawPrivKey)

	message := NodeAuthenticationTokenMessage()
	// Sign authenticationtoken message
	rawAuthenticationToken, err := privKey.Sign(nil, message, NodeAuthenticationTokenOptions)
	if err != nil {
		return nil, 0, err
	}

	// Verify message just in case
	pubKey := privKey.Public().(ed25519.PublicKey)
	err = ed25519.VerifyWithOptions(pubKey, message, rawAuthenticationToken, NodeAuthenticationTokenOptions)
	if err != nil {
		return nil, 0, err
	}

	// base64 encode token
	nodeAuthenticationToken := base64.RawStdEncoding.EncodeToString(rawAuthenticationToken)

	request := ConnectionConfigurationRequest{
		PeerID:                  s.host.ID().String(),
		NodeAuthenticationToken: nodeAuthenticationToken,
	}
	b, err := json.Marshal(&request)
	if err != nil {
		return nil, 0, err
	}
	resp, err := s.client.Post(fmt.Sprintf("%s/api/v1/node/connnection_configuration", s.address), "application/json", bytes.NewReader(b))
	if err != nil {
		return nil, resp.StatusCode, err
	}

	defer resp.Body.Close()

	if resp.StatusCode > 399 {
		return nil, resp.StatusCode, fmt.Errorf("HTTP Error: %s", resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}

	r := &ConnectionConfigurationResponse{}

	err = json.Unmarshal(body, r)

	return r, resp.StatusCode, err
}
