package admin

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/natsukagami/kjudge/models"
	"github.com/natsukagami/kjudge/server/httperr"
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
		return httperr.BadRequestf("expected a number, number not given")
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

// Collect the ID and get the corresponding problem.
func (g *Group) getProblem(c echo.Context) (*ProblemCtx, error) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return nil, httperr.NotFoundf("Problem not found: %v", idStr)
	}
	problem, err := models.GetProblem(g.db, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, httperr.NotFoundf("Problem not found: %v", idStr)
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
	files, err := models.GetProblemFilesMeta(g.db, problem.ID)
	if err != nil {
		return nil, err
	}
	return &ProblemCtx{Problem: problem, Contest: contest, TestGroups: tests, Files: files}, err
}

// ProblemCtx is the context for rendering admin/problem.
type ProblemCtx struct {
	*models.Problem
	Contest    *models.Contest
	TestGroups []*models.TestGroupWithTests
	Files      []*models.File

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
	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/admin/test_groups/%d", tg.ID))
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

// ProblemAddFile implements POST /admin/problems/:id/add_file
func (g *Group) ProblemAddFile(c echo.Context) error {
	ctx, err := g.getProblem(c)
	if err != nil {
		return err
	}
	makePublic := c.FormValue("public") == "true"
	form, err := c.MultipartForm()
	if err != nil {
		return httperr.BindFail(err)
	}
	var files []*models.File
	for _, file := range form.File["file"] {
		r, err := file.Open()
		if err != nil {
			return errors.Wrapf(err, "file %s", file.Filename)
		}
		defer r.Close()
		content, err := ioutil.ReadAll(r)
		if err != nil {
			return errors.Wrapf(err, "file %s", file.Filename)
		}
		files = append(files, &models.File{
			Filename: file.Filename,
			Content:  content,
			Public:   makePublic,
		})
	}
	rename := c.FormValue("filename")
	if rename != "" && len(files) == 1 {
		files[0].Filename = rename
	}
	if err := ctx.Problem.WriteFiles(g.db, files); err != nil {
		return httperr.BadRequestf("cannot write files: %v", err)
	}
	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/admin/problems/%d#files", ctx.Problem.ID))
}

// ProblemSubmissionsGet implements GET /admin/problems/:id/submissions.
func (g *Group) ProblemSubmissionsGet(c echo.Context) error {
	p, err := g.getProblem(c)
	if err != nil {
		return err
	}
	subs, err := SubmissionsBy(g.db, nil, nil, p.Problem)
	if err != nil {
		return err
	}
	return c.Render(http.StatusOK, "admin/problem_submissions", subs)
}

// ProblemRejudgePost implements POST /admin/problems/:id/rejudge
func (g *Group) ProblemRejudgePost(c echo.Context) error {
	p, err := g.getProblem(c)
	if err != nil {
		return err
	}
	tx, err := g.db.Beginx()
	if err != nil {
		return errors.WithStack(err)
	}
	defer tx.Rollback()
	subs, err := models.GetProblemSubmissions(tx, p.Problem.ID)
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
	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/admin/problems/%d/submissions", p.Problem.ID))
}
