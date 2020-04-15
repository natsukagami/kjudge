package admin

import (
	"bytes"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/natsukagami/kjudge/db"
	"github.com/natsukagami/kjudge/models"
	"github.com/natsukagami/kjudge/server/httperr"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

// ScoreboardCtx is the context required to display the scoreboard page
type ScoreboardCtx struct {
	*models.Scoreboard
}

// Show decides whether the scoreboard can be shown. For compability with contests.ScoreboardCtx
func (s *ScoreboardCtx) Show() error {
	return nil
}

// JSONLink returns the link to the JSON scoreboard.
func (s *ScoreboardCtx) JSONLink() string {
	return fmt.Sprintf("/admin/contests/%d/scoreboard/json", s.Contest.ID)
}

// Render renders the scoreboard context
func (s *ScoreboardCtx) Render(c echo.Context, wide bool) error {
	if wide {
		return c.Render(http.StatusOK, "contests/scoreboard_wide", s)
	}
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
	return ctx.Render(c, c.QueryParam("wide") == "true")
}

// ScoreboardJSONGet implements GET /admin/contests/:id/scoreboard/json
func (g *Group) ScoreboardJSONGet(c echo.Context) error {
	ctx, err := getScoreboardCtx(g.db, c)
	if err != nil {
		return err
	}
	return ctx.RenderJSON(c)
}

// ScoreboardCSVGet implements GET /admin/contests/:id/scoreboard/csv
func (g *Group) ScoreboardCSVGet(c echo.Context) error {
	ctx, err := getScoreboardCtx(g.db, c)
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	if c.QueryParam("scores_only") == "true" {
		if err := ctx.CSVScoresOnly(&buf); err != nil {
			return err
		}
	} else {
		if err := ctx.CSV(&buf); err != nil {
			return err
		}
	}
	c.Response().Header().Add("Content-Disposition", `attachment; filename="scoreboard.csv"`)
	return c.Blob(http.StatusOK, "text/csv", buf.Bytes())
}
