module github.com/gfleury/solo/server/oauth2

go 1.19

replace github.com/gfleury/solo/server/core-api => ../core-api

require (
	github.com/alexedwards/scs/v2 v2.5.1
	github.com/golang-jwt/jwt v3.2.2+incompatible
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/pat v1.0.1
	github.com/markbates/goth v1.77.0
	github.com/rs/cors v1.9.0
)

require (
	cloud.google.com/go/compute v1.19.0 // indirect
	cloud.google.com/go/compute/metadata v0.2.3 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/gorilla/context v1.1.1 // indirect
	github.com/mrjones/oauth v0.0.0-20180629183705-f4e24b6d100c // indirect
	github.com/stretchr/testify v1.8.4 // indirect
	golang.org/x/net v0.12.0 // indirect
	golang.org/x/oauth2 v0.7.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
)
