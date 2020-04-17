package admin

import (
	"bytes"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/natsukagami/kjudge/db"
	"github.com/natsukagami/kjudge/models"
	"github.com/natsukagami/kjudge/server/httperr"
	"github.com/natsukagami/kjudge/worker"
	"github.com/pkg/errors"
)

func getFile(db db.DBContext, c echo.Context) (*models.File, error) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return nil, httperr.NotFoundf("File not found: %s", idStr)
	}
	file, err := models.GetFile(db, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, httperr.NotFoundf("File not found: %d", id)
	} else if err != nil {
		return nil, err
	}
	return file, nil
}

// FileGet implements GET /admin/files/:id
func (g *Group) FileGet(c echo.Context) error {
	file, err := getFile(g.db, c)
	if err != nil {
		return err
	}
	http.ServeContent(c.Response(), c.Request(), file.Filename, time.Now(), bytes.NewReader(file.Content))
	return nil
}

// FileDelete implements POST /admin/files/:id/delete
func (g *Group) FileDelete(c echo.Context) error {
	file, err := getFile(g.db, c)
	if err != nil {
		return err
	}
	if err := file.Delete(g.db); err != nil {
		return err
	}
	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/admin/problems/%d#files", file.ProblemID))
}

// FileCompile implements POST /admin/files/:id/compile
func (g *Group) FileCompile(c echo.Context) error {
	tx, err := g.db.Beginx()
	if err != nil {
		return errors.WithStack(err)
	}
	defer db.Rollback(tx)

	file, err := getFile(tx, c)
	if err != nil {
		return err
	}
	if !file.Compilable() {
		return httperr.BadRequestf("File is not a compilable file.")
	}

	// Collect all files
	files, err := models.GetProblemFiles(tx, file.ProblemID)
	if err != nil {
		return err
	}

	output, err := worker.CustomCompile(file, files)
	if err != nil {
		return httperr.BadRequestf("%v", err)
	}
	output.ProblemID = file.ProblemID

	// Check if there is a filename conflict
	if f, err := models.GetFileWithName(tx, file.ProblemID, output.Filename); err == nil {
		if err := f.Delete(tx); err != nil {
			return err
		}
	} else if !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	if err := output.Write(tx); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return errors.WithStack(err)
	}
	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/admin/problems/%d#files", file.ProblemID))
}
