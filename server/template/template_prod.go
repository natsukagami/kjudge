// +build production

package template

import (
	"html/template"
	"io"
	"log"

	"github.com/pkg/errors"
)

var rootTemplate map[string]*template.Template

func init() {
	var err error
	rootTemplate, err = parseAllTemplates()
	if err != nil {
		log.Fatalf("%+v", err)
	}
}

// Render renders a template available in the compiled binary.
func Render(w io.Writer, name string, root interface{}) error {
	return errors.WithStack(rootTemplate[name].Execute(w, root))
}
