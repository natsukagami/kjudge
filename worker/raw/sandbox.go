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

	"git.nkagami.me/natsukagami/kjudge/worker"
)

// Sandbox implements worker.Sandbox.
type Sandbox struct{}

var _ worker.Sandbox = (*Sandbox)(nil)

// Run implements Sandbox.Run
func (s *Sandbox) Run(input *worker.SandboxInput) (*worker.SandboxOutput, error) {
	dir := os.TempDir()

	log.Printf("[SANDBOX] Running %s %v\n", input.Command, input.Args)

	return s.RunFrom(dir, input)
}

// RunFrom runs the input, assuming that it has write access to "cwd".
//
// Raw sandbox assumes that:
//   - MEMORY LIMITS ARE NOT SET. It always reports a memory usage of 0 (it cannot measure them).
//   - THE PROGRAM DOES NOT MESS WITH THE COMPUTER. LMAO
//   - The folder will be thrown away later.
func (s *Sandbox) RunFrom(cwd string, input *worker.SandboxInput) (*worker.SandboxOutput, error) {
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
		return &worker.SandboxOutput{
			Success:      false,
			MemoryUsed:   0,
			RunningTime:  input.TimeLimit,
			Stdout:       []byte{},
			Stderr:       []byte{},
			ErrorMessage: "Command timed out",
		}, nil
	case commandErr := <-done:
		runningTime := time.Now().Sub(startTime)
		return &worker.SandboxOutput{
			Success:      commandErr == nil,
			MemoryUsed:   0,
			RunningTime:  runningTime,
			Stdout:       stdout.Bytes(),
			Stderr:       stderr.Bytes(),
			ErrorMessage: "",
		}, nil

	}

}
