package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/gfleury/solo/common/models"
	"github.com/gfleury/solo/server/core-api/db"
	"github.com/libp2p/go-libp2p"
	check "gopkg.in/check.v1"
)

func (s *S) TestRegisterNode(c *check.C) {
	host, err := libp2p.New()
	c.Assert(err, check.IsNil)

	node := models.NewLocalNode(host, "")

	j, err := node.Json()
	c.Assert(err, check.IsNil)

	body := strings.NewReader(string(j))
	request, err := http.NewRequest("POST", "/api/v1/node/register", body)
	c.Assert(err, check.IsNil)

	recorder := httptest.NewRecorder()
	s.p2pMuxer.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusCreated)

	time.Sleep(500 * time.Millisecond)
	result := db.NonProtectedDB().Find(&node)
	c.Assert(result.Error, check.IsNil)
	c.Assert(int(result.RowsAffected), check.Equals, 1)
}

func (s *S) TestUpdateNode(c *check.C) {
	host, err := libp2p.New()
	c.Assert(err, check.IsNil)

	node := models.NewLocalNode(host, "")

	result := db.NonProtectedDB().Create(&node)
	c.Assert(result.Error, check.IsNil)

	node.Hostname = "ThisIsANodeUpdateTest"

	j, err := node.Json()
	c.Assert(err, check.IsNil)

	body := strings.NewReader(string(j))
	request, err := http.NewRequest("POST", "/api/v1/node", body)
	c.Assert(err, check.IsNil)

	recorder := httptest.NewRecorder()
	s.p2pMuxer.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusCreated)

	time.Sleep(500 * time.Millisecond)
	result = db.NonProtectedDB().Find(&node)
	c.Assert(result.Error, check.IsNil)
	c.Assert(int(result.RowsAffected), check.Equals, 1)
}
