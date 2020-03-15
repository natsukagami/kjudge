package user

import (
	"database/sql"
	"net/http"
	"time"

	"git.nkagami.me/natsukagami/kjudge/db"
	"git.nkagami.me/natsukagami/kjudge/models"
	"git.nkagami.me/natsukagami/kjudge/server/auth"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

// AuthCtx is a context that holds the "me" user.
type AuthCtx struct {
	Me *models.User
}

// GetMe implements template.Me
func (a *AuthCtx) GetMe() *models.User {
	return a.Me
}

// Me returns the "me" user context.
func Me(db db.DBContext, c echo.Context) (*AuthCtx, error) {
	u, err := auth.Authenticate(db, c)
	if err != nil {
		return nil, err
	}
	return &AuthCtx{Me: u}, nil
}

// LoginCtx represents the context to render logins.
type LoginCtx struct {
	Error error

	ID       string `form:"id"`
	Password string `form:"password"`
	Remember bool   `form:"remember"`
}

// LoginGet implements GET /user/login.
func (g *Group) LoginGet(c echo.Context) error {
	if ok, err := alreadyLoggedIn(g.db, c); err != nil {
		return err
	} else if ok {
		return nil
	}
	return (&LoginCtx{}).Render(c)
}

// Render performs rendering for the context.
func (l *LoginCtx) Render(c echo.Context) error {
	status := http.StatusOK
	if l.Error != nil {
		status = http.StatusBadRequest
	}
	return c.Render(status, "user/login", l)
}

// LoginPost implements POST /user/login.
func (g *Group) LoginPost(c echo.Context) error {
	if ok, err := alreadyLoggedIn(g.db, c); err != nil {
		return err
	} else if ok {
		return nil
	}
	var ctx LoginCtx
	if err := c.Bind(&ctx); err != nil {
		return err
	}
	// get an user with corresponding username
	u, err := models.GetUser(g.db, ctx.ID)
	if errors.Is(err, sql.ErrNoRows) {
		ctx.Error = errors.Errorf("user `%s` does not exist", ctx.ID)
		return ctx.Render(c)
	} else if err != nil {
		return err
	}
	if ok, err := auth.CheckPassword(ctx.Password, u.Password); err != nil {
		return err
	} else if !ok {
		ctx.Error = errors.New("invalid password")
		return ctx.Render(c)
	}

	// Store a cookie
	if ctx.Remember {
		if err := auth.Store(u, time.Hour*24*30, c); err != nil {
			return err
		}
	} else {
		if err := auth.Store(u, 0, c); err != nil {
			return err
		}
	}

	// Perform redirect
	last := c.QueryParam("last")
	if last == "" {
		last = "/"
	}
	return c.Redirect(http.StatusSeeOther, last)
}

// LogoutPost implements GET/POST /user/logout.
func (g *Group) LogoutPost(c echo.Context) error {
	if err := auth.Remove(c); err != nil {
		return err
	}
	last := c.QueryParam("last")
	if last == "" {
		last = "/"
	}
	return c.Redirect(http.StatusSeeOther, last)
}

// Redirect if already logged in.
func alreadyLoggedIn(db db.DBContext, c echo.Context) (bool, error) {
	u, err := auth.Authenticate(db, c)
	if err != nil {
		return false, err
	}
	// Just redirect if the user already logged in.
	if u != nil {
		last := c.QueryParam("last")
		if last == "" {
			last = "/"
		}
		return true, c.Redirect(http.StatusSeeOther, last)
	}
	return false, nil
}
