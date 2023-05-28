package compile

// Compiling anything that's more compicated than single file:
//
// - Prepare a "compile_%s.%ext" file, with %s being the language (cc, go, rs, java, py2, py3, pas)
// - Prepare any more files as needed. They will all be put into the CWD of the script
// - The CWD also contains "code.%s" (%s being the language's respective extension) file, which is the contestant's source code.
// - The script should do whatever it wants (unsandboxed, because it's not my job to do so) within 20 seconds.
// - It should produce a single binary called "code" in the CWD.

import (
	"github.com/natsukagami/kjudge/models"
	"github.com/pkg/errors"
)

var batchFilenames = []string{
	"compile_cc.sh", "compile_go.sh", "compile_java.sh", "compile_rs.sh", "compile_pas.sh", "compile_py2.sh", "compile_py3.sh",
}

func verifyBatchScheme(c *CompileContext, expected string) (bool, error) {
	hasBatch := false
	for _, file := range c.Files {
		for _, f := range batchFilenames {
			if f == file.Filename {
				hasBatch = true
			}
		}
		if file.Filename == expected {
			return true, nil
		}
	}
	if hasBatch {
		// Batch compile mode enabled, but this language is not supported.
		c.Sub.CompiledSource = nil
		c.Sub.Verdict = models.VerdictCompileError
		c.Sub.CompilerOutput = []byte("Custom Compilers are not enabled for this language.")
		return false, c.Sub.Write(c.DB)
	}
	return false, nil
}

// getCompileBatchAction returns a compile action, along with the required batch filename
// in order to successfully compile.
func getCompileBatchAction(l models.Language) (*CompileAction, string, error) {
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

// CompileBatch returns nil if the problem is not setup for batch-compile
// and returns an action otherwise. Throws if batch-compilable but language
// of the submission is not supported (no batch file for specified language).
func CompileBatch(c *CompileContext) (*CompileAction, error) {
	action, batchFile, err := getCompileBatchAction(c.Sub.Language)
	if err != nil {
		return nil, err
	}
	compilable, err := verifyBatchScheme(c, batchFile)
	if !compilable || err != nil {
		return nil, err
	}

	return action, nil
}
