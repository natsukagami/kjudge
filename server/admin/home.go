package admin

import (
	"log"
	"net/http"

	"github.com/natsukagami/kjudge/db"
	"github.com/natsukagami/kjudge/models"
	"github.com/labstack/echo/v4"
)

// HomeCtx is a context for rendering Home page.
type HomeCtx struct {
	Contests []*models.Contest
	Queue    *models.QueueOverview

	NewVersionMessage string
}

func getHomeCtx(db db.DBContext, c echo.Context) (*HomeCtx, error) {
	contests, err := models.GetContestsUnfinished(db)
	if err != nil {
		return nil, err
	}
	queue, err := models.GetQueueOverview(db)
	if err != nil {
		return nil, err
	}
	message, err := NewVersionMessageGet()
	if err != nil {
		log.Printf("Falied to get kjudge's release version: %+v", err)
		message = ""
	}
	return &HomeCtx{Contests: contests, Queue: queue, NewVersionMessage: message}, nil
}

// Render renders the context.
func (h *HomeCtx) Render(c echo.Context) error {
	return c.Render(http.StatusOK, "admin/home", h)
}

// Home renders the home page.
func (g *Group) Home(c echo.Context) error {
	ctx, err := getHomeCtx(g.db, c)
	if err != nil {
		return err
	}
	return ctx.Render(c)
}
