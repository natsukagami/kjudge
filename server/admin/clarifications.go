package admin

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/natsukagami/kjudge/db"
	"github.com/natsukagami/kjudge/models"
)

// ClarificationsCtx is the context for rendering clarifications.
type ClarificationsCtx struct {
	Clarifications []*models.Clarification
	Problems       map[int]*models.Problem
	Contests       map[int]*models.Contest
}

// Render renders the context.
func (ctx *ClarificationsCtx) Render(c echo.Context) error {
	return c.Render(http.StatusOK, "admin/clarifications", ctx)
}

func getClarificationsCtx(db db.DBContext, c echo.Context) (*ClarificationsCtx, error) {
	clars, err := models.GetAllClarifications(db)
	if err != nil {
		return nil, err
	}
	var problemIDs, contestIDs []int
	for _, c := range clars {
		if c.ProblemID.Valid {
			problemIDs = append(problemIDs, int(c.ProblemID.Int64))
		}
		contestIDs = append(contestIDs, c.ContestID)
	}
	problems, err := models.CollectProblemsByID(db, problemIDs...)
	if err != nil {
		return nil, err
	}
	contests, err := models.CollectContestsByID(db, contestIDs...)
	if err != nil {
		return nil, err
	}
	return &ClarificationsCtx{
		Clarifications: clars,
		Problems:       problems,
		Contests:       contests,
	}, nil
}

// ClarificationsGet implements GET /admin/clarifications
func (g *Group) ClarificationsGet(c echo.Context) error {
	ctx, err := getClarificationsCtx(g.db, c)
	if err != nil {
		return err
	}
	return ctx.Render(c)
}
