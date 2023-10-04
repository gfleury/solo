package api

import (
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/gfleury/solo/server/core-api/db"
	ourjwt "github.com/gfleury/solo/server/core-api/jwt"
	"github.com/gfleury/solo/server/core-api/models"
	"github.com/golang-jwt/jwt"
	check "gopkg.in/check.v1"
)

var _ = check.Suite(&S{})

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

	// first_account := models.NewAccount("FirstAccount", models.NewProvider(models.Instagram), models.NewTags([]string{"main", "facebook"}))
	// fillLoginPassword(first_account)

	// result = db.NonProtectedDB().Create(first_account)
	// c.Assert(result.Error, check.IsNil)
}

func (s *S) TearDownSuite(c *check.C) {
	defer os.Remove(db.TEST_DB)
}

func (s *S) TestAddAccount(c *check.C) {
	// account := models.NewAccount("George", models.NewProvider(models.Instagram), models.NewTags([]string{"main", "instagram"}))
	// fillLoginPassword(account)

	// j, err := account.Json()
	// c.Assert(err, check.IsNil)

	// body := strings.NewReader(string(j))
	// request, err := http.NewRequest("POST", "/api/v1/account", body)
	// c.Assert(err, check.IsNil)

	// recorder := httptest.NewRecorder()
	// s.muxer.ServeHTTP(recorder, request)
	// c.Assert(recorder.Code, check.Equals, http.StatusCreated)

	// time.Sleep(500 * time.Millisecond)
	// result := db.NonProtectedDB().Find(account)
	// c.Assert(result.Error, check.IsNil)
	// c.Assert(account.Session, check.NotNil)
}

func (s *S) TestGetAccountById(c *check.C) {
	// var account models.Account

	// result := db.NonProtectedDB().First(&account)
	// c.Assert(result.Error, check.IsNil)

	// request, err := http.NewRequest("GET", fmt.Sprintf("/api/v1/account/%d", account.ID), nil)
	// c.Assert(err, check.IsNil)

	// recorder := httptest.NewRecorder()
	// s.muxer.ServeHTTP(recorder, request)
	// c.Assert(recorder.Code, check.Equals, http.StatusOK)

}

func (s *S) TestDeleteAccount(c *check.C) {
	// var account models.Account

	// result := db.NonProtectedDB().First(&account)
	// c.Assert(result.Error, check.IsNil)

	// request, err := http.NewRequest("DELETE", fmt.Sprintf("/api/v1/account/%d", account.ID), nil)
	// c.Assert(err, check.IsNil)

	// recorder := httptest.NewRecorder()
	// s.muxer.ServeHTTP(recorder, request)
	// c.Assert(recorder.Code, check.Equals, http.StatusNoContent)
}

func (s *S) TestGetAccounts(c *check.C) {

	// account := models.Account{
	// 	User:     &test_user2,
	// 	Name:     "user2firstaccount",
	// 	Login:    "user2firstaccount",
	// 	Provider: models.NewProvider(models.Instagram),
	// 	Password: "xxxx",
	// }
	// result := db.NonProtectedDB().Create(&account)
	// c.Assert(result.Error, check.IsNil)

	// request, err := http.NewRequest("GET", "/api/v1/accounts", nil)
	// c.Assert(err, check.IsNil)

	// recorder := httptest.NewRecorder()
	// s.muxer.ServeHTTP(recorder, request)
	// c.Assert(recorder.Code, check.Equals, http.StatusOK)

	// var accounts []models.Account
	// err = json.Unmarshal(recorder.Body.Bytes(), &accounts)
	// c.Assert(err, check.IsNil)

	// for _, account := range accounts {
	// 	c.Assert(account.User.Email, check.Equals, test_user1.Email)
	// }
}

func (s *S) TestUpdateAccount(c *check.C) {
	// account := models.NewAccount("George", models.NewProvider(models.Instagram), models.NewTags([]string{"main", "instagram"}))

	// account.User = &test_user1

	// j, err := account.Json()
	// c.Assert(err, check.IsNil)

	// body := strings.NewReader(string(j))
	// request, err := http.NewRequest("PUT", "/api/v1/account", body)
	// c.Assert(err, check.IsNil)

	// recorder := httptest.NewRecorder()
	// s.muxer.ServeHTTP(recorder, request)
	// c.Assert(recorder.Code, check.Equals, http.StatusCreated)
}

func (s *S) TestFetchAccountsUser1(c *check.C) {

	// request, err := http.NewRequest("GET", "/api/v1/accounts", nil)
	// c.Assert(err, check.IsNil)

	// recorder := httptest.NewRecorder()
	// s.muxer.ServeHTTP(recorder, request)
	// c.Assert(recorder.Code, check.Equals, http.StatusOK)
}
