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

// NormalizeEndingsUnix normalize file line endings to LF
func NormalizeEndingsUnix(content []byte) ([]byte, error) {
	return bytes.ReplaceAll(content, []byte("\r\n"), []byte("\n")), nil
}

// NormalizeEndingsWindows normalize file line endings to CRLF
// and throws if there is LF and CRLF mixed together
func NormalizeEndingsWindows(content []byte) ([]byte, error) {
	lf := bytes.Count(content, []byte("\n"))
	crlf := bytes.Count(content, []byte("\r\n"))
	if crlf == lf {
		return content, nil
	}
	var err error = nil
	if crlf != 0 {
		err = errors.Errorf("number of crlf and lf (%v, %v) does not match", crlf, lf)
	}
	return bytes.ReplaceAll(content, []byte("\r\n"), []byte("\n")), err
}

// NormalizeEndings normalize file line endings to the target OS's endings
// target accepts "windows" or "linux". Returns error if OS is not supported
// or there is LF and CRLF mixed together
func NormalizeEndings(content []byte, target string) ([]byte, error) {
	switch target {
	case "windows":
		return NormalizeEndingsWindows(content)
	case "linux":
		return NormalizeEndingsUnix(content)
	default:
		return nil, errors.Errorf("%s not supported for line ending conversion", runtime.GOOS)
	}
}

// IsTextFile applies heuristics to determine
// whether specified filename is a text file
func IsTextFile(filename string) bool {
	ext := filepath.Ext(filename)
	return !(ext == "" || ext == "exe" || ext == "pdf" || ext == "zip")
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
