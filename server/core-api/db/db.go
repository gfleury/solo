package db

import (
	"context"
	"flag"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gfleury/solo/common/models"
	"github.com/testcontainers/testcontainers-go"
	testpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DBCtx string

var DB DBCtx = "DB"

var db *gorm.DB

var POSTGRESCONTAINER *testpostgres.PostgresContainer

func DestroyTestPostgresContainer(ctx context.Context) {
	if err := POSTGRESCONTAINER.Terminate(ctx); err != nil {
		panic(err)
	}
}

func init() {
	var err error
	var dsn string

	if os.Getenv("DEV") != "" || strings.Contains(flag.CommandLine.Name(), "debug") || strings.Contains(flag.CommandLine.Name(), "test") {
		ctx := context.Background()
		POSTGRESCONTAINER, err = testpostgres.RunContainer(ctx,
			testcontainers.WithImage("postgres"),
			testpostgres.WithDatabase("core_api"),
			testpostgres.WithUsername("postgres"),
			testpostgres.WithPassword("mysecretpassword"),
			testcontainers.WithWaitStrategy(
				wait.ForLog("database system is ready to accept connections").
					WithOccurrence(2).
					WithStartupTimeout(5*time.Second)),
		)
		if err != nil {
			panic(err)
		}

		dsn, err = POSTGRESCONTAINER.ConnectionString(ctx, "")
		if err != nil {
			panic(err)
		}
	} else {
		dsn = "postgres://postgres:mysecretpassword@localhost/core_api"
	}
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	// Migrate the schema
	err = db.AutoMigrate(&models.Network{}, &models.NetworkNode{}, &models.LinkedUser{}, &models.User{}, &models.RegistrationRequest{})
	if err != nil {
		panic(err)
	}

}

func SetDBMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		timeoutContext, _ := context.WithTimeout(context.Background(), time.Second)
		ctx := context.WithValue(r.Context(), DB, db.WithContext(timeoutContext))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetDB(ctx context.Context) *gorm.DB {
	dbx, _ := ctx.Value(DB).(*gorm.DB)
	return dbx
}

func NonProtectedDB() *gorm.DB {
	return db
}
