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
	"git.nkagami.me/natsukagami/kjudge/tests"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

func getTestGroup(db db.DBContext, c echo.Context) (*models.TestGroup, error) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return nil, echo.ErrNotFound
	}
	tg, err := models.GetTestGroup(db, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, echo.ErrNotFound
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
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
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
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
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
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	file, err := readFromForm("file", mp)
	if err != nil {
		return err
	}
	tests, err := tests.Unpack(bytes.NewReader(file), int64(len(file)), c.FormValue("input"), c.FormValue("output"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := tg.WriteTests(tx, tests, override); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/admin/problems/%d", tg.ProblemID))
}

func readFromForm(name string, form *multipart.Form) ([]byte, error) {
	file, ok := form.File[name]
	if !ok {
		return nil, echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("file %s not found", name))
	}
	if len(file) != 1 {
		return nil, echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("file %s: expected one file, got %d", name, len(file)))
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
