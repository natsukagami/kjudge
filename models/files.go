package models

import (
	"bytes"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/natsukagami/kjudge/db"
	"github.com/natsukagami/kjudge/models/verify"
	"github.com/pkg/errors"
)

func crlftoLF(content []byte) ([]byte, error) {
	return bytes.ReplaceAll(content, []byte("\r\n"), []byte("\n")), nil
}

func lftoCRLF(content []byte) ([]byte, error) {
	lf := bytes.Count(content, []byte("\n"))
	crlf := bytes.Count(content, []byte("\r\n"))
	if crlf == 0 {
		return bytes.ReplaceAll(content, []byte("\r\n"), []byte("\n")), nil
	}
	if crlf == lf {
		return content, nil
	}
	return nil, errors.Errorf("number of crlf and lf (%v, %v) does not match", crlf, lf)
}

// NormalizeEndings normalize file line endings to the target OS's endings
// target accepts "windows" or "linux"
func NormalizeEndings(content []byte, target string) ([]byte, error) {
	switch (target){
	case "windows":
		return lftoCRLF(content)
	case "linux":
		return crlftoLF(content)
	default:
		return nil, errors.Errorf("%s not supported for line ending conversion", runtime.GOOS)
	}
}

// NormalizeEndings normalize file line endings to the current OS's endings
// target accepts "windows" or "linux"
func NormalizeEndingsNative(content []byte) ([]byte, error) {
	return NormalizeEndings(content, runtime.GOOS)
}

// IsTextFile applies heuristics to determine
// whether specified filename is a text file
func IsTextFile(filename string) bool {
	ext := filepath.Ext(filename)
	return !(ext == "" || ext == "exe" || ext == "pdf")
}

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
