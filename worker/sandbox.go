package worker

import (
	"github.com/natsukagami/kjudge/worker/sandbox"
	"github.com/natsukagami/kjudge/worker/sandbox/isolate"
	"github.com/natsukagami/kjudge/worker/sandbox/raw"
	"github.com/pkg/errors"
)

func NewSandbox(name string, options ...sandbox.Option) (sandbox.Runner, error) {
	setting := sandbox.MakeSettings(options...)
	switch name {
	case "raw":
		return raw.New(setting), nil
	case "isolate_v1":
		return isolate.New(1, setting), nil
	case "isolate":
		fallthrough
	case "isolate_v2":
		return isolate.New(2, setting), nil
	default:
		return nil, errors.Errorf("Sandbox %s doesn't exists or not yet implemented.", name)
	}
}
