package models

import (
	"git.nkagami.me/natsukagami/kjudge/models/verify"
)

// Verify verifies an User's contents.
func (r *User) Verify() error {
	return verify.All(map[string]error{
		"ID":           verify.Names(r.ID),
		"DisplayName":  verify.Names(r.DisplayName),
		"Organization": verify.StringEmptyOr(verify.StringMaxLength(64))(r.Organization),
	})
}
