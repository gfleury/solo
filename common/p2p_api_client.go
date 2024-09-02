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

	// Get Challenge
	request := ConnectionConfigurationChallengeRequest{
		PeerID: s.host.ID().String(),
	}
	b, err := json.Marshal(&request)
	if err != nil {
		return nil, 0, err
	}
	resp, err := s.client.Post(fmt.Sprintf("%s/api/v1/node/connnection_configuration", s.address), "application/json", bytes.NewReader(b))
	if err != nil {
		code := 0
		if resp != nil {
			code = resp.StatusCode
		}
		return nil, code, err
	}
	if resp.StatusCode > 399 {
		return nil, resp.StatusCode, fmt.Errorf("HTTP Error: %s", resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}
	challengeResponse := &ConnectionConfigurationChallengeResponse{}
	err = json.Unmarshal(body, challengeResponse)
	if err != nil {
		return nil, resp.StatusCode, err
	}
	resp.Body.Close()

	// Sign authenticationtoken message
	rawSignedChallenge, err := privKey.Sign(nil, []byte(challengeResponse.Challenge), NodeAuthenticationTokenOptions)
	if err != nil {
		return nil, 0, err
	}

	// Verify message just in case
	pubKey := privKey.Public().(ed25519.PublicKey)
	err = ed25519.VerifyWithOptions(pubKey, []byte(challengeResponse.Challenge), rawSignedChallenge, NodeAuthenticationTokenOptions)
	if err != nil {
		return nil, 0, err
	}

	// base64 encode token
	signedChallenge := base64.RawStdEncoding.EncodeToString(rawSignedChallenge)

	configurationRequest := ConnectionConfigurationRequest{
		PeerID:          s.host.ID().String(),
		SignedChallenge: signedChallenge,
	}
	b, err = json.Marshal(&configurationRequest)
	if err != nil {
		return nil, 0, err
	}
	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/api/v1/node/connnection_configuration", s.address), bytes.NewReader(b))
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err = s.client.Do(req)
	if err != nil {
		return nil, resp.StatusCode, err
	}

	defer resp.Body.Close()

	if resp.StatusCode > 399 {
		return nil, resp.StatusCode, fmt.Errorf("HTTP Error: %s", resp.Status)
	}
	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}

	r := &ConnectionConfigurationResponse{}

	err = json.Unmarshal(body, r)

	return r, resp.StatusCode, err
}

func (s *SoloAPIP2PClient) UpdateNode(updateRequest NodeUpdateRequest) (int, error) {
	b, err := json.Marshal(&updateRequest)
	if err != nil {
		return 0, err
	}

	resp, err := s.client.Post(fmt.Sprintf("%s/api/v1/node", s.address), "application/json", bytes.NewReader(b))
	return resp.StatusCode, err
}
