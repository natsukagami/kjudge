// +build production

package template

import (
	"fmt"
	"html/template"
	"io"
	"log"

	"git.nkagami.me/natsukagami/kjudge"
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

func version() string { return fmt.Sprintf("%s \"%s\"", kjudge.Version, kjudge.Codename) }

// Render renders a template available in the compiled binary.
func Render(w io.Writer, name string, root interface{}) error {
	return errors.WithStack(rootTemplate[name].Execute(w, root))
}
