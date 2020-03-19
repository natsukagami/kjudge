package admin

import (
	"net/http"

	"git.nkagami.me/natsukagami/kjudge/server/auth"
	"github.com/labstack/echo/v4"
)

// LoginCtx is the context for rendering the login page.
type LoginCtx struct {
	Error error
}

// Render renders the context.
func (l *LoginCtx) Render(c echo.Context) error {
	status := http.StatusOK
	if l.Error != nil {
		status = http.StatusBadRequest
	}
	return c.Render(status, "admin/login", l)
}

// LoginGet implements GET /admin/login.
func (g *Group) LoginGet(c echo.Context) error {
	if ok, err := alreadyLoggedIn(c); err != nil {
		return err
	} else if ok {
		return nil
	}

	return (&LoginCtx{}).Render(c)
}

// LoginPost implements POST /admin/login.
func (g *Group) LoginPost(c echo.Context) error {
	if ok, err := alreadyLoggedIn(c); err != nil {
		return err
	} else if ok {
		return nil
	}

	key := c.FormValue("key")
	if err := auth.SaveAdmin(key, c); err != nil {
		return (&LoginCtx{Error: err}).Render(c)
	}
	last := c.QueryParam("last")
	if last == "" {
		last = "/admin"
	}
	return c.Redirect(http.StatusSeeOther, last)
}

// LogoutPost implements GET/POST /admin/logout.
func (g *Group) LogoutPost(c echo.Context) error {
	if err := auth.RemoveAdmin(c); err != nil {
		return err
	}
	last := c.QueryParam("last")
	if last == "" {
		last = "/admin"
	}
	return c.Redirect(http.StatusSeeOther, last)
}

// Redirect if already logged in.
func alreadyLoggedIn(c echo.Context) (bool, error) {
	ok, err := auth.AuthenticateAdmin(c)
	if err != nil {
		return false, err
	}
	// Just redirect if the user already logged in.
	if ok {
		last := c.QueryParam("last")
		if last == "" {
			last = "/"
		}
		return true, c.Redirect(http.StatusSeeOther, last)
	}
	return false, nil
}
