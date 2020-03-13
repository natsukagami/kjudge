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
	g.GET("/contests/:id/submissions", grp.ContestSubmissionsGet)
	g.POST("/contests/:id", grp.ContestEdit)
	g.POST("/contests/:id/delete", grp.ContestDelete)
	g.POST("/contests/:id/add_problem", grp.ContestAddProblem)
	// Problem Management
	g.GET("/problems/:id", grp.ProblemGet)
	g.GET("/problems/:id/submissions", grp.ProblemSubmissionsGet)
	g.POST("/problems/:id", grp.ProblemEdit)
	g.POST("/problems/:id/add_test_group", grp.ProblemAddTestGroup)
	g.POST("/problems/:id/add_file", grp.ProblemAddFile)
	g.POST("/problems/:id/delete", grp.ProblemDelete)
	// Test groups
	g.POST("/test_groups/:id/upload_single", grp.TestGroupUploadSingle)
	g.POST("/test_groups/:id/upload_multiple", grp.TestGroupUploadMultiple)
	g.POST("/test_groups/:id", grp.TestGroupEdit)
	g.POST("/test_groups/:id/delete", grp.TestGroupDelete)
	// Test
	g.GET("/tests/:id/input", grp.TestInput)
	g.GET("/tests/:id/output", grp.TestOutput)
	g.POST("/tests/:id/delete", grp.TestDelete)
	// File
	g.GET("/files/:id", grp.FileGet)
	g.POST("/files/:id/delete", grp.FileDelete)
	// Users
	g.GET("/users", grp.UsersGet)
	g.POST("/users", grp.UsersAdd)
	g.GET("/users/:id", grp.UserGet)
	g.POST("/users/:id", grp.UserEdit)
	g.POST("/users/:id/delete", grp.UserDelete)
	g.POST("/config/toggle_enable_registration", grp.ToggleEnableRegistration)
	// Submissions
	g.GET("/submissions", grp.SubmissionsGet)
	return grp
}
