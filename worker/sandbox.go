package worker

import (
	"github.com/natsukagami/kjudge/worker/sandbox"
	"github.com/natsukagami/kjudge/worker/sandbox/isolate"
	"github.com/natsukagami/kjudge/worker/sandbox/raw"
	"github.com/pkg/errors"
)

func NewSandbox(name string, options ...sandbox.Option) (sandbox.Runner, error) {
	setting := sandbox.DefaultSettings
	for _, option := range options {
		setting = option(setting)
	}
	switch name {
	case "raw":
		return raw.New(setting), nil
	case "isolate":
		return isolate.New(setting), nil
	default:
		return nil, errors.Errorf("Sandbox %s doesn't exists or not yet implemented.", name)
	}
}
