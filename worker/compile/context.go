package compile
import (
	"github.com/jmoiron/sqlx"
	"github.com/natsukagami/kjudge/models"
)

// CompileContext is the information needed to perform compilation.
type CompileContext struct {
	DB      *sqlx.Tx
	Sub     *models.Submission
	Problem *models.Problem
	Files 	[]*models.File
}

func NewCompileContext(DB *sqlx.Tx, Sub *models.Submission, Problem *models.Problem) (*CompileContext, error) {
	files, err := models.GetProblemFiles(DB, Problem.ID)
	if err != nil {
		return nil, err
	}
	return &CompileContext{DB: DB, Sub: Sub, Problem: Problem, Files: files}, nil
}