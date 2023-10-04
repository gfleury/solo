/*
 * Swagger solo - OpenAPI 3.0
 *
 * solo API
 *
 * API version: 1.0.0
 * Contact: apiteam@solo.io
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */
package main

import (
	"log"
	"net/http"

	"github.com/gfleury/solo/server/core-api/api"
	"github.com/gfleury/solo/server/core-api/db"
	"github.com/gfleury/solo/server/core-api/jwt"
	"github.com/rs/cors"
)

func main() {
	log.Printf("Server started")

	router := api.NewRouter()

	router.Use(db.SetDBMiddleware)

	log.Fatal(http.ListenAndServe(":8081", cors.New(cors.Options{
		AllowedOrigins: []string{"https://oauth2.fleury.gg", "https://web.fleury.gg", "https://core-api.fleury.gg", "http://localhost:8080", "http://localhost:3000", "http://localhost:8081"},
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
