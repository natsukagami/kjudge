package worker

import (
	"time"
)

// Sandbox provides a way to run an arbitary command
// within a sandbox, with configured input/outputs and
// proper time and memory limits.
//
// kjudge currently implements two sandboxes, "isolate" (which requires "github.com/ioi/isolate" to be available)
// and "raw" (NOT RECOMMENDED, RUN AT YOUR OWN RISK).
// Which sandbox is used can be set at runtime with a command-line switch.
type Sandbox interface {
	Run(*SandboxInput) (*SandboxOutput, error)
}

// SandboxInput is the input to a sandbox.
type SandboxInput struct {
	Command     string            `json:"command"`      // The passed command
	Args        []string          `json:"args"`         // any additional arguments, if needed
	Files       map[string][]byte `json:"files"`        // Any additional files needed
	TimeLimit   time.Duration     `json:"time_limit"`   // The given time-limit
	MemoryLimit int               `json:"memory_limit"` // in KBs

	CompiledSource []byte `json:"compiled_source"` // Should be written down to the CWD as a file named "code", as the command expects
	Input          []byte `json:"input"`
}

// SandboxOutput is the output which the sandbox needs to give back.
type SandboxOutput struct {
	Success     bool          `json:"success"`      // Whether the command exited zero.
	RunningTime time.Duration `json:"running_time"` // The running time of the command.
	MemoryUsed  int           `json:"memory_used"`  // in KBs

	Stdout       []byte `json:"stdout"`
	Stderr       []byte `json:"stderr"`
	ErrorMessage string `json:"error_message,omitempty"`
}
