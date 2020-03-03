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

// OptionalInt64 implements a NullInt64 with binding capabilities.
type OptionalInt64 struct {
	sql.NullInt64
}

func (o OptionalInt64) String() string {
	if o.Valid {
		return fmt.Sprintf("%d", o.Int64)
	}
	return ""
}

// UnmarshalParam implement echo's Bind.
func (o *OptionalInt64) UnmarshalParam(src string) error {
	if src == "" {
		o.Valid = false
		return nil
	}
	n, err := strconv.Atoi(src)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "expected a number, number not given")
	}
	o.Valid = true
	o.Int64 = int64(n)
	return nil
}

type TestGroupForm struct {
	MemoryLimit OptionalInt64          `form:"memory_limit"`
	Name        string                 `form:"name"`
	Score       float64                `form:"score"`
	ScoringMode models.TestScoringMode `form:"scoring_mode"`
	TimeLimit   OptionalInt64          `form:"time_limit"`
}

// Bind binds the form's values to the TestGroup.
func (f *TestGroupForm) Bind(t *models.TestGroup) {
	t.Name = f.Name
	t.Score = f.Score
	t.ScoringMode = f.ScoringMode
	t.TimeLimit = f.TimeLimit.NullInt64
	t.MemoryLimit = f.MemoryLimit.NullInt64
}

// TestGroup is the wrapper for a TestGroupWithTests, with to-form conversion.
type TestGroup struct {
	*models.TestGroupWithTests
}

func (t TestGroup) ToForm() TestGroupForm {
	return TestGroupForm{
		Name:        t.Name,
		Score:       t.Score,
		ScoringMode: t.ScoringMode,
		MemoryLimit: OptionalInt64{t.MemoryLimit},
		TimeLimit:   OptionalInt64{t.TimeLimit},
	}
}

// Collect the ID and get the corresponding problem.
func (g *Group) getProblem(c echo.Context) (*ProblemCtx, error) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return nil, echo.ErrNotFound
	}
	problem, err := models.GetProblem(g.db, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, echo.ErrNotFound
	} else if err != nil {
		return nil, err
	}
	contest, err := models.GetContest(g.db, problem.ContestID)
	if err != nil {
		return nil, err
	}
	tests, err := models.GetProblemTestsMeta(g.db, problem.ID)
	if err != nil {
		return nil, err
	}
	var testGroups []TestGroup
	for _, t := range tests {
		testGroups = append(testGroups, TestGroup{t})
	}
	return &ProblemCtx{Problem: problem, Contest: contest, TestGroups: testGroups}, err
}

// ProblemCtx is the context for rendering admin/problem.
type ProblemCtx struct {
	*models.Problem
	Contest    *models.Contest
	TestGroups []TestGroup

	// Edit Problem Form
	EditForm      ProblemForm
	EditFormError error

	// New TestGroup form
	TestGroupForm      TestGroupForm
	TestGroupFormError error
}

// ProblemGet implements GET /admin/problems/:id
func (g *Group) ProblemGet(c echo.Context) error {
	ctx, err := g.getProblem(c)
	if err != nil {
		return err
	}
	ctx.EditForm = ProblemToForm(ctx.Problem)
	return g.problemRender(ctx, c)
}

// Render the context.
func (g *Group) problemRender(ctx *ProblemCtx, c echo.Context) error {
	status := http.StatusOK
	if ctx.EditFormError != nil || ctx.TestGroupFormError != nil {
		status = http.StatusBadRequest
	}
	return c.Render(status, "admin/problem", ctx)
}

// ProblemEdit implements POST /admin/problems/:id
func (g *Group) ProblemEdit(c echo.Context) error {
	ctx, err := g.getProblem(c)
	if err != nil {
		return err
	}
	if err := c.Bind(&ctx.EditForm); err != nil {
		return err
	}
	nw := *ctx.Problem
	ctx.EditForm.Bind(&nw)
	if err := nw.Write(g.db); err != nil {
		ctx.EditFormError = err
		return g.problemRender(ctx, c)
	}
	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/admin/problems/%d", nw.ID))
}

// ProblemAddTestGroup implements /admin/problems/:id/add_test_group
func (g *Group) ProblemAddTestGroup(c echo.Context) error {
	ctx, err := g.getProblem(c)
	if err != nil {
		return err
	}
	if err := c.Bind(&ctx.TestGroupForm); err != nil {
		return err
	}
	var tg models.TestGroup
	ctx.TestGroupForm.Bind(&tg)
	tg.ProblemID = ctx.Problem.ID
	if err := tg.Write(g.db); err != nil {
		ctx.EditForm = ProblemToForm(ctx.Problem)
		ctx.TestGroupFormError = err
		return g.problemRender(ctx, c)
	}
	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/admin/problems/%d", ctx.Problem.ID))
}

// ProblemDelete implements POST /admin/problems/:id/delete
func (g *Group) ProblemDelete(c echo.Context) error {
	ctx, err := g.getProblem(c)
	if err != nil {
		return err
	}
	if err := ctx.Problem.Delete(g.db); err != nil {
		return err
	}
	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/admin/contests/%d", ctx.Contest.ID))
}
