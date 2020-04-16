package admin

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/natsukagami/kjudge/db"
	"github.com/natsukagami/kjudge/models"
	"github.com/natsukagami/kjudge/server/httperr"
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
			return httperr.BadRequestf("submission id `%s`: %v", i, err)
		}
		id = append(id, v)
	}
	tx, err := g.db.Beginx()
	if err != nil {
		return errors.WithStack(err)
	}
	defer tx.Rollback()

	if err := DoRejudge(tx, id, stage); err != nil {
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

// DoRejudge performs rejudge on a given stage and list of IDs.
func DoRejudge(db db.DBContext, id []int, stage string) error {
	var err error
	switch stage {
	case "score":
		err = models.RejudgeScore(db, id...)
	case "run":
		err = models.RejudgeRun(db, id...)
	case "compile":
		err = models.RejudgeCompile(db, id...)
	default:
		err = httperr.BadRequestf("Invalid rejudge stage: %s", stage)
	}
	if err != nil {
		return httperr.BadRequestf("Cannot rejudge: %v", err)
	}
	return nil
}
