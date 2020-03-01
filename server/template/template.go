package template

import (
	"html/template"
	"io"
	"log"
	"strings"

	"git.nkagami.me/natsukagami/kjudge/static"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

// List of all template's requirements.
// All requirements are then prepended (recursively) into the requirement list.
//
// The root template "root" is always prepended at the beginning.
var templateList = map[string][]string{
	"admin/home":     []string{"admin/root", "admin/contest_inputs"},
	"admin/contests": []string{"admin/root", "admin/contest_inputs"},
	"admin/contest":  []string{"admin/root", "admin/contest_inputs", "admin/problem_inputs"},
	"admin/problem":  []string{"admin/root", "admin/problem_inputs", "admin/test_inputs"},
}

// From a single template name, resolve the requirement tree into a list of template names.
func resolveTemplate(name string, into []string) []string {
	reqs, ok := templateList[name]
	// We're at a non-leaf template.
	if ok {
		// resolve the inner requirements one by one
		for _, req := range reqs {
			into = resolveTemplate(req, into)
		}
	}
	return append(into, name)
}

// Renderer implements echo.Renderer
type Renderer struct{}

var _ echo.Renderer = Renderer{}

// Render implement echo.Renderer.Render
func (r Renderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return Render(w, name, data)
}

func templateFilename(name string) string {
	return "templates/" + name + ".html"
}

func parseTemplateTree(root *template.Template, name string) (*template.Template, error) {
	names := resolveTemplate(name, nil)
	t, err := root.Clone()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	for _, name := range names {
		content, err := static.ReadFile(templateFilename(name))
		if err != nil {
			return nil, errors.Wrapf(err, "file %s", name)
		}
		if _, err := t.New(name).Parse(string(content)); err != nil {
			return nil, errors.Wrapf(err, "file %s", name)
		}
	}
	return t, nil
}

func parseRootTemplate() (*template.Template, error) {
	root, err := static.ReadFile("templates/root.html")
	if err != nil {
		return nil, errors.WithStack(err)
	}
	tRoot, err := template.New("").Parse(string(root))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return tRoot, nil
}

// Searches for and load all html templates.
func parseAllTemplates() (map[string]*template.Template, error) {
	tRoot, err := parseRootTemplate()
	if err != nil {
		return nil, err
	}
	mp := make(map[string]*template.Template)
	names := []string{}
	for file := range templateList {
		names = append(names, file)
		t, err := parseTemplateTree(tRoot, file)
		if err != nil {
			return nil, errors.Wrapf(err, "file %s", file)
		}
		mp[file] = t
	}
	log.Printf("defined templates: %s", strings.Join(names, ", "))
	return mp, nil
}
