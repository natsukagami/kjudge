package admin

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"git.nkagami.me/natsukagami/kjudge/models"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

// ProblemForm is a form for creating/updating a problem.
type ProblemForm struct {
	DisplayName   string               `form:"display_name"`
	MemoryLimit   int                  `form:"memory_limit"`
	Name          string               `form:"name"`
	PenaltyPolicy models.PenaltyPolicy `form:"penalty_policy"`
	ScoringMode   models.ScoringMode   `form:"scoring_mode"`
	TimeLimit     int                  `form:"time_limit"`
}

// Bind binds the form's content into the Problem.
func (f *ProblemForm) Bind(p *models.Problem) {
	p.DisplayName = f.DisplayName
	p.MemoryLimit = f.MemoryLimit
	p.Name = f.Name
	p.PenaltyPolicy = f.PenaltyPolicy
	p.ScoringMode = f.ScoringMode
	p.TimeLimit = f.TimeLimit
}

// ProblemForm produces an edit form from the problem.
func ProblemToForm(p *models.Problem) ProblemForm {
	var f ProblemForm
	f.DisplayName = p.DisplayName
	f.MemoryLimit = p.MemoryLimit
	f.Name = p.Name
	f.PenaltyPolicy = p.PenaltyPolicy
	f.ScoringMode = p.ScoringMode
	f.TimeLimit = p.TimeLimit
	return f
}

func (g *Group) getContest(c echo.Context) (*models.Contest, error) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return nil, echo.ErrNotFound
	}
	contest, err := models.GetContest(g.db, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, echo.ErrNotFound
	} else if err != nil {
		return nil, err
	}
	return contest, nil
}

// ContestCtx is the context for rendering admin/contest.
type ContestCtx struct {
	*models.Contest
	FormError error
	Form      ContestForm

	Problems         []*models.Problem
	ProblemForm      ProblemForm
	ProblemFormError error
}

// ContestGet implements GET /admin/contest/:id
func (g *Group) ContestGet(c echo.Context) error {
	contest, err := g.getContest(c)
	if err != nil {
		return err
	}
	ctx := &ContestCtx{Contest: contest, Form: *ContestToForm(contest)}
	return g.contestGetRender(ctx, c)
}

func (g *Group) contestGetRender(ctx *ContestCtx, c echo.Context) error {
	problems, err := models.GetContestProblems(g.db, ctx.Contest.ID)
	if err != nil {
		return err
	}
	ctx.Problems = problems
	code := http.StatusOK
	if ctx.FormError != nil || ctx.ProblemFormError != nil {
		code = http.StatusBadRequest
	}
	return c.Render(code, "admin/contest", ctx)
}

// ContestDelete implement POST /admin/contest/:id/delete
func (g *Group) ContestDelete(c echo.Context) error {
	contest, err := g.getContest(c)
	if err != nil {
		return err
	}
	if err := contest.Delete(g.db); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.Redirect(http.StatusSeeOther, "/admin/contests")
}

// ContestEdit implements POST /admin/contest/:id
func (g *Group) ContestEdit(c echo.Context) error {
	contest, err := g.getContest(c)
	if err != nil {
		return err
	}
	original := *contest
	var form ContestForm
	if err := c.Bind(&form); err != nil {
		return err
	}
	form.Bind(contest)
	if err := contest.Write(g.db); err != nil {
		return g.contestGetRender(&ContestCtx{Contest: &original, Form: form, FormError: err}, c)
	}
	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/admin/contests/%d", contest.ID))
}

// ContestAddProblem implements POST /admin/contest/:id/add_problem
func (g *Group) ContestAddProblem(c echo.Context) error {
	contest, err := g.getContest(c)
	if err != nil {
		return err
	}
	var (
		form    ProblemForm
		problem models.Problem
	)
	if err := c.Bind(&form); err != nil {
		return err
	}
	problem.ContestID = contest.ID
	form.Bind(&problem)
	if err := problem.Write(g.db); err != nil {
		return g.contestGetRender(&ContestCtx{Contest: contest, ProblemForm: form, ProblemFormError: err}, c)
	}
	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/admin/problems/%d", problem.ID))
}
