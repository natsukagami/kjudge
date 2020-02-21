package models

import (
	"git.nkagami.me/natsukagami/kjudge/models/verify"
	"github.com/pkg/errors"
)

// Verify verifies an User's contents.
func (r *User) Verify() error {
	return errors.Wrap(verify.Names(r.ID), "field id")
}
