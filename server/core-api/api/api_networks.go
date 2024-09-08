/*
 *
 * solo Server API
 *
 */
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

func AddNetwork(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	var n models.Network

	err := json.NewDecoder(r.Body).Decode(&n)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err = n.Valid(); err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	db_handler := db.GetDB(r.Context())

	n.User = &models.User{
		Email: jwt.GetEmailFromClaim(r.Context()),
	}

	result := db_handler.Find(n.User)

	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusBadRequest)
		return
	}

	result = db_handler.Save(&n)

	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusBadRequest)
		return
	}

	b, err := n.Json()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write(b)
}

func DeleteNetwork(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	vars := mux.Vars(r)
	db_handler := db.GetDB(r.Context())

	result := db_handler.InnerJoins("User", db_handler.Where(models.User{Email: jwt.GetEmailFromClaim(r.Context())})).Delete(&models.Network{}, vars["networkId"])
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func GetNetworkById(w http.ResponseWriter, r *http.Request) {
	var a models.Network
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	vars := mux.Vars(r)
	db_handler := db.GetDB(r.Context())

	result := db_handler.InnerJoins("User", db_handler.Where(models.User{Email: jwt.GetEmailFromClaim(r.Context())})).Preload(clause.Associations).First(&a, vars["networkId"])
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusBadRequest)
		return
	}

	b, err := a.Json()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

func GetNetworks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	db_handler := db.GetDB(r.Context())

	var networks []models.Network
	result := db_handler.InnerJoins("User", db_handler.Where(models.User{Email: jwt.GetEmailFromClaim(r.Context())})).Find(&networks)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusBadRequest)
		return
	}

	j, err := json.Marshal(networks)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(j)
}

func UpdateNetwork(w http.ResponseWriter, r *http.Request) {
	AddNetwork(w, r)
}

func UpdateNode(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	var n models.NetworkNode

	err := json.NewDecoder(r.Body).Decode(&n)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err = n.Valid(); err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	db_handler := db.GetDB(r.Context())

	// Check if request has access to network
	err = db.RequestHasPermissionsToNetwork(db_handler, r, n.NetworkID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	result := db_handler.Save(&n)

	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusBadRequest)
		return
	}

	JsonResponse(&n, http.StatusCreated, w)
}

func UpdateNodeSelf(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	var n common.NodeUpdateRequest

	err := json.NewDecoder(r.Body).Decode(&n)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err = n.Node.Valid(); err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	db_handler := db.GetDB(r.Context())

	// Check if it was signed correctly

	result := db_handler.Find(&n.Node)

	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusBadRequest)
		return
	}

	result = db_handler.Save(&n.Node)

	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusBadRequest)
		return
	}
	JsonResponse(&n.Node, http.StatusCreated, w)
}

func GetNodes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	db_handler := db.GetDB(r.Context())

	networks, err := db.FetchAllNetworksThatRequestHasAccess(db_handler, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	nodes := []models.NetworkNode{}

	for _, network := range networks {
		nodes = append(nodes, network.Nodes...)
	}

	j, err := json.Marshal(nodes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(j)
}

func DeleteNode(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	vars := mux.Vars(r)
	db_handler := db.GetDB(r.Context())

	result := db_handler.InnerJoins("User", db_handler.Where(models.User{Email: jwt.GetEmailFromClaim(r.Context())})).Delete(&models.NetworkNode{}, vars["nodeId"])
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
