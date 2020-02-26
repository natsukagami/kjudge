// +build production

package server

import (
	"html/template"
	"io"

	"github.com/pkg/errors"
)

var rootTemplate *template.Template

func init() {
	rootTemplate = template.Must(parseAllTemplates())
}

// Render renders a template available in the compiled binary.
func Render(w io.Writer, name string, root interface{}) error {
	return errors.WithStack(rootTemplate.ExecuteTemplate(w, name, root))
}
