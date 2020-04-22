package admin

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/natsukagami/kjudge/db"
	"github.com/natsukagami/kjudge/models"
	"github.com/natsukagami/kjudge/models/verify"
	"github.com/natsukagami/kjudge/server/httperr"
	"github.com/pkg/errors"
)

// AnnouncementsCtx is the context for rendering admin/contest_announcements
type AnnouncementsCtx struct {
	Contest       *models.Contest
	Problems      map[int]*models.Problem
	Announcements []*models.Announcement

	Error error
	Form  AnnouncementForm
}

// Render renders the context.
func (a *AnnouncementsCtx) Render(c echo.Context) error {
	status := http.StatusOK
	if a.Error != nil {
		status = http.StatusBadRequest
	}
	return c.Render(status, "admin/contest_announcements", a)
}

// AnnouncementForm is the form for creating a new announcement.
type AnnouncementForm struct {
	Problem int    `form:"problem"`
	Content string `form:"content"`
}

// Bind binds the form into a model.
func (f *AnnouncementForm) Bind(a *models.Announcement) {
	if f.Problem == 0 {
		a.ProblemID = sql.NullInt64{Valid: false}
	} else {
		a.ProblemID = sql.NullInt64{Valid: true, Int64: int64(f.Problem)}
	}
	a.Content = []byte(f.Content)
	a.CreatedAt = time.Now()
}

// get an announcements ctx.
func getAnnouncementsCtx(db db.DBContext, c echo.Context) (*AnnouncementsCtx, error) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return nil, httperr.NotFoundf("Contest not found: %s", idStr)
	}
	contest, err := models.GetContest(db, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, httperr.NotFoundf("Contest not found: %s", idStr)
	} else if err != nil {
		return nil, err
	}
	problemsList, err := models.GetContestProblems(db, contest.ID)
	if err != nil {
		return nil, err
	}
	problems := make(map[int]*models.Problem)
	for _, p := range problemsList {
		problems[p.ID] = p
	}
	announcements, err := models.GetContestAnnouncements(db, contest.ID)
	if err != nil {
		return nil, err
	}

	return &AnnouncementsCtx{
		Contest:       contest,
		Problems:      problems,
		Announcements: announcements,
	}, nil
}

// AnnouncementsGet implements GET /admin/contests/:id/announcements
func (g *Group) AnnouncementsGet(c echo.Context) error {
	ctx, err := getAnnouncementsCtx(g.db, c)
	if err != nil {
		return err
	}
	return ctx.Render(c)
}

// AnnouncementAddPost implements POST /admin/contests/:id/announcements
func (g *Group) AnnouncementAddPost(c echo.Context) error {
	ctx, err := getAnnouncementsCtx(g.db, c)
	if err != nil {
		return err
	}
	if err := c.Bind(&ctx.Form); err != nil {
		return httperr.BindFail(err)
	}
	var ann models.Announcement
	ctx.Form.Bind(&ann)
	ann.ContestID = ctx.Contest.ID
	perform := func() error {
		if ann.ProblemID.Valid {
			if _, ok := ctx.Problems[int(ann.ProblemID.Int64)]; !ok {
				return verify.Errorf("Problem does not belong to the current contest")
			}
		}
		if err := ann.Write(g.db); err != nil {
			return err
		}
		return nil
	}
	if err := perform(); err != nil {
		ctx.Error = err
		return ctx.Render(c)
	}
	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/admin/contests/%d/announcements#list", ctx.Contest.ID))
}
