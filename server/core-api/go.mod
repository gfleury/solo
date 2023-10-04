module github.com/gfleury/solo/server/core-api

go 1.20

replace github.com/gfleury/solo/server/oauth2 => ../oauth2

require (
	github.com/Davincible/goinsta v0.0.0-20220425072628-96aad7267204
	github.com/golang-jwt/jwt v3.2.2+incompatible
	github.com/gorilla/mux v1.8.0
	github.com/huandu/facebook/v2 v2.7.0
	github.com/markbates/goth v1.77.0
	github.com/rs/cors v1.9.0
	github.com/gfleury/solo/server/oauth2 v0.0.0
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c
	gorm.io/driver/postgres v1.5.2
	gorm.io/driver/sqlite v1.5.2
	gorm.io/gorm v1.25.2
)

require (
	github.com/alexedwards/scs/v2 v2.5.1 // indirect
	github.com/chromedp/cdproto v0.0.0-20230722233645-dbf72f61037f // indirect
	github.com/chromedp/chromedp v0.9.1 // indirect
	github.com/chromedp/sysutil v1.0.0 // indirect
	github.com/gobwas/httphead v0.1.0 // indirect
	github.com/gobwas/pool v0.2.1 // indirect
	github.com/gobwas/ws v1.2.1 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/pgx/v5 v5.3.1 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-sqlite3 v1.14.17 // indirect
	github.com/rogpeppe/go-internal v1.9.0 // indirect
	golang.org/x/crypto v0.11.0 // indirect
	golang.org/x/net v0.12.0 // indirect
	golang.org/x/oauth2 v0.7.0 // indirect
	golang.org/x/sys v0.10.0 // indirect
	golang.org/x/text v0.11.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
)
