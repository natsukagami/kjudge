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
	// Contest List
	g.GET("/contests", grp.ContestsGet)
	g.POST("/contests", grp.ContestsPost)
	// Contest Management
	g.GET("/contests/:id", grp.ContestGet)
	g.POST("/contests/:id", grp.ContestEdit)
	g.POST("/contests/:id/delete", grp.ContestDelete)
	g.POST("/contests/:id/add_problem", grp.ContestAddProblem)
	// Problem Management
	g.GET("/problems/:id", grp.ProblemGet)
	g.POST("/problems/:id", grp.ProblemEdit)
	g.POST("/problems/:id/add_test_group", grp.ProblemAddTestGroup)
	// Test groups
	g.POST("/test_groups/:id/upload_single", grp.TestGroupUploadSingle)
	g.POST("/test_groups/:id/upload_multiple", grp.TestGroupUploadMultiple)
	return grp
}
