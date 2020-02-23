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
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"git.nkagami.me/natsukagami/kjudge/worker"
	"github.com/pkg/errors"
)

// Sandbox implements worker.Sandbox.
type Sandbox struct{}

var _ worker.Sandbox = (*Sandbox)(nil)

// Run implements Sandbox.Run
func (s *Sandbox) Run(input *worker.SandboxInput) (*worker.SandboxOutput, error) {
	dir := os.TempDir()
	defer os.RemoveAll(dir)

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
	// Copy all the files into "cwd"
	for name, file := range input.Files {
		if err := ioutil.WriteFile(filepath.Join(cwd, name), file, 0666); err != nil {
			return nil, errors.Wrapf(err, "writing file %s", name)
		}
	}
	// Copy and set chmod the "code" file
	if input.CompiledSource != nil {
		if err := ioutil.WriteFile(filepath.Join(cwd, "code"), input.CompiledSource, 0777); err != nil {
			return nil, errors.WithStack(err)
		}
	}

	// Prepare the command
	cmd := exec.Command(input.Command, input.Args...)
	cmd.Dir = cwd
	cmd.Env = []string{"ONLINE_JUDGE=true", "KJUDGE=true"} // No env access
	cmd.Stdin = bytes.NewBuffer(input.Input)

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	startTime := time.Now()
	if err := cmd.Start(); err != nil {
		return nil, errors.WithStack(err)
	}

	var stdout, stderr []byte

	// Collect output BUT don't do it for too long
	done := make(chan error)
	go func() {
		stdout, err = ioutil.ReadAll(stdoutPipe)
		if err != nil {
			done <- errors.WithStack(err)
			return
		}
		stderr, err = ioutil.ReadAll(stderrPipe)
		if err != nil {
			done <- errors.WithStack(err)
			return
		}
		done <- cmd.Wait()
	}()

	select {
	case <-time.After(input.TimeLimit):
		cmd.Process.Kill()
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
		if err != nil {
			return nil, err // ReadAll errors
		}
		return &worker.SandboxOutput{
			Success:      commandErr == nil,
			MemoryUsed:   0,
			RunningTime:  runningTime,
			Stdout:       stdout,
			Stderr:       stderr,
			ErrorMessage: "",
		}, nil

	}

}
