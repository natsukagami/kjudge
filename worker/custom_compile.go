package worker

import (
	"os"
	"path/filepath"

	"github.com/natsukagami/kjudge/models"
	"github.com/pkg/errors"
)

// CustomCompile tries to compile a file from the given source file.
func CustomCompile(source *models.File, files []*models.File) (*models.File, error) {
	ext := filepath.Ext(source.Filename)
	language, err := models.LanguageByExt(ext)
	if err != nil {
		return nil, err
	}
	// Find out the batch command
	action, err := CompileSingle(language)
	if err != nil {
		return nil, err
	}
	// Perform compilation
	dir, err := os.MkdirTemp("", "*")
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer action.Cleanup(dir)
	action.Source.Content = source.Content
	action.Files = files

	if err := action.Prepare(dir); err != nil {
		return nil, err
	}

	result, message := action.Perform(dir)
	if result {
		output, err := os.ReadFile(filepath.Join(dir, action.Output))
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return &models.File{
			Filename: source.Filename[:len(source.Filename)-len(ext)],
			Content:  output,
		}, nil
	} else {
		return nil, errors.Errorf("Compilation failed with message:\n%s", string(message))
	}
}
