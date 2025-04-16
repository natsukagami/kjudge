package admin

import (
	"archive/zip"
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/natsukagami/kjudge/db"
	"github.com/natsukagami/kjudge/models"
	"github.com/natsukagami/kjudge/server/httperr"
	"github.com/natsukagami/kjudge/tests"
	"github.com/pkg/errors"
)

// TestGroupCtx is the context for rendering test-group.
type TestGroupCtx struct {
	*models.TestGroupWithTests
	Contest *models.Contest
	Problem *models.Problem
}

func getTestGroup(db db.DBContext, c echo.Context) (*TestGroupCtx, error) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return nil, httperr.NotFoundf("Test group not found: %v", idStr)
	}
	tg, err := models.GetTestGroup(db, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, httperr.NotFoundf("Test group not found: %v", idStr)
	} else if err != nil {
		return nil, err
	}
	tests, err := models.GetTestGroupTests(db, tg.ID)
	if err != nil {
		return nil, err
	}
	problem, err := models.GetProblem(db, tg.ProblemID)
	if err != nil {
		return nil, err
	}
	contest, err := models.GetContest(db, problem.ContestID)
	if err != nil {
		return nil, err
	}
	return &TestGroupCtx{
		TestGroupWithTests: &models.TestGroupWithTests{
			TestGroup: tg,
			Tests:     tests,
		},
		Problem: problem,
		Contest: contest,
	}, nil
}

// Render renders the context.
func (ctx *TestGroupCtx) Render(c echo.Context) error {
	return c.Render(http.StatusOK, "admin/test_group", ctx)
}

// ToForm converts the context into a nicer form format.
func (ctx *TestGroupCtx) ToForm() TestGroupForm {
	return TestGroupForm{
		Name:        ctx.Name,
		Score:       ctx.Score,
		ScoringMode: ctx.ScoringMode,
		MemoryLimit: OptionalInt64{ctx.MemoryLimit},
		TimeLimit:   OptionalInt64{ctx.TimeLimit},
	}
}

// TestGroupGet implements GET /admin/test_groups/:id
func (g *Group) TestGroupGet(c echo.Context) error {
	ctx, err := getTestGroup(g.db, c)
	if err != nil {
		return err
	}
	return ctx.Render(c)
}

// TestGroupUploadSingle implements POST /admin/test_groups/:id/upload_single.
func (g *Group) TestGroupUploadSingle(c echo.Context) error {
	tx, err := g.db.Beginx()
	if err != nil {
		return errors.WithStack(err)
	}
	defer db.Rollback(tx)
	tg, err := getTestGroup(tx, c)
	if err != nil {
		return err
	}
	// Parse the form
	name := c.FormValue("name")
	mp, err := c.MultipartForm()
	if err != nil {
		return httperr.BindFail(err)
	}
	input, err := readFromForm("input", mp)
	if err != nil {
		return err
	}
	input, err = models.NormalizeEndingsUnix(input)
	if err != nil {
		return err
	}
	output, err := readFromForm("output", mp)
	if err != nil {
		return err
	}
	output, err = models.NormalizeEndingsUnix(output)
	if err != nil {
		return err
	}
	// Make the test
	test := &models.Test{
		TestGroupID: tg.ID,
		Name:        name,
		Input:       input,
		Output:      output,
	}
	if err := test.Write(tx); err != nil {
		return httperr.BadRequestf("Cannot write test: %v", err)
	}
	if err := tx.Commit(); err != nil {
		return errors.WithStack(err)
	}
	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/admin/test_groups/%d", tg.ID))
}

