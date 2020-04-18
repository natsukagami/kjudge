package contests

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/natsukagami/kjudge/db"
	"github.com/natsukagami/kjudge/models"
)

// AnnouncementsCtx is the context for rendering contests/announcements.
type AnnouncementsCtx struct {
	*ContestCtx

	Problems      map[int]*models.Problem
	Announcements []*models.Announcement
}

func (a *AnnouncementsCtx) Render(c echo.Context) error {
	return c.Render(http.StatusOK, "contests/announcements", a)
}

func getAnnouncementsCtx(db db.DBContext, c echo.Context) (*AnnouncementsCtx, error) {
	contest, err := getContestCtx(db, c)
	if err != nil {
		return nil, err
	}
	problems := make(map[int]*models.Problem)
	for _, p := range contest.Problems {
		problems[p.ID] = p
	}
	announcements, err := models.GetContestAnnouncements(db, contest.Contest.ID)
	if err != nil {
		return nil, err
	}
	return &AnnouncementsCtx{
		ContestCtx:    contest,
		Problems:      problems,
		Announcements: announcements,
	}, nil
}

// AnnouncementsGet implements GET /contests/:id/announcements.
func (g *Group) AnnouncementsGet(c echo.Context) error {
	ctx, err := getAnnouncementsCtx(g.db, c)
	if err != nil {
		return err
	}
	return ctx.Render(c)
}
