package admin

import (
	"database/sql"
	"net/http"
	"strconv"

	"git.nkagami.me/natsukagami/kjudge/models"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

// Collect the ID and get the corresponding problem.
func (g *Group) getProblem(c echo.Context) (*models.Problem, *models.Contest, error) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return nil, nil, echo.ErrNotFound
	}
	problem, err := models.GetProblem(g.db, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil, echo.ErrNotFound
	} else if err != nil {
		return nil, nil, err
	}
	contest, err := models.GetContest(g.db, problem.ContestID)
	return problem, contest, err
}

// ProblemCtx is the context for rendering admin/problem.
type ProblemCtx struct {
	*models.Problem
	Contest *models.Contest

	// Edit Problem Form
	EditForm      ProblemForm
	EditFormError error
}

// ProblemGet implements GET /admin/problems/:id
func (g *Group) ProblemGet(c echo.Context) error {
	problem, contest, err := g.getProblem(c)
	if err != nil {
		return err
	}
	return g.problemRender(&ProblemCtx{Problem: problem, Contest: contest, EditForm: ProblemToForm(problem)}, c)
}

// Render the context.
func (g *Group) problemRender(ctx *ProblemCtx, c echo.Context) error {
	status := http.StatusOK
	if ctx.EditFormError != nil {
		status = http.StatusBadRequest
	}
	return c.Render(status, "admin/problem", ctx)
}
