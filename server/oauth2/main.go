package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"sort"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/gfleury/solo/server/oauth2/session"
	"github.com/gorilla/pat"
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/amazon"
	"github.com/markbates/goth/providers/facebook"
	"github.com/markbates/goth/providers/github"
	"github.com/markbates/goth/providers/google"
	"github.com/markbates/goth/providers/instagram"
	"github.com/markbates/goth/providers/linkedin"
	"github.com/markbates/goth/providers/openidConnect"
	"github.com/markbates/goth/providers/tiktok"
	"github.com/markbates/goth/providers/twitter"
	"github.com/markbates/goth/providers/twitterv2"
	"github.com/rs/cors"
)

var MY_LOCAL_URL = "http://localhost:8080"
var WEB_LOCAL_URL = "http://localhost:3000"

var MY_URL = "https://oauth.fleury.gg"
var WEB_URL = "https://web.fleury.gg"

func init() {
	if os.Getenv("DEV") != "" {
		MY_URL = MY_LOCAL_URL
		WEB_URL = WEB_LOCAL_URL
	}
}

func GetCallbackURL(path string) string {
	return fmt.Sprintf("%s%s", MY_URL, path)
}

func main() {
	goth.UseProviders(
		// Use twitterv2 instead of twitter if you only have access to the Essential API Level
		// the twitter provider uses a v1.1 API that is not available to the Essential Level
		twitterv2.New(os.Getenv("TWITTER_KEY"), os.Getenv("TWITTER_SECRET"), GetCallbackURL("/auth/twitterv2/callback")),
		// If you'd like to use authenticate instead of authorize in TwitterV2 provider, use this instead.
		// twitterv2.NewAuthenticate(os.Getenv("TWITTER_KEY"), os.Getenv("TWITTER_SECRET"), GetCallbackURL("/auth/twitterv2/callback")),

		twitter.New(os.Getenv("TWITTER_KEY"), os.Getenv("TWITTER_SECRET"), GetCallbackURL("/auth/twitter/callback")),
		// If you'd like to use authenticate instead of authorize in Twitter provider, use this instead.
		// twitter.NewAuthenticate(os.Getenv("TWITTER_KEY"), os.Getenv("TWITTER_SECRET"), GetCallbackURL("/auth/twitter/callback")),

		tiktok.New(os.Getenv("TIKTOK_KEY"), os.Getenv("TIKTOK_SECRET"), GetCallbackURL("/auth/tiktok/callback")),
		facebook.New(os.Getenv("FACEBOOK_KEY"), os.Getenv("FACEBOOK_SECRET"), GetCallbackURL("/auth/facebook/callback"), "instagram_basic", "pages_show_list", "business_management"),
		google.New(os.Getenv("GOOGLE_KEY"), os.Getenv("GOOGLE_SECRET"), GetCallbackURL("/auth/google/callback"), []string{"https://www.googleapis.com/auth/userinfo.profile", "https://www.googleapis.com/auth/userinfo.email"}...),
		github.New(os.Getenv("GITHUB_KEY"), os.Getenv("GITHUB_SECRET"), GetCallbackURL("/auth/github/callback")),
		linkedin.New(os.Getenv("LINKEDIN_KEY"), os.Getenv("LINKEDIN_SECRET"), GetCallbackURL("/auth/linkedin/callback")),
		instagram.New(os.Getenv("INSTAGRAM_KEY"), os.Getenv("INSTAGRAM_SECRET"), GetCallbackURL("/auth/instagram/callback")),
		amazon.New(os.Getenv("AMAZON_KEY"), os.Getenv("AMAZON_SECRET"), GetCallbackURL("/auth/amazon/callback")),
	)

	// OpenID Connect is based on OpenID Connect Auto Discovery URL (https://openid.net/specs/openid-connect-discovery-1_0-17.html)
	// because the OpenID Connect provider initialize itself in the New(), it can return an error which should be handled or ignored
	// ignore the error for now
	openidConnect, _ := openidConnect.New(os.Getenv("OPENID_CONNECT_KEY"), os.Getenv("OPENID_CONNECT_SECRET"), GetCallbackURL("/auth/openid-connect/callback"), os.Getenv("OPENID_CONNECT_DISCOVERY_URL"))
	if openidConnect != nil {
		goth.UseProviders(openidConnect)
	}

	m := map[string]string{
		"amazon":         "Amazon",
		"facebook":       "Facebook",
		"github":         "Github",
		"google":         "Google",
		"instagram":      "Instagram",
		"linkedin":       "LinkedIn",
		"openid-connect": "OpenID Connect",
		"tiktok":         "TikTok",
		"twitter":        "Twitter",
		"twitterv2":      "Twitter",
	}
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	providerIndex := &ProviderIndex{Providers: keys, ProvidersMap: m}

	p := pat.New()
	p.Get("/auth/{provider}/callback", func(res http.ResponseWriter, req *http.Request) {

		_, err := session.CompleteUserAuth(res, req)
		if err != nil {
			fmt.Fprintln(res, err)
			return
		}
		reqState := session.GetState(req)
		res.Header().Set("Location", fmt.Sprintf("%s/?state=%s", WEB_URL, reqState))
		res.WriteHeader(http.StatusTemporaryRedirect)
	})

	p.Get("/logout/{provider}", func(res http.ResponseWriter, req *http.Request) {
		session.Logout(res, req)
		res.Header().Set("Location", "/")
		res.WriteHeader(http.StatusTemporaryRedirect)
	})

	p.Get("/auth/{provider}", func(res http.ResponseWriter, req *http.Request) {
		// try to get the user without re-authenticating
		if gothUser, err := session.CompleteUserAuth(res, req); err == nil {
			t, _ := template.New("foo").Parse(userTemplate)
			t.Execute(res, gothUser)
		} else {
			fmt.Println(err)
			session.BeginAuthHandler(res, req)
		}
	})

	p.Get("/token", func(res http.ResponseWriter, req *http.Request) {
		if gothUser, err := session.CompleteUserAuth(res, req); err == nil {
			token, err := session.AddJwtCookie(gothUser, res)
			if err != nil {
				http.Error(res, err.Error(), http.StatusBadRequest)
				return
			}
			json_token, err := json.Marshal(&map[string]string{"token": token})
			if err != nil {
				http.Error(res, err.Error(), http.StatusBadRequest)
				return
			}
			res.Header().Set("Content-Type", "application/json; charset=UTF-8")
			res.Write([]byte(json_token))
		} else {
			http.Error(res, err.Error(), http.StatusUnauthorized)
		}
	})

	p.Get("/", func(res http.ResponseWriter, req *http.Request) {
		t, _ := template.New("foo").Parse(indexTemplate)
		t.Execute(res, providerIndex)
	})

	session.SessionManager = scs.New()
	session.SessionManager.Lifetime = 31 * 24 * time.Hour
	session.SessionManager.Cookie.Domain = session.CookieDomain
	session.SessionManager.Cookie.HttpOnly = false
	session.SessionManager.Cookie.Secure = session.CookieSecure
	session.SessionManager.Cookie.SameSite = http.SameSiteLaxMode

	log.Println("listening on localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", cors.New(cors.Options{
		AllowedOrigins: []string{"https://oauth.fleury.gg", "https://web.fleury.gg", "https://core-api.fleury.gg", "http://localhost:8080", "http://localhost:3000", "http://localhost:8081"},
		AllowedMethods: []string{
			http.MethodHead,
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
		},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	}).Handler(session.SessionManager.LoadAndSave(p))))

}

type ProviderIndex struct {
	Providers    []string
	ProvidersMap map[string]string
}

var indexTemplate = `{{range $key,$value:=.Providers}}
    <p><a href="/auth/{{$value}}">Log in with {{index $.ProvidersMap $value}}</a></p>
{{end}}`

var userTemplate = `
<p><a href="/logout/{{.Provider}}">logout</a></p>
<p>Name: {{.Name}} [{{.LastName}}, {{.FirstName}}]</p>
<p>Email: {{.Email}}</p>
<p>NickName: {{.NickName}}</p>
<p>Location: {{.Location}}</p>
<p>AvatarURL: {{.AvatarURL}} <img src="{{.AvatarURL}}"></p>
<p>Description: {{.Description}}</p>
<p>UserID: {{.UserID}}</p>
<p>AccessToken: {{.AccessToken}}</p>
<p>ExpiresAt: {{.ExpiresAt}}</p>
<p>RefreshToken: {{.RefreshToken}}</p>
`
