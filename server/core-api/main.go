/*
 *
 * solo Server API
 *
 */
package main

import (
	"context"
	"log"
	"net/http"

	"github.com/gfleury/solo/server/core-api/api"
	"github.com/gfleury/solo/server/core-api/db"
	"github.com/gfleury/solo/server/core-api/jwt"
	"github.com/gfleury/solo/server/core-api/rendezvous"
	"github.com/rs/cors"
)

func main() {
	log.Printf("Server started")
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	p2pRouter := api.NewP2PRouter()

	p2pRouter.Use(db.SetDBMiddleware)

	// Starts the libp2p host/network (rendezvous + registration)
	l, err := rendezvous.StartRendezvous(ctx, false)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		err := http.Serve(l.HTTPListener, p2pRouter)
		if err != nil {
			log.Fatal(err)
		}
	}()

	router := api.NewRouter()

	router.Use(db.SetDBMiddleware)

	log.Fatal(http.ListenAndServe(":8081", cors.New(cors.Options{
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
	}).Handler(jwt.VerifyJWT(router))))
}
