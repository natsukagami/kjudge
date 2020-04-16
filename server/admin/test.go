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

func getTest(db db.DBContext, c echo.Context) (*models.Test, error) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return nil, httperr.BadRequestf("Test not found: %v", idStr)
	}
	test, err := models.GetTest(db, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, httperr.BadRequestf("Test not found: %v", idStr)
	} else if err != nil {
		return nil, err
	}
	return test, nil
}

// TestInput implements GET /admin/tests/:id/input
func (g *Group) TestInput(c echo.Context) error {
	test, err := getTest(g.db, c)
	if err != nil {
		return err
	}
	return c.Blob(http.StatusOK, "text/plain", test.Input)
}

// TestOutput implements GET /admin/tests/:id/output
func (g *Group) TestOutput(c echo.Context) error {
	test, err := getTest(g.db, c)
	if err != nil {
		return err
	}
	return c.Blob(http.StatusOK, "text/plain", test.Output)
}

// TestDelete implements GET /admin/tests/:id/delete
func (g *Group) TestDelete(c echo.Context) error {
	test, err := getTest(g.db, c)
	if err != nil {
		return err
	}
	tg, err := models.GetTestGroup(g.db, test.TestGroupID)
	if err != nil {
		return err
	}
	if err := test.Delete(g.db); err != nil {
		return err
	}
	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/admin/problems/%d", tg.ProblemID))
}
