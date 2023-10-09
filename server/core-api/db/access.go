package db

import (
	"net/http"

	"github.com/gfleury/solo/common/models"
	"github.com/gfleury/solo/server/core-api/jwt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func RequestHasPermissionsToNetwork(db_handler *gorm.DB, r *http.Request, networkID *uint) error {
	// Node is still waiting to be activated
	if networkID == nil {
		return nil
	}
	// Check if the User owns the node
	var network models.Network
	result := db_handler.InnerJoins("User", db_handler.Where(models.User{Email: jwt.GetEmailFromClaim(r.Context())})).Preload(clause.Associations).First(&network, networkID)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func FetchAllNetworksThatRequestHasAccess(db_handler *gorm.DB, r *http.Request) ([]models.Network, error) {
	var networks []models.Network
	result := db_handler.InnerJoins("User", db_handler.Where(models.User{Email: jwt.GetEmailFromClaim(r.Context())})).Preload(clause.Associations).Find(&networks)
	if result.Error != nil {
		return nil, result.Error
	}
	return networks, nil
}
