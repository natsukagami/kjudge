package admin

import (
	"bytes"
	"database/sql"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"strconv"

	"git.nkagami.me/natsukagami/kjudge/db"
	"git.nkagami.me/natsukagami/kjudge/models"
	"git.nkagami.me/natsukagami/kjudge/server/httperr"
	"git.nkagami.me/natsukagami/kjudge/tests"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

func getTestGroup(db db.DBContext, c echo.Context) (*models.TestGroup, error) {
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
	return tg, nil
}

// TestGroupUploadSingle implements POST /admin/test_groups/:id/upload_single.
func (g *Group) TestGroupUploadSingle(c echo.Context) error {
	tx, err := g.db.Beginx()
	if err != nil {
		return errors.WithStack(err)
	}
	defer tx.Rollback()
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
	output, err := readFromForm("output", mp)
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
	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/admin/problems/%d", tg.ProblemID))
}

// TestGroupUploadMultiple implements POST /admin/test_groups/:id/upload_multiple
func (g *Group) TestGroupUploadMultiple(c echo.Context) error {
	tx, err := g.db.Beginx()
	if err != nil {
		return errors.WithStack(err)
	}
	defer tx.Rollback()
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
	if err := tg.WriteTests(tx, tests, override); err != nil {
		return httperr.BadRequestf("Cannot write tests: %v", err)
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/admin/problems/%d", tg.ProblemID))
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
	form.Bind(tg)
	if err := tg.Write(g.db); err != nil {
		return httperr.BadRequestf("Cannot update test group: %v", err)
	}
	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/admin/problems/%d", tg.ProblemID))
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
	content, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return content, nil
}
