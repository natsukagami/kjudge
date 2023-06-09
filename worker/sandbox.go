package worker

import (
	"log"

	"github.com/natsukagami/kjudge/worker/sandbox"
	"github.com/natsukagami/kjudge/worker/sandbox/isolate"
	"github.com/natsukagami/kjudge/worker/sandbox/raw"
	"github.com/pkg/errors"
)

func NewSandbox(name string, ignoreWarnings ...bool) (sandbox.Runner, error) {
	if len(ignoreWarnings) > 1 {
		return nil, errors.New("Function NewSandbox only takes 1 or 2 arguments")
	}
	warningsIgnored := false
	if len(ignoreWarnings) == 1 && ignoreWarnings[0] {
		warningsIgnored = true
	}

	switch name {
	case "raw":
		if !warningsIgnored {
			log.Println("'raw' sandbox selected. WE ARE NOT RESPONSIBLE FOR ANY BREAKAGE CAUSED BY FOREIGN CODE.")
		}
		return &raw.Runner{}, nil
	case "isolate":
		return isolate.New(), nil
	default:
		return nil, errors.Errorf("Sandbox %s doesn't exists or not yet implemented.", name)
	}
}
