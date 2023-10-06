/*
 *
 * solo Server API
 *
 */
package models

type User struct {
	Model

	Username string `json:"username,omitempty" gorm:"unique"`

	Email string `json:"email,omitempty" gorm:"unique"`

	// User Status
	UserStatus int32 `json:"userStatus,omitempty"`
	// Accounts that belong to user
	Networks []Network `json:"networks,omitempty"`

	Verified bool `json:"verified,omitempty"`

	Avatar string `json:"avatar,omitempty"`
}
