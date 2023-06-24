// Raw implements a "raw" sandbox.
//
// The sandbox does NOT prevent against ANY malicious attacks,
// not even a single shell command. Therefore this is NOT RECOMMENDED
// to run an online judge with this box UNLESS you ABSOLUTELY TRUST all
// the users.
//
// That said, it is useful if somehow you want to use this alone when you
// don't have access to a sandbox like isolate.
package raw

import (
	"bytes"
	"context"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/natsukagami/kjudge/worker/sandbox"
)

// Runner implements worker.Runner.
type Runner struct {
	settings sandbox.Settings
}

var _ sandbox.Runner = (*Runner)(nil)

func New(settings sandbox.Settings) *Runner {
	if !settings.IgnoreWarning {
		log.Println("'raw' sandbox selected. WE ARE NOT RESPONSIBLE FOR ANY BREAKAGE CAUSED BY FOREIGN CODE.")
	}
	return &Runner{settings: settings}
}

func (s *Runner) Settings() *sandbox.Settings {
	return &s.settings
}

// Run implements Runner.Run
func (s *Runner) Run(input *sandbox.Input) (*sandbox.Output, error) {
	dir := os.TempDir()

	if s.Settings().LogSandbox {
		log.Printf("[SANDBOX] Running %s %v\n", input.Command, input.Args)
	}

	return s.RunFrom(dir, input)
}

// RunFrom runs the input, assuming that it has write access to "cwd".
//
// Raw sandbox assumes that:
//   - MEMORY LIMITS ARE NOT SET. It always reports a memory usage of 0 (it cannot measure them).
//   - THE PROGRAM DOES NOT MESS WITH THE COMPUTER. LMAO
//   - The folder will be thrown away later.
func (s *Runner) RunFrom(cwd string, input *sandbox.Input) (*sandbox.Output, error) {
	if err := input.CopyTo(cwd); err != nil {
		return nil, err
	}

	// Prepare the command
	if !strings.HasPrefix(input.Command, "/") {
		input.Command = "./" + input.Command
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cmd := exec.CommandContext(ctx, input.Command, input.Args...)
	cmd.Dir = cwd
	cmd.Env = []string{"ONLINE_JUDGE=true", "KJUDGE=true"} // No env access
	cmd.Stdin = bytes.NewBuffer(input.Input)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Collect output BUT don't do it for too long
	done := make(chan error)
	var startTime time.Time
	go func() {
		startTime = time.Now()
		done <- cmd.Run()
	}()

	select {
	case <-time.After(input.TimeLimit):
		cancel()
		<-done
		return &sandbox.Output{
			Success:      false,
			MemoryUsed:   0,
			RunningTime:  input.TimeLimit,
			Stdout:       []byte{},
			Stderr:       []byte{},
			ErrorMessage: "Command timed out",
		}, nil
	case commandErr := <-done:
		runningTime := time.Since(startTime)
		return &sandbox.Output{
			Success:      commandErr == nil,
			MemoryUsed:   0,
			RunningTime:  runningTime,
			Stdout:       stdout.Bytes(),
			Stderr:       stderr.Bytes(),
			ErrorMessage: "",
		}, nil

	}
}
