// Package admin defines admin routes.
package admin

import (
	"net/http"

	"git.nkagami.me/natsukagami/kjudge/db"
	"github.com/labstack/echo/v4"
)

// Group represents a router Group with handling functions.
type Group struct {
	*echo.Group
	db *db.DB
}

// New creates a new group.
func New(g *echo.Group, db *db.DB) *Group {
	grp := &Group{
		Group: g,
		db:    db,
	}
	g.GET("/", grp.Home)
	return grp
}

// Home renders the home page.
func (g *Group) Home(c echo.Context) error {
	return c.Render(http.StatusOK, "admin/home", nil)
}
