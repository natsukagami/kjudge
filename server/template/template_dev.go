// +build !production

package template

import (
	"fmt"
	"io"
	"log"
	"os/exec"
	"strings"

	"github.com/natsukagami/kjudge"
	"github.com/pkg/errors"
)

func init() {
	log.Println("Development environment detected. Templates will be re-parsed on every render")
}

var developmentVersion = "unknown " + kjudge.Version

func init() {
	output, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		log.Println("git not found or not inside a git repository. You are running an unknown development version.")
		log.Println("Please build and run the development directly from the official git repository https://github.com/natsukagami/kjudge")
	}
	developmentVersion = strings.TrimSpace(string(output))
}

func version() string { return fmt.Sprintf("dev [%s]", developmentVersion) }

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
