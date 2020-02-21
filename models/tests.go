package models

import (
	"git.nkagami.me/natsukagami/kjudge/models/verify"
	"github.com/pkg/errors"
)

// Verify verifies Test's contents.
func (r *Test) Verify() error {
	if r.Input == nil {
		return errors.New("input must not be null")
	}
	if r.Output == nil {
		return errors.New("output must not be null")
	}
	return errors.Wrapf(verify.Names(r.Name), "field name")
}
