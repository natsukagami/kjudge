package admin

import (
	"net/http"
	"time"

	"git.nkagami.me/natsukagami/kjudge/models"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

// Timestamp is a wrapped time.Time for form-parsing.
type Timestamp time.Time

// String formats the timestamp as RFC3339.
func (t Timestamp) String() string {
	return time.Time(t).Format(time.RFC3339)
}

// UnmarshalParam implement echo's Bind.
func (t *Timestamp) UnmarshalParam(src string) error {
	ts, err := time.Parse(time.RFC3339, src)
	*t = Timestamp(ts)
	return err
}

// ContestsCtx is a context for rendering contests.
type ContestsCtx struct {
	Contests []*models.Contest

	FormError error
	Form      NewContestForm
}

// NewContestForm is a form for uploading a new contest.
type NewContestForm struct {
	Name      string             `form:"name"`
	StartTime Timestamp          `form:"start_time"`
	EndTime   Timestamp          `form:"end_time"`
	Type      models.ContestType `form:"type"`
}

// ContestsGet handles GET /admin/contests
func (g *Group) ContestsGet(c echo.Context) error {
	contests, err := models.GetContests(g.db)
	if err != nil {
		return err
	}
	return c.Render(http.StatusOK, "admin/contests", &ContestsCtx{Contests: contests})
}

func (g *Group) contestsWithFormError(formError error, form NewContestForm, c echo.Context) error {
	contests, err := models.GetContests(g.db)
	if err != nil {
		return err
	}
	return c.Render(http.StatusBadRequest, "admin/contests", &ContestsCtx{Contests: contests, FormError: formError, Form: form})
}

// ContestsPost handles POST /admin/contests.
// TODO: redirect to /admin/contests/[id]
func (g *Group) ContestsPost(c echo.Context) error {
	var form NewContestForm
	if err := c.Bind(&form); err != nil {
		return errors.WithStack(err)
	}
	contest := &models.Contest{
		Name:        form.Name,
		StartTime:   time.Time(form.StartTime),
		EndTime:     time.Time(form.EndTime),
		ContestType: form.Type,
	}
	if err := contest.Write(g.db); err != nil {
		return g.contestsWithFormError(err, form, c)
	}
	return g.ContestsGet(c)
}
