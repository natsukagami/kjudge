package admin

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"git.nkagami.me/natsukagami/kjudge/models"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

// RejudgePost implements POST /admin/rejudge.
func (g *Group) RejudgePost(c echo.Context) error {
	stage := c.FormValue("stage")
	idStr := strings.Split(c.FormValue("id"), ",")
	var id []int
	for _, i := range idStr {
		v, err := strconv.Atoi(i)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("id `%s`: %v", i, err.Error()))
		}
		id = append(id, v)
	}
	tx, err := g.db.Beginx()
	if err != nil {
		return errors.WithStack(err)
	}
	defer tx.Rollback()

	switch stage {
	case "score":
		err = models.RejudgeScore(tx, id...)
	case "run":
		err = models.RejudgeRun(tx, id...)
	case "compile":
		err = models.RejudgeCompile(tx, id...)
	default:
		err = echo.NewHTTPError(http.StatusBadRequest, "Invalid rejudge stage")
	}
	if err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return errors.WithStack(err)
	}
	last := c.FormValue("last")
	if last == "" {
		last = "/admin/submissions"
	}
	return c.Redirect(http.StatusSeeOther, last)
}
