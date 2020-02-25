package models

import (
	"git.nkagami.me/natsukagami/kjudge/db"
	"git.nkagami.me/natsukagami/kjudge/models/verify"
	"github.com/pkg/errors"
)

// GetFileWithName returns a file with a given name.
func GetFileWithName(db db.DBContext, problemID int, filename string) (*File, error) {
	var f File
	if err := db.Get(&f, "SELECT * FROM files WHERE problem_id = ? AND filename = ?", problemID, filename); err != nil {
		return nil, errors.WithStack(err)
	}
	return &f, nil
}

// Verify verifies a file's content.
func (f *File) Verify() error {
	return verify.All(map[string]error{
		"Filename": verify.Names(f.Filename),
		"Content":  verify.NotNull(f.Content),
	})
}
