package server

import (
	"html/template"
	"strings"

	"git.nkagami.me/natsukagami/kjudge/static"
	"github.com/pkg/errors"
)

// Searches for and load all html templates.
func parseAllTemplates() (*template.Template, error) {
	t := template.New("_root")
	files, err := static.WalkDirs("templates", false)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	for _, file := range files {
		if strings.HasSuffix(file, ".html") {
			// Strip "templates/" and ".html"
			name := file[len("templates/") : len(file)-len(".html")]
			content, err := static.ReadFile(file)
			if err != nil {
				return nil, errors.Wrapf(err, "file %s", file)
			}
			if _, err := t.New(name).Parse(string(content)); err != nil {
				return nil, errors.Wrapf(err, "file %s", file)
			}
		}
	}
	return t, nil
}
