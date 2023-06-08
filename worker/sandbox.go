package worker

import (
	"github.com/natsukagami/kjudge/worker/sandbox"
	"github.com/natsukagami/kjudge/worker/sandbox/isolate"
	"github.com/natsukagami/kjudge/worker/sandbox/raw"
	"github.com/pkg/errors"
)

func NewSandbox(name string) (sandbox.Runner, error) {
	switch name {
	case "raw":
		return &raw.Runner{}, nil
	case "isolate":
		return isolate.New(), nil
	default:
		return nil, errors.Errorf("Sandbox %s doesn't exists or not yet implemented.", name)
	}
}
