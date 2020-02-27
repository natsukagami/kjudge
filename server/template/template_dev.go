// +build !production

package template

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
	tRoot, err := parseRootTemplate()
	if err != nil {
		return err
	}
	t, err := parseTemplateTree(tRoot, name)
	if err != nil {
		return err
	}
	return errors.WithStack(t.Execute(w, root))
}
