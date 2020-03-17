package contests

import (
	"net/http"

	"git.nkagami.me/natsukagami/kjudge/db"
	"git.nkagami.me/natsukagami/kjudge/models"
	"github.com/labstack/echo/v4"
)

// OverviewCtx is the context for rendering "/contests/:id"
type OverviewCtx struct {
	*ContestCtx

	Problems []*models.ProblemWithTestGroups
	Scores   map[int]*models.ProblemResult
}

// Render renders the template corresponding to the context.
func (o *OverviewCtx) Render(c echo.Context) error {
	return c.Render(http.StatusOK, "contests/overview", o)
}

// Collect a OverviewCtx from the context.
func getOverviewCtx(db db.DBContext, c echo.Context) (*OverviewCtx, error) {
	contest, err := getContestCtx(db, c)
	if err != nil {
		return nil, err
	}
	problems, err := models.CollectTestGroups(db, contest.Problems, false)
	if err != nil {
		return nil, err
	}
	scores, err := models.CollectProblemResults(db, contest.Me.ID, contest.Problems)
	if err != nil {
		return nil, err
	}
	return &OverviewCtx{
		ContestCtx: contest,
		Problems:   problems,
		Scores:     scores,
	}, nil
}

// OverviewGet implements GET "/contests/:id"
func (g *Group) OverviewGet(c echo.Context) error {
	ctx, err := getOverviewCtx(g.db, c)
	if err != nil {
		return err
	}
	return ctx.Render(c)
}
