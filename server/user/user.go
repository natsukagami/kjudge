// User implements a group that handle /user requests.
package user

import (
	"github.com/natsukagami/kjudge/db"
	"github.com/natsukagami/kjudge/server/auth"
	"github.com/labstack/echo/v4"
)

// Group is the /user handling group.
type Group struct {
	group *echo.Group
	db    *db.DB
}

// New creates a new Group.
func New(db *db.DB, g *echo.Group) (*Group, error) {
	grp := &Group{
		group: g,
		db:    db,
	}

	g.GET("/login", grp.LoginGet)
	g.POST("/login", grp.LoginPost)
	g.POST("/register", grp.RegisterPost)

	authed := g.Group("", auth.MustAuth(db))
	authed.GET("", grp.HomeGet)
	authed.POST("/change_password", grp.ChangePassword)
	authed.POST("/customize", grp.CustomizePost)
	authed.GET("/logout", grp.LogoutPost)
	authed.POST("/logout", grp.LogoutPost)

	return grp, nil
}
