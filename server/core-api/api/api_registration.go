package api

import (
	"encoding/json"
	"net/http"

	"github.com/gfleury/solo/common"
	"github.com/gfleury/solo/common/models"
	"github.com/gfleury/solo/server/core-api/db"
	"github.com/gfleury/solo/server/core-api/jwt"
	"github.com/gorilla/mux"
	"gorm.io/gorm/clause"
)

func GetNodeRegistration(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "No code specified", http.StatusBadRequest)
		return
	}

	db_handler := db.GetDB(r.Context())

	registration := &models.RegistrationRequest{
		Code: code,
	}

	result := db_handler.First(&registration)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusBadRequest)
		return
	}

	JsonResponse(registration, http.StatusCreated, w)
}

func RegisterNode(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	var n models.NetworkNode

	err := json.NewDecoder(r.Body).Decode(&n)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err = n.Valid(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	db_handler := db.GetDB(r.Context())

	// Find if there is another registration request pending
	result := db_handler.Where(&n).Find(&n)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusBadRequest)
		return
	}
	var registrationRequest *models.RegistrationRequest

	// Node exists now look for existing registration request
	if result.RowsAffected > 0 {
		registrationRequest = &models.RegistrationRequest{NodeID: n.ID}
		result := db_handler.Where(registrationRequest).First(registrationRequest)
		if result.Error != nil {
			http.Error(w, result.Error.Error(), http.StatusNotFound)
			return
		}
	} else {
		registrationRequest = &models.RegistrationRequest{Node: n}

		registrationRequest.Code, err = models.GenerateNewRandomCode(6)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		result = db_handler.Create(registrationRequest)
		if result.Error != nil {
			http.Error(w, result.Error.Error(), http.StatusBadRequest)
			return
		}
	}

	result = db_handler.Save(registrationRequest)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusBadRequest)
		return
	}

	registerResponse := &common.RegistrationResponse{Code: registrationRequest.Code}

	JsonResponse(registerResponse, http.StatusCreated, w)
}

func NetworkAssignNodeFromRegistrationCode(w http.ResponseWriter, r *http.Request) {
	var network models.Network
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	vars := mux.Vars(r)
	db_handler := db.GetDB(r.Context())

	result := db_handler.InnerJoins("User", db_handler.Where(models.User{Email: jwt.GetEmailFromClaim(r.Context())})).Preload(clause.Associations).First(&network, vars["networkId"])
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusBadRequest)
		return
	}

	code, codeFound := vars["code"]
	if !codeFound {
		http.Error(w, "No code specified", http.StatusBadRequest)
		return
	}

	registration := &models.RegistrationRequest{
		Code: code,
	}

	result = db_handler.Preload(clause.Associations).First(&registration)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusBadRequest)
		return
	}

	networkNode := registration.Node
	// Update node with new network ID that it belongs
	// and next free ip from network
	networkNode.NetworkID = &network.ID
	networkNode.IP = network.NextFreeIP()
	networkNode.Actived = true

	// Save node with networkID
	result = db_handler.Save(&networkNode)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusBadRequest)
		return
	}

	// Delete RegistrationRequest
	result = db_handler.Delete(registration)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "node successfully associated with network"}`))
}
