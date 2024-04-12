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

var challenges = map[string]string{}

func GetConnectionConfigurationChallenge(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	var request common.ConnectionConfigurationChallengeRequest

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Fetch node from DB
	db_handler := db.GetDB(r.Context())

	networkNode := models.NetworkNode{}
	result := db_handler.Preload(clause.Associations).Where("peer_id = ?", request.PeerID).First(&networkNode)
	if result.Error != nil || result.RowsAffected < 1 {
		http.Error(w, result.Error.Error(), http.StatusNotFound)
		return
	}

	challenges[networkNode.PeerID], err = models.GenerateNewRandomCode(128)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := common.ConnectionConfigurationChallengeResponse{Challenge: challenges[networkNode.PeerID]}

	JsonResponse(&response, http.StatusOK, w)
}

func GetConnectionConfiguration(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	var request common.ConnectionConfigurationRequest

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	challenge, found := challenges[request.PeerID]
	if !found {
		http.Error(w, "Challenge not found", http.StatusBadRequest)
		return
	}
	delete(challenges, request.PeerID)

	db_handler := db.GetDB(r.Context())

	networkNode := models.NetworkNode{}
	result := db_handler.Preload(clause.Associations).Where("peer_id = ?", request.PeerID).First(&networkNode)
	if result.Error != nil || result.RowsAffected < 1 {
		http.Error(w, result.Error.Error(), http.StatusNotFound)
		return
	}

	rawSignedChallenge, err := base64.RawStdEncoding.DecodeString(request.SignedChallenge)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	pubKey := ed25519.PublicKey(networkNode.PublicKey)
	err = ed25519.VerifyWithOptions(pubKey, []byte(challenge), rawSignedChallenge, common.NodeAuthenticationTokenOptions)
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

func GetNextFreeIPAddress(w http.ResponseWriter, r *http.Request) {
	var network models.Network
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	vars := mux.Vars(r)
	db_handler := db.GetDB(r.Context())

	result := db_handler.InnerJoins("User", db_handler.Where(models.User{Email: jwt.GetEmailFromClaim(r.Context())})).Preload(clause.Associations).First(&network, vars["networkId"])
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusBadRequest)
		return
	}

	nextIp := struct {
		NextIP  string
		Network string
	}{network.NextFreeIP(), network.CIDR}

	JsonResponse(&nextIp, http.StatusOK, w)
}
