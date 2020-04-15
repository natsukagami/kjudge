package admin

import (
	"bytes"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/natsukagami/kjudge/db"
	"github.com/natsukagami/kjudge/models"
	"github.com/natsukagami/kjudge/server/httperr"
	"github.com/natsukagami/kjudge/worker"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

// SubmissionCtx is the context for rendering the submission interface.
type SubmissionCtx struct {
	Submission *models.Submission

	Problem     *models.Problem
	Contest     *models.Contest
	TestGroups  []*models.TestGroupWithTests
	TestResults map[int]*models.TestResult
}

// Render renders the context.
func (s *SubmissionCtx) Render(c echo.Context) error {
	return c.Render(http.StatusOK, "admin/submission", s)
}

// Collect a SubmissionCtx.
func getSubmissionCtx(db db.DBContext, c echo.Context) (*SubmissionCtx, error) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return nil, httperr.NotFoundf("Submission not found: %v", idStr)
	}
	sub, err := models.GetSubmission(db, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, httperr.NotFoundf("Submission not found: %v", idStr)
	} else if err != nil {
		return nil, err
	}

	problem, err := models.GetProblem(db, sub.ProblemID)
	if err != nil {
		return nil, err
	}
	contest, err := models.GetContest(db, problem.ContestID)
	if err != nil {
		return nil, err
	}

	ctx := &SubmissionCtx{Submission: sub, Problem: problem, Contest: contest}

	if sub.Score.Valid {
		testGroups, err := models.GetProblemTestsMeta(db, problem.ID)
		if err != nil {
			return nil, err
		}
		testResults, err := models.GetSubmissionTestResults(db, sub.ID)
		if err != nil {
			return nil, err
		}
		trMap := make(map[int]*models.TestResult)
		for _, tr := range testResults {
			trMap[tr.TestID] = tr
		}
		ctx.TestGroups = testGroups
		ctx.TestResults = trMap
	}

	return ctx, nil
}

// SubmissionGet implement GET /admin/submissions/:id
func (g *Group) SubmissionGet(c echo.Context) error {
	ctx, err := getSubmissionCtx(g.db, c)
	if err != nil {
		return err
	}
	return ctx.Render(c)
}

// SubmissionBinaryGet implements GET /admin/submissions/:id/binary
func (g *Group) SubmissionBinaryGet(c echo.Context) error {
	ctx, err := getSubmissionCtx(g.db, c)
	if err != nil {
		return err
	}
	if ctx.Submission.CompiledSource != nil {
		http.ServeContent(c.Response(), c.Request(), fmt.Sprintf("compiled_s%d", ctx.Submission.ID), ctx.Submission.SubmittedAt, bytes.NewReader(ctx.Submission.CompiledSource))
		return nil
	} else {
		return httperr.BadRequestf("Compiled binary not available")
	}
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
