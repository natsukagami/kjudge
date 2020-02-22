package worker

// Compiling anything that's more compicated than single file:
//
// - Prepare a "compile_%s.sh" file, with %s being the language (cc, go, rs, java, py2, py3, pas).
// - Prepare any more files as needed. They will all be put into the CWD of the script.
// - The CWD also contains "code.%s" (%s being the language's respective extension) file, which is the contestant's source code.
// - The script should do whatever it wants (unsandboxed, because it's not my job to do so) within 20 seconds.
// - It should produce a single binary called "code" in the CWD.

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"git.nkagami.me/natsukagami/kjudge/models"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

// CompileContext is the information needed to perform compilation.
type CompileContext struct {
	DB      *sqlx.Tx
	Sub     *models.Submission
	Problem *models.Problem
}

// Compile performs compilation.
// Returns whether the compilation succeeds.
func (c *CompileContext) Compile() (bool, error) {
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
	for _, file := range files {
		if file.Filename == batchFile {
			hasFile = true
			break
		}
	}
	if !hasFile {
		// No batch file, compiling as a single file.
		action, err = CompileSingle(c.Sub.Language)
		if err != nil {
			return false, err
		}
	}

	log.Printf("[WORKER] Compiling submission %v\n", c.Sub.ID)

	// Now, create a temporary directory.
	dir := os.TempDir()
	defer os.RemoveAll(dir)
	// Copy over all files and the source code.
	if err := ioutil.WriteFile(filepath.Join(dir, action.Source), c.Sub.Source, 0666); err != nil {
		return false, errors.WithStack(err)
	}
	for _, file := range files {
		if err := ioutil.WriteFile(filepath.Join(dir, file.Filename), file.Content, 0666); err != nil {
			return false, errors.Wrapf(err, "copying file %s", file.Filename)
		}
	}
	// Perform compilation
	result, messages := action.Perform(dir)
	c.Sub.CompilerOutput = messages

	if result {
		// Success!
		output, err := ioutil.ReadFile(filepath.Join(dir, action.Output))
		if err != nil {
			return false, errors.WithStack(err)
		}
		c.Sub.CompiledSource = output
	} else {
		c.Sub.CompiledSource = nil
		c.Sub.Verdict = "Compile Error"
	}
	log.Printf("[WORKER] Compiling submission %v succeeded (result = %v).", c.Sub.ID, result)

	return result, c.Sub.Write(c.DB)
}

// CompileAction is an action revolving writing the source into a file in "Source",
// compile it with "Command" and taking the "Output" as the result.
type CompileAction struct {
	Source  string
	Command []*exec.Cmd
	Output  string
}

// Perform performs the compile action on the given directory.
// The directory MUST contain all files given by the Problem, PLUS the written "Source" file.
func (c *CompileAction) Perform(cwd string) (succeeded bool, messages []byte) {
	allOutputs := bytes.Buffer{}
	for _, cmd := range c.Command {
		allOutputs.WriteString(fmt.Sprintf("%s:\n", strings.Join(cmd.Args, " ")))
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
			return false, allOutputs.Bytes()
		}
	}
	return true, allOutputs.Bytes()
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
		Source:  source,
		Command: []*exec.Cmd{exec.Command("./" + batch)},
		Output:  "code",
	}, batch, nil
}

// CompileSingle creates a compilation command for a single source code file.
// Sometimes this is as simple as "copy".
func CompileSingle(l models.Language) (*CompileAction, error) {
	switch l {
	case models.LanguageCpp:
		return &CompileAction{
			Source:  "code.cc",
			Command: []*exec.Cmd{exec.Command("g++", "-std=c++17", "-O2", "-s", "-lm", "-DONLINE_JUDGE", "-DKJUDGE", "-o", "code", "code.cc")},
			Output:  "code",
		}, nil
	case models.LanguageGo:
		return &CompileAction{
			Source:  "code.go",
			Command: []*exec.Cmd{exec.Command("go", "build", "-buildmode=exe", "-tags", "online_judge,kjudge", "-o", "code", "code.go")},
			Output:  "code",
		}, nil
	case models.LanguageJava:
		return &CompileAction{
			Source: "code.java",
			Command: []*exec.Cmd{
				exec.Command("javac", "-d", ".", "code.java"),
				exec.Command("sh", "-c", "jar cf code *.class"),
				exec.Command("sh", "-c", "rm *.class"),
			},
			Output: "code",
		}, nil
	case models.LanguagePas:
		return &CompileAction{
			Source: "code.pas",
			Command: []*exec.Cmd{
				exec.Command("fpc", "-O3", "-dONLINE_JUDGE", "-dKJUDGE", "-ocode", "code.pas"),
			},
			Output: "code",
		}, nil
	case models.LanguageRust:
		return &CompileAction{
			Source: "code.rs",
			Command: []*exec.Cmd{
				exec.Command("rustc", "-O", "--cfg", "online_judge", "--cfg", "kjudge", "-o", "code", "code.rs"),
			},
			Output: "code",
		}, nil
	case models.LanguagePy2:
		return &CompileAction{
			Source: "code.py",
			Command: []*exec.Cmd{
				exec.Command("python2", "-m", "py_compile", "code.py"),
			},
			Output: "code.pyc",
		}, nil
	case models.LanguagePy3:
		return &CompileAction{
			Source:  "code.py",
			Command: []*exec.Cmd{exec.Command("python3", "-c", `import py_compile as m; m.compile("code.py", "code.pyc", doraise=True)`)},
			Output:  "code.pyc",
		}, nil
	default:
		return nil, errors.Errorf("Unknown language: %v", l)
	}
}
