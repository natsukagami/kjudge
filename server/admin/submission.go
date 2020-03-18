package admin

import (
	"database/sql"
	"net/http"
	"strconv"

	"git.nkagami.me/natsukagami/kjudge/db"
	"git.nkagami.me/natsukagami/kjudge/models"
	"git.nkagami.me/natsukagami/kjudge/worker"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

// SubmissionCtx is the context for rendering the submission interface.
type SubmissionCtx struct {
	Submission *models.Submission
}

// Collect a SubmissionCtx.
func getSubmissionCtx(db db.DBContext, c echo.Context) (*SubmissionCtx, error) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return nil, echo.ErrNotFound
	}
	sub, err := models.GetSubmission(db, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, echo.ErrNotFound
	} else if err != nil {
		return nil, err
	}
	return &SubmissionCtx{Submission: sub}, nil
}

// SubmissionVerdictGet implements GET /admin/submissions/:id/verdict
func (g *Group) SubmissionVerdictGet(c echo.Context) error {
	ctx, err := getSubmissionCtx(g.db, c)
	if err != nil {
		return err
	}
	if ctx.Submission.Verdict == "..." || ctx.Submission.Verdict == worker.VerdictCompileError {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"verdict": ctx.Submission.Verdict,
		})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"verdict": ctx.Submission.Verdict,
		"score":   ctx.Submission.Score.Float64,
		"penalty": ctx.Submission.Penalty.Int64,
	})
}
