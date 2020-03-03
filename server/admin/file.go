package admin

import (
	"bytes"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"git.nkagami.me/natsukagami/kjudge/db"
	"git.nkagami.me/natsukagami/kjudge/models"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

func getFile(db db.DBContext, c echo.Context) (*models.File, error) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return nil, echo.ErrNotFound
	}
	file, err := models.GetFile(db, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, echo.ErrNotFound
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
