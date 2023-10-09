package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gfleury/solo/common/models"
	"github.com/gfleury/solo/server/core-api/db"
	"github.com/libp2p/go-libp2p"
	check "gopkg.in/check.v1"
)

type R struct {
	muxer http.Handler
}

func TestRegistrationApi(t *testing.T) {
	check.TestingT(t)
}

func (s *R) SetUpSuite(c *check.C) {
	router := NewP2PRouter()
	router.Use(db.SetDBMiddleware)
	s.muxer = router
}

func (s *R) TearDownSuite(c *check.C) {
	defer db.DestroyTestPostgresContainer(context.Background())
}

func (s *R) TestRegisterNode(c *check.C) {
	host, err := libp2p.New()
	c.Assert(err, check.IsNil)

	node := models.NewLocalNode(host, "")

	j, err := node.Json()
	c.Assert(err, check.IsNil)

	body := strings.NewReader(string(j))
	request, err := http.NewRequest("POST", "/api/v1/register", body)
	c.Assert(err, check.IsNil)

	recorder := httptest.NewRecorder()
	s.muxer.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusCreated)

	time.Sleep(500 * time.Millisecond)
	result := db.NonProtectedDB().Find(node)
	c.Assert(result.Error, check.IsNil)
	c.Assert(result.RowsAffected, check.Equals, 1)
}
