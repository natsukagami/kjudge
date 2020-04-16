package contests

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/natsukagami/kjudge/db"
	"github.com/natsukagami/kjudge/models"
	"github.com/natsukagami/kjudge/server/auth"
	"github.com/natsukagami/kjudge/server/user"
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
	g.GET("/:id/scoreboard", grp.ScoreboardGet)
	g.GET("/:id/scoreboard/json", grp.ScoreboardJSONGet)
	g.GET("/:id/scoreboard/csv", grp.ScoreboardCSVGet)
	authed := g.Group("/", auth.MustAuth(db))
	authed.GET(":id", grp.OverviewGet)
	authed.GET(":id/problems/:problem", grp.ProblemGet)
	authed.GET(":id/problems/:problem/files/:file", grp.FileGet)
	authed.POST(":id/problems/:problem/submit", grp.SubmitPost)
	authed.GET(":id/submissions/:submission", grp.SubmissionGet)
	authed.GET(":id/submissions/:submission/download", grp.SubmissionDownload)
	authed.GET(":id/submissions/:submission/verdict", grp.SubmissionVerdictGet)

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

	user, err := user.Me(g.db, c)
	if err != nil {
		return err
	}
	if user.Me == nil {
		// redirect to contests/ if scoreboard view status is not public
		if contest.ScoreboardViewStatus != models.ScoreboardViewStatusPublic {
			return c.Redirect(http.StatusSeeOther, "/contests")
		} else {
			return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/contests/%d/scoreboard", contest.ID))
		}
	}

	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/contests/%d", contest.ID))
}
