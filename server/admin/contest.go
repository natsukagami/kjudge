package admin

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/natsukagami/kjudge/db"
	"github.com/natsukagami/kjudge/models"
	"github.com/natsukagami/kjudge/server/httperr"
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

// ContestCtx is the context for rendering admin/contest.
type ContestCtx struct {
	*models.Contest
	FormError error
	Form      ContestForm

	Problems         []*models.Problem
	ProblemForm      ProblemForm
	ProblemFormError error
}

func getContest(db db.DBContext, c echo.Context) (*ContestCtx, error) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return nil, httperr.NotFoundf("Contest not found: %s", idStr)
	}
	contest, err := models.GetContest(db, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, httperr.NotFoundf("Contest not found: %d", id)
	} else if err != nil {
		return nil, err
	}
	problems, err := models.GetContestProblems(db, contest.ID)
	if err != nil {
		return nil, err
	}
	return &ContestCtx{
		Contest:  contest,
		Problems: problems,
		Form:     *ContestToForm(contest),
		ProblemForm: ProblemForm{
			TimeLimit:     1000,
			MemoryLimit:   262144,
			ScoringMode:   models.ScoringModeBest,
			PenaltyPolicy: models.PenaltyPolicyNone,
		},
	}, nil
}

func (ctx *ContestCtx) Render(c echo.Context) error {
	code := http.StatusOK
	if ctx.FormError != nil || ctx.ProblemFormError != nil {
		code = http.StatusBadRequest
	}
	return c.Render(code, "admin/contest", ctx)
}

// ContestGet implements GET /admin/contest/:id
func (g *Group) ContestGet(c echo.Context) error {
	ctx, err := getContest(g.db, c)
	if err != nil {
		return err
	}
	return ctx.Render(c)
}

// ContestDelete implement POST /admin/contest/:id/delete
func (g *Group) ContestDelete(c echo.Context) error {
	ctx, err := getContest(g.db, c)
	if err != nil {
		return err
	}
	if err := ctx.Delete(g.db); err != nil {
		return err
	}
	return c.Redirect(http.StatusSeeOther, "/admin/contests")
}

// ContestEdit implements POST /admin/contest/:id
func (g *Group) ContestEdit(c echo.Context) error {
	ctx, err := getContest(g.db, c)
	if err != nil {
		return err
	}
	nw := *ctx.Contest
	var form ContestForm
	if err := c.Bind(&form); err != nil {
		return httperr.BindFail(err)
	}
	form.Bind(&nw)
	if err := nw.Write(g.db); err != nil {
		ctx.Form = form
		ctx.FormError = err
		return ctx.Render(c)
	}
	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/admin/contests/%d", ctx.ID))
}

// ContestAddProblem implements POST /admin/contest/:id/add_problem
func (g *Group) ContestAddProblem(c echo.Context) error {
	ctx, err := getContest(g.db, c)
	if err != nil {
		return err
	}
	var (
		form    ProblemForm
		problem models.Problem
	)
	if err := c.Bind(&form); err != nil {
		return httperr.BindFail(err)
	}
	problem.ContestID = ctx.ID
	form.Bind(&problem)
	if err := problem.Write(g.db); err != nil {
		ctx.ProblemForm = form
		ctx.ProblemFormError = err
		return ctx.Render(c)
	}
	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/admin/problems/%d", problem.ID))
}

// ContestSubmissionsGet implements GET /admin/contests/:id/submissions.
func (g *Group) ContestSubmissionsGet(c echo.Context) error {
	ctx, err := getContest(g.db, c)
	if err != nil {
		return err
	}
	subs, err := SubmissionsBy(g.db, nil, ctx.Contest, nil)
	if err != nil {
		return err
	}
	return c.Render(http.StatusOK, "admin/contest_submissions", subs)
}

// ContestRejudgePost implements POST /admin/contests/:id/rejudge
func (g *Group) ContestRejudgePost(c echo.Context) error {
	ctx, err := getContest(g.db, c)
	if err != nil {
		return err
	}
	tx, err := g.db.Beginx()
	if err != nil {
		return errors.WithStack(err)
	}
	defer tx.Rollback()
	var problemIDs []int
	for _, p := range ctx.Problems {
		problemIDs = append(problemIDs, p.ID)
	}
	subs, err := models.GetProblemsSubmissions(tx, problemIDs...)
	if err != nil {
		return err
	}
	var id []int
	for _, sub := range subs {
		id = append(id, sub.ID)
	}
	if err := DoRejudge(tx, id, c.FormValue("stage")); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return errors.WithStack(err)
	}
	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/admin/contests/%d/submissions", ctx.ID))
}
