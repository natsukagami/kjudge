package contests

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/natsukagami/kjudge/db"
	"github.com/natsukagami/kjudge/models"
)

// MessagesCtx is the context for rendering contests/messages.
type MessagesCtx struct {
	*ContestCtx

	Problems map[int]*models.Problem
	Messages []Message
}

type Message struct {
	*models.Announcement
	*models.Clarification
}

func (a *MessagesCtx) Render(c echo.Context) error {
	return c.Render(http.StatusOK, "contests/messages", a)
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
	return &MessagesCtx{
		ContestCtx: contest,
		Problems:   problems,
		Messages:   messages,
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
