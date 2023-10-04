package db

import (
	"context"
	"flag"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gfleury/solo/server/core-api/models"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type DBCtx string

var DB DBCtx = "DB"

var db *gorm.DB

var TEST_DB = "test.db"

func init() {
	var err error

	if os.Getenv("DEV") != "" || strings.Contains(flag.CommandLine.Name(), "debug") || strings.Contains(flag.CommandLine.Name(), "test") {
		db, err = gorm.Open(sqlite.Open(TEST_DB), &gorm.Config{})
	} else {
		dsn := "postgres://postgres:mysecretpassword@localhost/core_api"
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	}
	if err != nil {
		panic(err)
	}

	// Migrate the schema
	err = db.AutoMigrate(&models.Network{}, &models.Node{}, &models.LinkedUser{}, &models.User{})
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
