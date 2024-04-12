package rendezvous

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gfleury/solo/client/node"
	"github.com/gfleury/solo/common"
	"github.com/gfleury/solo/common/models"
	"github.com/gfleury/solo/server/core-api/api"
	"github.com/gfleury/solo/server/core-api/db"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/require"
)

func TestRendzvousNodeRegisterHttpOverP2P(t *testing.T) {
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()
	defer db.DestroyTestPostgresContainer(ctx)

	r, err := StartRendezvous(ctx, true)
	require.NoError(t, err)

	p2pRouter := api.NewP2PRouter()
	p2pRouter.Use(db.SetDBMiddleware)

	go func() {
		err := http.Serve(r.HTTPListener, p2pRouter)
		if err != nil {
			require.NoError(t, err)
		}
	}()

	// Create the client host
	i := node.NewIdentityWithName("testFixed")
	priv, err := i.LoadOrGeneratePrivateKey(0)
	require.NoError(t, err)

	h, err := libp2p.New(libp2p.Identity(priv))
	require.NoError(t, err)

	// Connect two hosts
	err = h.Connect(ctx, peer.AddrInfo{ID: r.host.ID(), Addrs: r.host.Addrs()})
	require.NoError(t, err)

	// HTTP over p2p client
	client := common.GetSoloAPIP2PClient(r.host.ID(), h)

	m := models.NewLocalNode(h, "")

	// Make the request register Node request
	code, err := client.RegisterNode(m)
	require.NoError(t, err)

	require.NotEmpty(t, code)

	// Register normal Public API
	router := api.NewRouter()
	router.Use(db.SetDBMiddleware)

	configurationToken := models.GenerateNewConnectionData().Base64()
	// Create Dummy network
	network := &models.Network{
		Name:                  "DummyNetwork",
		CIDR:                  "10.1.0.1/24",
		ConnectionConfigToken: configurationToken,
		User: &models.User{
			Username: "Dummy",
			Email:    "dummy@gmail.com",
		},
	}
	result := db.NonProtectedDB().Create(network)
	require.NoError(t, result.Error)

	request, err := http.NewRequest("PUT", fmt.Sprintf("/api/v1/network/%d/register/%s", network.ID, code), nil)
	require.NoError(t, err)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	require.Equal(t, recorder.Result().StatusCode, http.StatusOK)

	// Make the request to fetch connection details
	connectionConfigurationResponse, statusCode, err := client.GetNodeNetworkConfiguration()
	require.NoError(t, err)
	require.Equal(t, statusCode, 200)

	require.Equal(t, connectionConfigurationResponse.ConnectionConfigToken, configurationToken)
	require.Equal(t, connectionConfigurationResponse.InterfaceAddress, "10.1.0.1/24")
}
