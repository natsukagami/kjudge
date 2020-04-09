package admin

import (
	"database/sql"
	"net/http"
	"strconv"

	"git.nkagami.me/natsukagami/kjudge/db"
	"git.nkagami.me/natsukagami/kjudge/models"
	"git.nkagami.me/natsukagami/kjudge/server/httperr"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

// ScoreboardCtx is the context required to display the scoreboard page
type ScoreboardCtx struct {
	*models.Scoreboard
}

// Render renders the scoreboard context
func (s *ScoreboardCtx) Render(c echo.Context) error {
	return c.Render(http.StatusOK, "admin/contest_scoreboard", s)
}

// RenderJSON renders a scoreboard in JSON.
func (s *ScoreboardCtx) RenderJSON(c echo.Context) error {
	return c.JSON(http.StatusOK, s.JSON())
}

// Collect a ScoreboardCtx
func getScoreboardCtx(db db.DBContext, c echo.Context) (*ScoreboardCtx, error) {
	// get contest information
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return nil, httperr.NotFoundf("Contest not found: %s", idStr)
	}
	contest, err := models.GetContest(db, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, httperr.NotFoundf("Contest not found: %d", id)
	} else if err != nil {
		return nil, err
	}

	// get contest's problems
	problems, err := models.GetContestProblems(db, contest.ID)
	if err != nil {
		return nil, err
	}

	scoreboard, err := models.GetScoreboard(db, contest, problems)
	if err != nil {
		return nil, err
	}

	return &ScoreboardCtx{
		Scoreboard: scoreboard,
	}, nil
}

// ScoreboardGet implements GET /admin/contests/:id/scoreboard
func (g *Group) ScoreboardGet(c echo.Context) error {
	ctx, err := getScoreboardCtx(g.db, c)
	if err != nil {
		return err
	}
	return ctx.Render(c)
}

// ScoreboardJSONGet implements GET /admin/contests/:id/scoreboard/json
func (g *Group) ScoreboardJSONGet(c echo.Context) error {
	ctx, err := getScoreboardCtx(g.db, c)
	if err != nil {
		return err
	}
	return ctx.RenderJSON(c)
}
