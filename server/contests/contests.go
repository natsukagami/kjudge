package contests

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"git.nkagami.me/natsukagami/kjudge/db"
	"git.nkagami.me/natsukagami/kjudge/models"
	"git.nkagami.me/natsukagami/kjudge/server/auth"
	"git.nkagami.me/natsukagami/kjudge/server/user"
	"github.com/labstack/echo/v4"
)

// ContestsCtx is a context for rendering a list of all contests
type ContestsCtx struct {
	*user.AuthCtx

	ActiveContests []*models.Contest
	PastContests   []*models.Contest
}

func getContestsCtx(db db.DBContext, c echo.Context) (*ContestsCtx, error) {
	activeContests, err := models.GetContestsUnfinished(db)
	if err != nil {
		return nil, err
	}
	pastContests, err := models.GetContestsFinished(db)
	if err != nil {
		return nil, err
	}
	me, err := user.Me(db, c)
	if err != nil {
		return nil, err
	}
	return &ContestsCtx{ActiveContests: activeContests, PastContests: pastContests, AuthCtx: me}, nil
}

// Group is the /contests handling group
type Group struct {
	group *echo.Group
	db    *db.DB
}

// New creates a new group
func New(db *db.DB, g *echo.Group) (*Group, error) {
	grp := &Group{
		group: g,
		db:    db,
	}

	g.GET("", grp.ContestsGet)
	authed := g.Group("/", auth.MustAuth(db))
	authed.GET(":id", grp.OverviewGet)
	authed.GET(":id/problems/:problem", grp.ProblemGet)
	authed.GET(":id/problems/:problem/files/:file", grp.FileGet)
	authed.POST(":id/problems/:problem/submit", grp.SubmitPost)

	return grp, nil
}

// Render a table of all contests
func (g *Group) ContestsGet(c echo.Context) error {
	ctx, err := getContestsCtx(g.db, c)
	if err != nil {
		return err
	}
	return c.Render(http.StatusOK, "contests/home", ctx)

}

// Render the nearest ongoing contest or a table of all contests
func (g *Group) ConestsGetNearestOngoingContest(c echo.Context) error {
	contest, err := models.GetNearestOngoingContest(g.db)
	if errors.Is(err, sql.ErrNoRows) {
		return c.Redirect(http.StatusSeeOther, "/contests")
	} else if err != nil {
		return err
	}

	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/contests/%d", contest.ID))
}
