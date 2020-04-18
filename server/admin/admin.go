// Package admin defines admin routes.
package admin

import (
	"github.com/labstack/echo/v4"
	"github.com/natsukagami/kjudge/db"
	"github.com/natsukagami/kjudge/server/auth"
)

// Group represents a router Group with handling functions.
type Group struct {
	*echo.Group
	db *db.DB
}

// New creates a new group.
func New(db *db.DB, unauthed *echo.Group) (*Group, error) {
	grp := &Group{
		Group: unauthed,
		db:    db,
	}
	// Authentication
	unauthed.GET("/login", grp.LoginGet)
	unauthed.POST("/login", grp.LoginPost)

	g := unauthed.Group("", auth.MustAdmin)
	g.GET("/logout", grp.LogoutPost)
	g.GET("", grp.Home)
	// Contest List
	g.GET("/contests", grp.ContestsGet)
	g.POST("/contests", grp.ContestsPost)
	// Contest Scoreboard
	g.GET("/contests/:id/scoreboard", grp.ScoreboardGet)
	g.GET("/contests/:id/scoreboard/json", grp.ScoreboardJSONGet)
	g.GET("/contests/:id/scoreboard/csv", grp.ScoreboardCSVGet)
	// Contest Management
	g.GET("/contests/:id", grp.ContestGet)
	g.GET("/contests/:id/submissions", grp.ContestSubmissionsGet)
	g.POST("/contests/:id", grp.ContestEdit)
	g.POST("/contests/:id/delete", grp.ContestDelete)
	g.POST("/contests/:id/add_problem", grp.ContestAddProblem)
	g.POST("/contests/:id/rejudge", grp.ContestRejudgePost)
	// Contest Announcements
	g.GET("/contests/:id/announcements", grp.AnnouncementsGet)
	g.POST("/contests/:id/announcements", grp.AnnouncementAddPost)
	// Problem Management
	g.GET("/problems/:id", grp.ProblemGet)
	g.GET("/problems/:id/submissions", grp.ProblemSubmissionsGet)
	g.POST("/problems/:id", grp.ProblemEdit)
	g.POST("/problems/:id/add_test_group", grp.ProblemAddTestGroup)
	g.POST("/problems/:id/add_file", grp.ProblemAddFile)
	g.POST("/problems/:id/delete", grp.ProblemDelete)
	g.POST("/problems/:id/rejudge", grp.ProblemRejudgePost)
	// Test groups
	g.GET("/test_groups/:id", grp.TestGroupGet)
	g.POST("/test_groups/:id/upload_single", grp.TestGroupUploadSingle)
	g.POST("/test_groups/:id/upload_multiple", grp.TestGroupUploadMultiple)
	g.POST("/test_groups/:id", grp.TestGroupEdit)
	g.POST("/test_groups/:id/delete", grp.TestGroupDelete)
	g.POST("/test_groups/:id/rejudge", grp.TestGroupRejudgePost)
	// Test
	g.GET("/tests/:id/input", grp.TestInput)
	g.GET("/tests/:id/output", grp.TestOutput)
	g.POST("/tests/:id/delete", grp.TestDelete)
	// File
	g.GET("/files/:id", grp.FileGet)
	g.POST("/files/:id/delete", grp.FileDelete)
	g.POST("/files/:id/compile", grp.FileCompile)
	// Users
	g.GET("/users", grp.UsersGet)
	g.POST("/users", grp.UsersAdd)
	g.GET("/users/:id", grp.UserGet)
	g.POST("/users/:id", grp.UserEdit)
	g.POST("/users/:id/delete", grp.UserDelete)
	g.POST("/config/toggle", grp.ConfigTogglePost)
	// Submissions
	g.GET("/submissions", grp.SubmissionsGet)
	g.GET("/submissions/:id", grp.SubmissionGet)
	g.GET("/submissions/:id/verdict", grp.SubmissionVerdictGet)
	g.GET("/submissions/:id/binary", grp.SubmissionBinaryGet)
	g.POST("/rejudge", grp.RejudgePost)
	// Jobs
	g.GET("/jobs", grp.JobsGet)
	// Batch users
	g.GET("/batch_users/empty", grp.BatchUsersEmptyGet)
	g.GET("/batch_users/generate", grp.BatchUsersGenerateGet)
	g.POST("/batch_users", grp.BatchUsersPost)

	return grp, nil
}
