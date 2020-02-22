package models

import (
	"git.nkagami.me/natsukagami/kjudge/models/verify"
)

// Verify verifies a file's content.
func (f *File) Verify() error {
	return verify.All(map[string]error{
		"Filename": verify.Names(f.Filename),
		"Content":  verify.NotNull(f.Content),
	})
}
