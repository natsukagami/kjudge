// User implements a group that handle /user requests.
package user

import (
	"net/http"

	"git.nkagami.me/natsukagami/kjudge/db"
	"git.nkagami.me/natsukagami/kjudge/server/auth"
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

	authed := g.Group("", auth.MustAuth(db))
	authed.GET("", grp.HomeGet)
	authed.GET("/logout", grp.LogoutPost)
	authed.POST("/logout", grp.LogoutPost)

	return grp, nil
}

// HomeGet implements "/user".
func (g *Group) HomeGet(c echo.Context) error {
	u, err := Me(g.db, c)
	if err != nil {
		return err
	}
	return c.Render(http.StatusOK, "user/home", u)
}
