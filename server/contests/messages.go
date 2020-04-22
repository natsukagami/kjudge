package contests

import (
	"database/sql"
	"errors"
	"net/http"
	"sort"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/natsukagami/kjudge/db"
	"github.com/natsukagami/kjudge/models"
	"github.com/natsukagami/kjudge/server/httperr"
)

// MessagesCtx is the context for rendering contests/messages.
type MessagesCtx struct {
	*ContestCtx

	ProblemsMap map[int]*models.Problem
	Messages    []Message

	FormError error
	Form      ClarificationForm
}

// ClarificationForm is a form for sending clarifications.
type ClarificationForm struct {
	Problem int64
	Content string
}

// Bind binds the form into a Clarification.
func (f *ClarificationForm) Bind(c *models.Clarification) {
	if f.Problem == 0 {
		c.ProblemID = sql.NullInt64{Valid: false}
	} else {
		c.ProblemID = sql.NullInt64{Valid: true, Int64: f.Problem}
	}
	c.Content = []byte(f.Content)
	c.UpdatedAt = time.Now()
}

// Message is either an Announcement or a Clarification.
type Message struct {
	*models.Announcement
	*models.Clarification
}

// UpdatedAt returns the last updated time of the Message.
func (m Message) UpdatedAt() time.Time {
	if m.Announcement != nil {
		return m.Announcement.CreatedAt
	}
	return m.Clarification.UpdatedAt
}

func (a *MessagesCtx) Render(c echo.Context) error {
	status := http.StatusOK
	if a.FormError != nil {
		status = http.StatusBadRequest
	}
	return c.Render(status, "contests/messages", a)
}

func getMessagesCtx(db db.DBContext, c echo.Context) (*MessagesCtx, error) {
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
	clars, err := models.GetContestUserClarifications(db, contest.Contest.ID, contest.Me.ID)
	if err != nil {
		return nil, err
	}
	var messages []Message
	for _, a := range announcements {
		messages = append(messages, Message{Announcement: a})
	}
	for _, a := range clars {
		messages = append(messages, Message{Clarification: a})
	}
	sort.Slice(messages, func(i, j int) bool { return messages[i].UpdatedAt().After(messages[j].UpdatedAt()) })
	return &MessagesCtx{
		ContestCtx:  contest,
		ProblemsMap: problems,
		Messages:    messages,
	}, nil
}

// MessagesGet implements GET /contests/:id/messages.
func (g *Group) MessagesGet(c echo.Context) error {
	ctx, err := getMessagesCtx(g.db, c)
	if err != nil {
		return err
	}
	return ctx.Render(c)
}

// SendClarificationPost implements POST /contests/:id/messages.
func (g *Group) SendClarificationPost(c echo.Context) error {
	ctx, err := getMessagesCtx(g.db, c)
	if err != nil {
		return err
	}
	if ctx.Contest.EndTime.Before(time.Now()) {
		return httperr.BadRequestf("Cannot send a clarification after the contest ends")
	}
	if err := c.Bind(&ctx.Form); err != nil {
		return httperr.BindFail(err)
	}
	var clar models.Clarification
	ctx.Form.Bind(&clar)
	clar.UserID = ctx.Me.ID
	clar.ContestID = ctx.Contest.ID
	if clar.ProblemID.Valid {
		if _, ok := ctx.ProblemsMap[int(clar.ProblemID.Int64)]; !ok {
			ctx.FormError = errors.New("Problem is not part of contest")
			return ctx.Render(c)
		}
	}

	if err := clar.Write(g.db); err != nil {
		ctx.FormError = err
		return ctx.Render(c)
	}
	return c.Redirect(http.StatusSeeOther, ctx.Contest.Link()+"/messages")
}

// MessagesUnreadGet returns the number of unread messages.
// Implements GET /contests/:id/messages/unread
func (g *Group) MessagesUnreadGet(c echo.Context) error {
	ctx, err := getContestCtx(g.db, c)
	if err != nil {
		return err
	}
	var last struct {
		LastAnnouncement  int `query:"last_announcement"`
		LastClarification int `query:"last_clarification"`
	}
	if err := c.Bind(&last); err != nil {
		return httperr.BindFail(err)
	}
	anns, err := models.GetUnreadAnnouncements(g.db, ctx.Contest.ID, last.LastAnnouncement)
	if err != nil {
		return err
	}
	clars, err := models.GetUnreadClarifications(g.db, ctx.Contest.ID, ctx.Me.ID, last.LastClarification)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, len(anns)+len(clars))
}
