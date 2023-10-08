/*
 *
 * solo Server API
 *
 */
package api

import (
	"crypto/ed25519"
	"encoding/base64"
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

	result := db_handler.Delete(&models.Network{}, vars["networkId"])
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

	result := db_handler.Preload("User").First(&a, vars["networkId"])
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

	result := db_handler.Save(&n)

	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusBadRequest)
		return
	}

	JsonResponse(&n, http.StatusCreated, w)
}

func NetworkAssignNodeFromRegistrationCode(w http.ResponseWriter, r *http.Request) {
	var network models.Network
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	vars := mux.Vars(r)
	db_handler := db.GetDB(r.Context())

	result := db_handler.Preload("User").First(&network, vars["networkId"])
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
}

func GetConnectionConfiguration(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	var request common.ConnectionConfigurationRequest

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	db_handler := db.GetDB(r.Context())

	networkNode := models.NetworkNode{}
	result := db_handler.Preload(clause.Associations).Where("peer_id = ?", request.PeerID).First(&networkNode)
	if result.Error != nil || result.RowsAffected < 1 {
		http.Error(w, result.Error.Error(), http.StatusNotFound)
		return
	}

	rawAuthenticationToken, err := base64.RawStdEncoding.DecodeString(request.NodeAuthenticationToken)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	pubKey := ed25519.PublicKey(networkNode.PublicKey)
	err = ed25519.VerifyWithOptions(pubKey, common.NodeAuthenticationTokenMessage(networkNode.Hostname), rawAuthenticationToken, common.NodeAuthenticationTokenOptions)
	if err != nil {
		http.Error(w, "Node authentication token is invalid", http.StatusBadRequest)
		return
	}

	if !networkNode.Actived {
		http.Error(w, "Node wasn't activated yet", http.StatusFailedDependency)
		return
	}

	response := common.ConnectionConfigurationResponse{
		ConnectionConfigToken: networkNode.Network.ConnectionConfigToken,
		InterfaceAddress:      networkNode.IP,
	}

	JsonResponse(&response, http.StatusOK, w)
}
