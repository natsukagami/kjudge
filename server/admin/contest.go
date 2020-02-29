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

func (g *Group) getContest(c echo.Context) (*models.Contest, error) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return nil, echo.ErrNotFound
	}
	contest, err := models.GetContest(g.db, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, echo.ErrNotFound
	} else if err != nil {
		return nil, err
	}
	return contest, nil
}

// ContestCtx is the context for rendering admin/contest.
type ContestCtx struct {
	*models.Contest

	FormError error
	Form      *ContestForm
}

// ContestGet implements GET /admin/contest/:id
func (g *Group) ContestGet(c echo.Context) error {
	contest, err := g.getContest(c)
	if err != nil {
		return err
	}
	ctx := &ContestCtx{Contest: contest, Form: ContestToForm(contest)}
	return g.contestGetRender(ctx, c)
}

func (g *Group) contestGetRender(ctx *ContestCtx, c echo.Context) error {
	code := http.StatusOK
	if ctx.FormError != nil {
		code = http.StatusBadRequest
	}
	return c.Render(code, "admin/contest", ctx)
}

// ContestDelete implement POST /admin/contest/:id/delete
func (g *Group) ContestDelete(c echo.Context) error {
	contest, err := g.getContest(c)
	if err != nil {
		return err
	}
	if err := contest.Delete(g.db); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.Redirect(http.StatusSeeOther, "/admin/contests")
}

// ContestEdit implements POST /admin/contest/:id
func (g *Group) ContestEdit(c echo.Context) error {
	contest, err := g.getContest(c)
	if err != nil {
		return err
	}
	original := *contest
	var form ContestForm
	if err := c.Bind(&form); err != nil {
		return err
	}
	form.Bind(contest)
	if err := contest.Write(g.db); err != nil {
		return g.contestGetRender(&ContestCtx{Contest: &original, Form: &form, FormError: err}, c)
	}
	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/admin/contests/%d", contest.ID))
}
