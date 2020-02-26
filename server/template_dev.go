// +build !production

package server

import (
	"io"
	"log"

	"github.com/pkg/errors"
)

func init() {
    log.Println("Development environment detected. Templates will be re-parsed on every render")
}

// Render renders a template available in the compiled binary.
func Render(w io.Writer, name string, root interface{}) error {
	t, err := parseAllTemplates()
	if err != nil {
		return err
	}
	return errors.WithStack(t.ExecuteTemplate(w, name, root))
}
