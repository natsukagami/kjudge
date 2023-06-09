// Package isolate provides a safe sandbox on Linux using the
// "isolate" program.
//
// The use of this sandbox requires "isolate" (https://github.com/ioi/isolate)
// to be installed and callable from $PATH.
package isolate

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/natsukagami/kjudge/worker/sandbox"
	"github.com/pkg/errors"
)

var (
	// The isolate command. Can be overridden with KJUDGE_ISOLATE environment variable.
	isolateCommand = "isolate"
)

func init() {
	if v, ok := os.LookupEnv("KJUDGE_ISOLATE"); ok {
		isolateCommand = v
	}
}

// Runner implements worker.Runner.
type Runner struct {
	settings sandbox.Settings
	private  struct{} // Makes the sandbox not simply constructible
}

var _ sandbox.Runner = (*Runner)(nil)

// Panics on not having "isolate" accessible.
func mustHaveIsolate() {
	output, err := exec.Command(isolateCommand, "--version").CombinedOutput()
	if err != nil {
		panic(errors.Wrap(err, "trying to run isolate"))
	}
	if !strings.Contains(string(output), "The process isolator") {
		panic("Wrong isolate command found. Override the KJUDGE_ISOLATE environment variable to set a different path.")
	}
}

// New returns a new sandbox.
// Panics if isolate is not installed.
func New(settings sandbox.Settings) *Runner {
	mustHaveIsolate()
	return &Runner{settings: settings, private: struct{}{}}
}

func (s *Runner) Settings() *sandbox.Settings {
	return &s.settings
}

// Run implements Runner.Run.
func (s *Runner) Run(input *sandbox.Input) (*sandbox.Output, error) {
	// Init the sandbox
	defer s.cleanup()
	dirBytes, err := exec.Command(isolateCommand, "--init", "--cg").Output()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	dir := filepath.Join(strings.TrimSpace(string(dirBytes)), "box")

	// Copy items to dir
	if err := input.CopyTo(dir); err != nil {
		return nil, err
	}

	// Prepare a meta file.
	tmp := os.TempDir()

	metaFile := filepath.Join(tmp, "meta.txt")
	cmd := buildCmd(dir, metaFile, input)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if e, ok := err.(*exec.ExitError); !ok || e.ExitCode() != 1 {
			return nil, errors.WithStack(err)
		}
	}

	// Parse the meta file
	output := &sandbox.Output{
		Stdout: stdout.Bytes(),
		Stderr: stderr.Bytes(),
	}
	if err := parseMetaFile(metaFile, output); err != nil {
		return nil, err
	}

	return output, nil
}

func parseMetaFile(path string, output *sandbox.Output) error {
	meta, err := ReadMetaFile(path)
	if err != nil {
		return err
	}
	if _, ok := meta.Fields["message"]; ok {
		output.ErrorMessage = meta.String("message")
		output.Success = false
	} else {
		output.Success = true
	}
	output.MemoryUsed = meta.Int("cg-mem")
	output.RunningTime = time.Duration(float64(time.Second) * meta.Float64("time"))
	return meta.Error()
}

// Build the command for isolate --run.
func buildCmd(dir, metaFile string, input *sandbox.Input) *exec.Cmd {
	// Calculate stuff
	timeLimit := float64(input.TimeLimit) / float64(time.Second)

	// Run the main isolate command
	cmd := exec.Command(
		"isolate",
		"--cg",
		"--run",
		"-M", metaFile,
		"-t", fmt.Sprintf("%.1f", timeLimit), // Time limit
		"-w", fmt.Sprintf("%.1f", 2*timeLimit+1.0), // Wall time
		"-x", "1.0", // Extra time
		"-f", "262144", // 256MBs of files
		"-p", // Allow multiple processes
		"-s", // Be silent
		"--env=ONLINE_JUDGE=true",
		"--env=KJUDGE=true",
		fmt.Sprintf("--cg-mem=%d", input.MemoryLimit), // total memory
		"--",
		input.Command,
	)
	if len(input.Args) > 0 {
		cmd.Args = append(cmd.Args, input.Args...)
	}
	cmd.Dir = dir

	// Pipe the stdin
	cmd.Stdin = bytes.NewBuffer(input.Input)

	return cmd
}

func (s *Runner) cleanup() {
	_ = exec.Command(isolateCommand, "--cleanup", "--cg").Run()
}
