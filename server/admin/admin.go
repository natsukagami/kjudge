// Package admin defines admin routes.
package admin

import (
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
	g.GET("", grp.Home)
	g.GET("/contests", grp.ContestsGet)
	g.POST("/contests", grp.ContestsPost)
	g.GET("/contests/:id", grp.ContestGet)
	g.POST("/contests/:id", grp.ContestEdit)
	g.POST("/contests/:id/delete", grp.ContestDelete)
	return grp
}
