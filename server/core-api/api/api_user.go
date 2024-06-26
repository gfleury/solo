/*
 *
 * solo Server API
 *
 */
package api

import (
	"encoding/json"
	"net/http"

	"github.com/gfleury/solo/common/models"
	"github.com/gfleury/solo/server/core-api/db"
	"github.com/gfleury/solo/server/core-api/jwt"
)

func CreateUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
}

func DeleteUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
}

func GetUserByName(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
}

func LoginUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	db_handler := db.GetDB(r.Context())

	user_email := jwt.GetEmailFromClaim(r.Context())

	user := models.User{
		Username: user_email,
		Email:    user_email,
		Avatar:   jwt.GetAvatarFromClaim(r.Context()),
	}

	tx := db_handler.Where(models.User{Email: user_email}).FirstOrCreate(&user)

	if tx.Error != nil {
		http.Error(w, tx.Error.Error(), http.StatusBadRequest)
		return
	}

	b, err := json.Marshal(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

func LogoutUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
}

func UpdateUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
}
