package compile

import (
	"github.com/natsukagami/kjudge/models"
	"github.com/pkg/errors"
)

// getCompileSingleAction creates a compilation command for a single source code file.
// Sometimes this is as simple as "copy".
func getCompileSingleAction(l models.Language) (*CompileAction, error) {
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

func CompileSingle(c *CompileContext) (*CompileAction, error) {
	return getCompileSingleAction(c.Sub.Language)
}
