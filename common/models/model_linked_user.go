/*
 *
 * solo Server API
 *
 */
package models

type LinkedUser struct {
	Model

	NetworkID uint

	UserID uint
	User   *User `json:"user,omitempty"`
	// Linked user permissions
	Permissions string `json:"permissions,omitempty"`
}
