package compile

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/natsukagami/kjudge/models"
	"github.com/pkg/errors"
)

// Compile performs compilation.
// Returns whether the compilation succeeds.
func Compile(c *CompileContext) (bool, error) {
	action, err := getAction(c)
	if err != nil {
		return false, errors.WithStack(err)
	}

	log.Printf("[WORKER] Compiling submission %v\n", c.Sub.ID)

	// Now, create a temporary directory.
	dir, err := os.MkdirTemp("", "*")
	if err != nil {
		return false, errors.WithStack(err)
	}
	defer action.Cleanup(dir)

	// Prepare source and files
	action.Source.Content = c.Sub.Source
	action.Files = c.Files
	if err := action.Prepare(dir); err != nil {
		return false, err
	}

	// Perform compilation
	result, messages := action.Perform(dir)
	c.Sub.CompilerOutput = messages

	if result {
		// Success!
		output, err := os.ReadFile(filepath.Join(dir, action.Output))
		if err != nil {
			return false, errors.WithStack(err)
		}
		c.Sub.CompiledSource = output
	} else {
		c.Sub.CompiledSource = nil
		c.Sub.Verdict = models.VerdictCompileError
	}
	log.Printf("[WORKER] Compiling submission %v succeeded (result = %v).", c.Sub.ID, result)

	return result, c.Sub.Write(c.DB)
}

// Try creating a compile action using available schemes
func getAction(c *CompileContext) (*CompileAction, error) {
	if action, err := CompileBatch(c); action != nil || err != nil {
		return action, err
	}
	if action, err := CompileSingle(c); action != nil || err != nil {
		return action, err
	}
	return nil, errors.Errorf("no scheme accepted context")
}

var recognizedFilenames = []string{
	"statements.pdf", "statements.md", "compare", ".stages",
}

func isRecognizedFile(filename string) bool {
	for _, f := range recognizedFilenames {
		if f == filename {
			return true
		}
	}
	return false
}

// CompileAction represents the following steps:
// 1. Write the source into a file in "Source".
// 2. Copy all files in Files into "Source"
// 3. Compile the source with "Command".
// 4. Produce "Output" as the result.
type CompileAction struct {
	Source   *models.File
	Files    []*models.File
	Commands [][]string
	Output   string
}

// Prepare prepares a temporary folder and copies all the content there.
func (c *CompileAction) Prepare(dir string) error {
	// Copy over all files and the source code.
	if err := os.WriteFile(filepath.Join(dir, c.Source.Filename), c.Source.Content, 0666); err != nil {
		return errors.WithStack(err)
	}
	for _, file := range c.Files {
		if isRecognizedFile(file.Filename) {
			continue
		}
		if err := os.WriteFile(filepath.Join(dir, file.Filename), file.Content, 0666); err != nil {
			return errors.Wrapf(err, "copying file %s", file.Filename)
		}
	}
	return nil
}

// Cleanup performs clean-up on the prepared directory.
func (c *CompileAction) Cleanup(dir string) {
	_ = os.RemoveAll(dir)
}

// Perform performs the compile action on the given directory.
// The directory MUST contain all files given by the Problem, PLUS the written "Source" file.
func (c *CompileAction) Perform(cwd string) (succeeded bool, messages []byte) {
	allOutputs := bytes.Buffer{}
	for _, command := range c.Commands {
		allOutputs.WriteString(fmt.Sprintf("%s:\n", strings.Join(command, " ")))
		cmd := exec.Command(command[0], command[1:]...)
		// Set the cwd
		cmd.Dir = cwd
		// Run the command and collect outputs
		var (
			output []byte
			err    error
			done   = make(chan struct{}, 1)
		)
		go func() {
			output, err = cmd.CombinedOutput()
			done <- struct{}{}
		}()
		select {
		case <-time.After(20 * time.Second):
			allOutputs.WriteString("Command has timed out\n")
			return false, allOutputs.Bytes()
		case <-done:
		}
		allOutputs.Write(output)
		allOutputs.WriteString("\n")
		if err != nil {
			log.Println(err)
			return false, allOutputs.Bytes()
		}
	}
	return true, allOutputs.Bytes()
}