// TestGroupUploadMultiple implements POST /admin/test_groups/:id/upload_multiple
func (g *Group) TestGroupUploadMultiple(c echo.Context) error {
	tx, err := g.db.Beginx()
	if err != nil {
		return errors.WithStack(err)
	}
	defer db.Rollback(tx)
	tg, err := getTestGroup(tx, c)
	if err != nil {
		return err
	}

	override := c.FormValue("override") == "true"
	mp, err := c.MultipartForm()
	if err != nil {
		return httperr.BindFail(err)
	}
	file, err := readFromForm("file", mp)
	if err != nil {
		return err
	}
	tests, err := tests.Unpack(bytes.NewReader(file), int64(len(file)), c.FormValue("input"), c.FormValue("output"))
	if err != nil {
		return httperr.BadRequestf("cannot unpack tests: %v", err)
	}
	if err := tg.WriteTestsNormalized(tx, tests, override); err != nil {
		return httperr.BadRequestf("Cannot write tests: %v", err)
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/admin/test_groups/%d", tg.ID))
}

// TestGroupEdit implements POST /admin/test_groups/:id
func (g *Group) TestGroupEdit(c echo.Context) error {
	tg, err := getTestGroup(g.db, c)
	if err != nil {
		return err
	}
	var form TestGroupForm
	if err := c.Bind(&form); err != nil {
		return httperr.BindFail(err)
	}
	form.Bind(tg.TestGroupWithTests.TestGroup)
	if err := tg.Write(g.db); err != nil {
		return httperr.BadRequestf("Cannot update test group: %v", err)
	}
	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/admin/test_groups/%d", tg.ID))
}

// TestGroupDelete implements POST /admin/test_groups/:id/delete
func (g *Group) TestGroupDelete(c echo.Context) error {
	tg, err := getTestGroup(g.db, c)
	if err != nil {
		return err
	}
	if err := tg.Delete(g.db); err != nil {
		return err
	}
	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/admin/problems/%d", tg.ProblemID))
}

func readFromForm(name string, form *multipart.Form) ([]byte, error) {
	file, ok := form.File[name]
	if !ok {
		return nil, httperr.BadRequestf("file %s not found", name)
	}
	if len(file) != 1 {
		return nil, httperr.BadRequestf("file %s: expected one file, got %d", name, len(file))
	}
	f, err := file[0].Open()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer f.Close()
	content, err := io.ReadAll(f)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return content, nil
}

// TestGroupRejudgePost implements POST /admin/test_groups/:id/rejudge
func (g *Group) TestGroupRejudgePost(c echo.Context) error {
	tg, err := getTestGroup(g.db, c)
	if err != nil {
		return err
	}
	tx, err := g.db.Beginx()
	if err != nil {
		return errors.WithStack(err)
	}
	defer db.Rollback(tx)
	subs, err := models.GetProblemSubmissions(tx, tg.ProblemID)
	if err != nil {
		return err
	}
	var id []int
	for _, sub := range subs {
		id = append(id, sub.ID)
	}
	// First we remove all the results related to a test group.
	if err := tg.DeleteResults(tx); err != nil {
		return err
	}
	// we still reset the score
	if err := models.RejudgeScore(tx, id...); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return errors.WithStack(err)
	}
	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/admin/problems/%d/submissions", tg.ProblemID))
}

// WriteTestsNormalized normalizes line endings and writes the given set of
// tests into the Database. If override is set, all tests in the test group
// gets deleted first. The LazyTests are STILL invalid models.Tests. DO NOT USE.
func (r *TestGroupCtx) WriteTestsNormalized(db db.DBContext, tests []*tests.LazyTest, override bool) error {
	for _, test := range tests {
		test.TestGroupID = r.ID
		if err := test.Verify(); err != nil {
			return errors.Wrapf(err, "test `%s`", test.Name)
		}
	}
	if override {
		if _, err := db.Exec("DELETE FROM tests WHERE test_group_id = ?", r.ID); err != nil {
			return errors.WithStack(err)
		}
	}
	for _, test := range tests {
		input, err := readZip(test.Input)
		if err != nil {
			return errors.Wrapf(err, "test %v input", test.Name)
		}
		input, err = models.NormalizeEndingsUnix(input)
		if err != nil {
			return errors.Wrapf(err, "test %v input", test.Name)
		}
		output, err := readZip(test.Output)
		if err != nil {
			return errors.Wrapf(err, "test %v output", test.Name)
		}
		output, err = models.NormalizeEndingsUnix(output)
		if err != nil {
			return errors.Wrapf(err, "test %v output", test.Name)
		}
		if _, err := db.Exec(
			"INSERT INTO tests(name, test_group_id, input, output) VALUES (?, ?, ?, ?)",
			test.Name,
			test.TestGroupID,
			input,
			output,
		); err != nil {
			return errors.Wrapf(err, "inserting test `%s`", test.Name)
		}
	}
	return nil
}

func readZip(f *zip.File) ([]byte, error) {
	reader, err := f.Open()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer reader.Close()
	res, err := io.ReadAll(reader)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return res, nil
}
