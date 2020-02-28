package admin

import (
	"net/http"

	"git.nkagami.me/natsukagami/kjudge/models"
	"github.com/labstack/echo/v4"
)

// HomeCtx is a context for rendering Home page.
type HomeCtx struct {
	Contests []*models.Contest
	Queue    *models.QueueOverview
}

// Home renders the home page.
func (g *Group) Home(c echo.Context) error {
	contests, err := models.GetContestsUnfinished(g.db)
	if err != nil {
		return err
	}
	queue, err := models.GetQueueOverview(g.db)
	if err != nil {
		return err
	}
	return c.Render(http.StatusOK, "admin/home", &HomeCtx{Contests: contests, Queue: queue})
}
