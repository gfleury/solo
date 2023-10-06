package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gfleury/solo/common/models"
	"github.com/gfleury/solo/server/core-api/db"
	ourjwt "github.com/gfleury/solo/server/core-api/jwt"
	"github.com/golang-jwt/jwt"
	check "gopkg.in/check.v1"
)

type S struct {
	muxer http.Handler
}

var test_user1 = models.User{
	Username: "testUser1",
	Email:    "testUser1@test.com",
}

var test_user2 = models.User{
	Username: "testUser2",
	Email:    "testUser2@test.com",
}

var test_network1 = models.Network{
	Name: "myTestNetwork2",
	CIDR: "10.0.0.1/24",
}

func setJWTTest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := jwt.MapClaims{
			"user":  test_user1.Username,
			"email": test_user1.Email,
		}
		ctx := context.WithValue(r.Context(), ourjwt.Claims, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
func Test(t *testing.T) {
	check.TestingT(t)
}

func (s *S) SetUpSuite(c *check.C) {
	router := NewRouter()
	router.Use(db.SetDBMiddleware)
	router.Use(setJWTTest)
	s.muxer = router

	result := db.NonProtectedDB().Create(&test_user1)
	c.Assert(result.Error, check.IsNil)

	result = db.NonProtectedDB().Create(&test_user2)
	c.Assert(result.Error, check.IsNil)

	result = db.NonProtectedDB().Create(&test_network1)
	c.Assert(result.Error, check.IsNil)
}

func (s *S) TearDownSuite(c *check.C) {
	defer os.Remove(db.TEST_DB)
}

func (s *S) TestAddNetwork(c *check.C) {
	network := models.NewNetwork("myTestNetwork", "10.1.0.0/3")

	j, err := network.Json()
	c.Assert(err, check.IsNil)

	body := strings.NewReader(string(j))
	request, err := http.NewRequest("POST", "/api/v1/network", body)
	c.Assert(err, check.IsNil)

	recorder := httptest.NewRecorder()
	s.muxer.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusCreated)

	time.Sleep(500 * time.Millisecond)
	result := db.NonProtectedDB().Find(network)
	c.Assert(result.Error, check.IsNil)
}

func (s *S) TestGetNetworkById(c *check.C) {
	var network models.Network

	result := db.NonProtectedDB().First(&network)
	c.Assert(result.Error, check.IsNil)

	request, err := http.NewRequest("GET", fmt.Sprintf("/api/v1/network/%d", network.ID), nil)
	c.Assert(err, check.IsNil)

	recorder := httptest.NewRecorder()
	s.muxer.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusOK)

}

func (s *S) TestDeleteNetwork(c *check.C) {
	var network models.Network

	result := db.NonProtectedDB().First(&network)
	c.Assert(result.Error, check.IsNil)

	request, err := http.NewRequest("DELETE", fmt.Sprintf("/api/v1/network/%d", network.ID), nil)
	c.Assert(err, check.IsNil)

	recorder := httptest.NewRecorder()
	s.muxer.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusNoContent)
}

func (s *S) TestGetNetworks(c *check.C) {

	network := models.Network{
		User: &test_user2,
		Name: "user2firstnetwork",
		CIDR: "10.0.0.1/24",
	}
	result := db.NonProtectedDB().Create(&network)
	c.Assert(result.Error, check.IsNil)

	request, err := http.NewRequest("GET", "/api/v1/networks", nil)
	c.Assert(err, check.IsNil)

	recorder := httptest.NewRecorder()
	s.muxer.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusOK)

	var networks []models.Network
	err = json.Unmarshal(recorder.Body.Bytes(), &networks)
	c.Assert(err, check.IsNil)

	for _, network := range networks {
		c.Assert(network.User.Email, check.Equals, test_user1.Email)
	}
}

func (s *S) TestUpdateNetwork(c *check.C) {
	network := models.NewNetwork("myTestNetworkUpdate", "10.1.0.0/3")

	network.User = &test_user1

	j, err := network.Json()
	c.Assert(err, check.IsNil)

	body := strings.NewReader(string(j))
	request, err := http.NewRequest("PUT", "/api/v1/network", body)
	c.Assert(err, check.IsNil)

	recorder := httptest.NewRecorder()
	s.muxer.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusCreated)
}

func (s *S) TestFetchNetworksUser1(c *check.C) {

	request, err := http.NewRequest("GET", "/api/v1/networks", nil)
	c.Assert(err, check.IsNil)

	recorder := httptest.NewRecorder()
	s.muxer.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusOK)
}
