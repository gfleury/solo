package session

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/golang-jwt/jwt"
	"github.com/gorilla/mux"
	"github.com/markbates/goth"
)

type key int

const ProviderParamKey key = iota

var SessionManager *scs.SessionManager

var JWT_SECRET = []byte("<add-a-real-secret-here>")

var CookieDomain = "fleury.gg"
var CookieSecure = true

var ErrNoSession = fmt.Errorf("no session found")
var ErrNoProvider = fmt.Errorf("you must select a provider")

func init() {
	if os.Getenv("DEV") != "" {
		CookieDomain = "localhost"
		CookieSecure = false
	}
}

/*
BeginAuthHandler is a convenience handler for starting the authentication process.
It expects to be able to get the name of the provider from the query parameters
as either "provider" or ":provider".

BeginAuthHandler will redirect the user to the appropriate authentication end-point
for the requested provider.

See https://github.com/markbates/goth/examples/main.go to see this in action.
*/
func BeginAuthHandler(res http.ResponseWriter, req *http.Request) {
	url, err := GetAuthURL(res, req)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(res, err)
		return
	}

	http.Redirect(res, req, url, http.StatusTemporaryRedirect)
}

// SetState sets the state string associated with the given request.
// If no state string is associated with the request, one will be generated.
// This state is sent to the provider and can be retrieved during the
// callback.
var SetState = func(req *http.Request) string {
	state := req.URL.Query().Get("state")
	if len(state) > 0 {
		return state
	}

	// If a state query param is not passed in, generate a random
	// base64-encoded nonce so that the state on the auth URL
	// is unguessable, preventing CSRF attacks, as described in
	//
	// https://auth0.com/docs/protocols/oauth2/oauth-state#keep-reading
	nonceBytes := make([]byte, 64)
	_, err := io.ReadFull(rand.Reader, nonceBytes)
	if err != nil {
		panic("gothic: source of randomness unavailable: " + err.Error())
	}
	return base64.URLEncoding.EncodeToString(nonceBytes)
}

// GetState gets the state returned by the provider during the callback.
// This is used to prevent CSRF attacks, see
// http://tools.ietf.org/html/rfc6749#section-10.12
var GetState = func(req *http.Request) string {
	params := req.URL.Query()
	if params.Encode() == "" && req.Method == http.MethodPost {
		return req.FormValue("state")
	}
	return params.Get("state")
}

/*
GetAuthURL starts the authentication process with the requested provided.
It will return a URL that should be used to send users to.

It expects to be able to get the name of the provider from the query parameters
as either "provider" or ":provider".

I would recommend using the BeginAuthHandler instead of doing all of these steps
yourself, but that's entirely up to you.
*/
func GetAuthURL(res http.ResponseWriter, req *http.Request) (string, error) {
	providerName, err := GetProviderName(req)
	if err != nil {
		return "", err
	}

	provider, err := goth.GetProvider(providerName)
	if err != nil {
		return "", err
	}
	sess, err := provider.BeginAuth(SetState(req))
	if err != nil {
		return "", err
	}

	url, err := sess.GetAuthURL()
	if err != nil {
		return "", err
	}

	SessionManager.Put(req.Context(), providerName, sess.Marshal())

	if err != nil {
		return "", err
	}

	return url, err
}

/*
CompleteUserAuth does what it says on the tin. It completes the authentication
process and fetches all the basic information about the user from the provider.

It expects to be able to get the name of the provider from the query parameters
as either "provider" or ":provider".

See https://github.com/markbates/goth/examples/main.go to see this in action.
*/
var CompleteUserAuth = func(res http.ResponseWriter, req *http.Request) (goth.User, error) {
	providerName, err := GetProviderName(req)
	if err != nil {
		return goth.User{}, err
	}

	provider, err := goth.GetProvider(providerName)
	if err != nil {
		return goth.User{}, err
	}

	value := SessionManager.GetString(req.Context(), providerName)
	if value == "" {
		return goth.User{}, ErrNoSession
	}

	sess, err := provider.UnmarshalSession(value)
	if err != nil {
		return goth.User{}, err
	}

	err = ValidateState(req, sess)
	if err != nil {
		return goth.User{}, err
	}

	user, err := provider.FetchUser(sess)
	if err == nil {
		// user can be found with existing session data
		return user, err
	}

	params := req.URL.Query()
	if params.Encode() == "" && req.Method == "POST" {
		req.ParseForm()
		params = req.Form
	}

	// get new token and retry fetch
	_, err = sess.Authorize(provider, params)
	if err != nil {
		return goth.User{}, err
	}

	SessionManager.Put(req.Context(), providerName, sess.Marshal())

	gu, err := provider.FetchUser(sess)
	if err != nil {
		return goth.User{}, err
	}

	_, err = AddJwtCookie(gu, res)

	return gu, err
}

