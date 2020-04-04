package admin

import (
	"net/http"
	"time"

	"git.nkagami.me/natsukagami/kjudge/models"
	"git.nkagami.me/natsukagami/kjudge/server/httperr"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

// Timestamp is a wrapped time.Time for form-parsing.
type Timestamp time.Time

const timeFormat = "2006-01-02T15:04"

// String formats the timestamp as RFC3339.
func (t Timestamp) String() string {
	return time.Time(t).Format(timeFormat)
}

// UnmarshalParam implement echo's Bind.
func (t *Timestamp) UnmarshalParam(src string) error {
	ts, err := time.Parse(timeFormat, src)
	*t = Timestamp(ts)
	return errors.WithStack(err)
}

// ContestsCtx is a context for rendering contests.
type ContestsCtx struct {
	Contests []*models.Contest

	FormError error
	Form      ContestForm
}

// ContestForm is a form for uploading a new contest.
type ContestForm struct {
	Name        string             `form:"name"`
	StartTime   Timestamp          `form:"start_time"`
	EndTime     Timestamp          `form:"end_time"`
	ContestType models.ContestType `form:"contest_type"`
}

// ContestToForm creates a form with the initial values of the contest.
func ContestToForm(c *models.Contest) *ContestForm {
	return &ContestForm{
		Name:        c.Name,
		StartTime:   Timestamp(c.StartTime),
		EndTime:     Timestamp(c.EndTime),
		ContestType: c.ContestType,
	}
}

// Bind binds the form's content to the contest's.
func (f *ContestForm) Bind(c *models.Contest) {
	c.Name = f.Name
	c.StartTime = time.Time(f.StartTime)
	c.EndTime = time.Time(f.EndTime)
	c.ContestType = f.ContestType
}

// ContestsGet handles GET /admin/contests
func (g *Group) ContestsGet(c echo.Context) error {
	contests, err := models.GetContests(g.db)
	if err != nil {
		return err
	}
	return c.Render(http.StatusOK, "admin/contests", &ContestsCtx{Contests: contests})
}

func (g *Group) contestsWithFormError(formError error, form ContestForm, c echo.Context) error {
	contests, err := models.GetContests(g.db)
	if err != nil {
		return err
	}
	return c.Render(http.StatusBadRequest, "admin/contests", &ContestsCtx{Contests: contests, FormError: formError, Form: form})
}

// ContestsPost handles POST /admin/contests.
// TODO: redirect to /admin/contests/[id]
func (g *Group) ContestsPost(c echo.Context) error {
	var form ContestForm
	if err := c.Bind(&form); err != nil {
		return httperr.BindFail(err)
	}
	var contest models.Contest
	form.Bind(&contest)
	if err := contest.Write(g.db); err != nil {
		return g.contestsWithFormError(err, form, c)
	}
	return g.ContestsGet(c)
}
