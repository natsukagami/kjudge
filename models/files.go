package models

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/natsukagami/kjudge/db"
	"github.com/natsukagami/kjudge/models/verify"
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

// GetProblemFilesMeta works like GetProblemFiles, but ignores all contents.
func GetProblemFilesMeta(db db.DBContext, problemID int) ([]*File, error) {
	var f []*File
	if err := db.Select(&f, "SELECT filename, id, problem_id, public FROM files WHERE problem_id = ?", problemID); err != nil {
		return nil, errors.WithStack(err)
	}
	return f, nil
}

// Verify verifies a file's content.
func (f *File) Verify() error {
	return verify.All(map[string]error{
		"Filename": verify.Names(f.Filename),
		"Content":  verify.NotNull(f.Content),
	})
}

// Compilable returns whether a file can be compiled.
func (f *File) Compilable() bool {
	_, err := LanguageByExt(filepath.Ext(f.Filename))
	return err == nil
}

// WriteFiles writes the given files as brand new, overwritting the old ones.
// Note that because of overwritting behaviour, we cannot ensure the validity of the indicies, hence they are not reflected into
// the *Files.
func (p *Problem) WriteFiles(db db.DBContext, files []*File) error {
	for _, f := range files {
		f.ProblemID = p.ID
		if err := f.Verify(); err != nil {
			return errors.Wrapf(err, "file %s", f.Filename)
		}
	}
	var (
		clauses []string
		params  []interface{}
	)
	for _, f := range files {
		clauses = append(clauses, "(?, ?, ?, ?)")
		params = append(params, f.ProblemID, f.Public, f.Content, f.Filename)
	}
	if _, err := db.Exec(
		fmt.Sprintf(`INSERT INTO files(problem_id, public, content, filename) VALUES %s 
		            ON CONFLICT (problem_id, filename) DO UPDATE SET public = excluded.public, content = excluded.content`, strings.Join(clauses, ", ")),
		params...); err != nil {
		return errors.WithStack(err)
	}
	return nil
}