func AddJwtCookie(user goth.User, res http.ResponseWriter) (string, error) {
	claims := jwt.MapClaims{}
	claims["exp"] = user.ExpiresAt.UTC().Unix()
	claims["authorized"] = true
	claims["user"] = user.UserID
	claims["email"] = user.Email
	claims["name"] = user.Name
	claims["avatar"] = user.AvatarURL
	claims["provider"] = user.Provider
	claims["sub"] = user.UserID
	claims["iat"] = time.Now().UTC().Unix()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(JWT_SECRET)
	if err != nil {
		return "", err
	}

	cookie := http.Cookie{
		Name:     "JWT",
		Value:    tokenString,
		Path:     "/",
		Domain:   CookieDomain,
		HttpOnly: false,
		Secure:   CookieSecure,
		SameSite: http.SameSiteLaxMode,
	}

	// Use the http.SetCookie() function to send the cookie to the client.
	// Behind the scenes this adds a `Set-Cookie` header to the response
	// containing the necessary cookie data.
	http.SetCookie(res, &cookie)

	return tokenString, nil
}

// validateState ensures that the state token param from the original
// AuthURL matches the one included in the current (callback) request.
func ValidateState(req *http.Request, sess goth.Session) error {
	rawAuthURL, err := sess.GetAuthURL()
	if err != nil {
		return err
	}

	authURL, err := url.Parse(rawAuthURL)
	if err != nil {
		return err
	}

	reqState := GetState(req)

	originalState := authURL.Query().Get("state")
	if originalState != "" && (originalState != reqState) {
		return errors.New("state token mismatch")
	}
	return nil
}

// Logout invalidates a user session.
func Logout(res http.ResponseWriter, req *http.Request) error {
	return SessionManager.Destroy(req.Context())
}

// GetProviderName is a function used to get the name of a provider
// for a given request. By default, this provider is fetched from
// the URL query string. If you provide it in a different way,
// assign your own function to this variable that returns the provider
// name for your request.
var GetProviderName = getProviderName

func getProviderName(req *http.Request) (string, error) {

	// try to get it from the url param "provider"
	if p := req.URL.Query().Get("provider"); p != "" {
		return p, nil
	}

	// try to get it from the url param ":provider"
	if p := req.URL.Query().Get(":provider"); p != "" {
		return p, nil
	}

	// try to get it from the context's value of "provider" key
	if p, ok := mux.Vars(req)["provider"]; ok {
		return p, nil
	}

	//  try to get it from the go-context's value of "provider" key
	if p, ok := req.Context().Value("provider").(string); ok {
		return p, nil
	}

	// try to get it from the go-context's value of providerContextKey key
	if p, ok := req.Context().Value(ProviderParamKey).(string); ok {
		return p, nil
	}

	// As a fallback, loop over the used providers, if we already have a valid session for any provider (ie. user has already begun authentication with a provider), then return that provider name
	providers := goth.GetProviders()
	for _, provider := range providers {
		p := provider.Name()

		if SessionManager.Exists(req.Context(), p) {
			return p, nil
		}
	}

	// if not found then return an empty string with the corresponding error
	return "", ErrNoProvider
}

// GetContextWithProvider returns a new request context containing the provider
func GetContextWithProvider(req *http.Request, provider string) *http.Request {
	return req.WithContext(context.WithValue(req.Context(), ProviderParamKey, provider))
}
