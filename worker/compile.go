package worker

// Compiling anything that's more compicated than single file:
//
// - Prepare a "compile_%s.%ext" file, with %s being the language (cc, go, rs, java, py2, py3, pas)
// - Prepare any more files as needed. They will all be put into the CWD of the script
// - The CWD also contains "code.%s" (%s being the language's respective extension) file, which is the contestant's source code.
// - The script should do whatever it wants (unsandboxed, because it's not my job to do so) within 20 seconds.
// - It should produce a single binary called "code" in the CWD.

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/natsukagami/kjudge/models"
	"github.com/pkg/errors"
)

// CompileContext is the information needed to perform compilation.
type CompileContext struct {
	DB        *sqlx.Tx
	Sub       *models.Submission
	Problem   *models.Problem
	AllowLogs bool
}

func (c *CompileContext) Log(format string, v ...interface{}) {
	if !c.AllowLogs {
		return
	}
	log.Printf(format, v...)
}

// Compile performs compilation.
// Returns whether the compilation succeeds.
func Compile(c *CompileContext) (bool, error) {
	// First we gotta know which compilation scheme we will be taking.
	files, err := models.GetProblemFiles(c.DB, c.Problem.ID)
	if err != nil {
		return false, err
	}
	action, batchFile, err := CompileBatch(c.Sub.Language)
	if err != nil {
		return false, err
	}
	hasFile := false
	hasBatch := false
	for _, file := range files {
		hasBatch = hasBatch || isBatchFile(file.Filename)
		if file.Filename == batchFile {
			hasFile = true
			break
		}
	}
	if !hasBatch {
		// No batch file, compiling as a single file.
		action, err = CompileSingle(c.Sub.Language)
		if err != nil {
			return false, err
		}
	} else if !hasFile {
		// Batch compile mode enabled, but this language is not supported.
		c.Sub.CompiledSource = nil
		c.Sub.Verdict = models.VerdictCompileError
		c.Sub.CompilerOutput = []byte("Custom Compilers are not enabled for this language.")
		return false, c.Sub.Write(c.DB)
	}

	c.Log("[WORKER] Compiling submission %v\n", c.Sub.ID)

	// Now, create a temporary directory.
	dir, err := os.MkdirTemp("", "*")
	if err != nil {
		return false, errors.WithStack(err)
	}
	defer action.Cleanup(dir)

	// Prepare source and files
	action.Source.Content = c.Sub.Source
	action.Files = files
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

	c.Log("[WORKER] Compiling submission %v succeeded (result = %v).", c.Sub.ID, result)

	return result, c.Sub.Write(c.DB)
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

var batchFilenames = []string{
	"compile_cc.sh", "compile_go.sh", "compile_java.sh", "compile_rs.sh", "compile_pas.sh", "compile_py2.sh", "compile_py3.sh",
}

func isBatchFile(filename string) bool {
	for _, f := range batchFilenames {
		if f == filename {
			return true
		}
	}
	return false
}

// CompileBatch returns a compile action, along with the required batch filename
// in order to successfully compile.
func CompileBatch(l models.Language) (*CompileAction, string, error) {
	var source, batch string
	switch l {
	case models.LanguageCpp:
		source = "code.cc"
		batch = "compile_cc.sh"
	case models.LanguageGo:
		source = "code.go"
		batch = "compile_go.sh"
	case models.LanguageJava:
		source = "code.java"
		batch = "compile_java.sh"
	case models.LanguageRust:
		source = "code.rs"
		batch = "compile_rs.sh"
	case models.LanguagePas:
		source = "code.pas"
		batch = "compile_pas.sh"
	case models.LanguagePy2:
		source = "code.py"
		batch = "compile_py2.sh"
	case models.LanguagePy3:
		source = "code.py"
		batch = "compile_py3.sh"
	default:
		return nil, "", errors.New("unknown language")
	}

	return &CompileAction{
		Source:   &models.File{Filename: source},
		Commands: [][]string{{"sh", batch}},
		Output:   "code",
	}, batch, nil
}

// CompileSingle creates a compilation command for a single source code file.
// Sometimes this is as simple as "copy".
func CompileSingle(l models.Language) (*CompileAction, error) {
	switch l {
	case models.LanguageCpp:
		return &CompileAction{
			Source:   &models.File{Filename: "code.cc"},
			Commands: [][]string{{"g++", "-std=c++17", "-O2", "-s", "-lm", "-DONLINE_JUDGE", "-DKJUDGE", "-o", "code", "code.cc"}},
			Output:   "code",
		}, nil
	case models.LanguageGo:
		return &CompileAction{
			Source:   &models.File{Filename: "code.go"},
			Commands: [][]string{{"go", "build", "-buildmode=exe", "-tags", "online_judge,kjudge", "-o", "code", "code.go"}},
			Output:   "code",
		}, nil
	case models.LanguageJava:
		return &CompileAction{
			Source: &models.File{Filename: "code.java"},
			Commands: [][]string{
				{"javac", "-d", ".", "code.java"},
				{"sh", "-c", "jar cf code *.class"},
				{"sh", "-c", "rm *.class"},
			},
			Output: "code",
		}, nil
	case models.LanguagePas:
		return &CompileAction{
			Source: &models.File{Filename: "code.pas"},
			Commands: [][]string{
				{"fpc", "-O3", "-dONLINE_JUDGE", "-dKJUDGE", "-ocode", "code.pas"},
			},
			Output: "code",
		}, nil
	case models.LanguageRust:
		return &CompileAction{
			Source: &models.File{Filename: "code.rs"},
			Commands: [][]string{
				{"rustc", "-O", "--cfg", "online_judge", "--cfg", "kjudge", "-o", "code", "code.rs"},
			},
			Output: "code",
		}, nil
	case models.LanguagePy2:
		return &CompileAction{
			Source: &models.File{Filename: "code.py"},
			Commands: [][]string{
				{"python2", "-m", "py_compile", "code.py"},
			},
			Output: "code.pyc",
		}, nil
	case models.LanguagePy3:
		return &CompileAction{
			Source:   &models.File{Filename: "code.py"},
			Commands: [][]string{{"python3", "-c", `import py_compile as m; m.compile("code.py", "code.pyc", doraise=True)`}},
			Output:   "code.pyc",
		}, nil
	default:
		return nil, errors.Errorf("Unknown language: %v", l)
	}
}
