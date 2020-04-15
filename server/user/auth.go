package user

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/natsukagami/kjudge/db"
	"github.com/natsukagami/kjudge/models"
	"github.com/natsukagami/kjudge/server/auth"
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
	AllowRegistration bool

	Error error
	RegisterForm
}

type LoginForm struct {
	ID       string `form:"id"`
	Password string `form:"password"`
	Remember bool   `form:"remember"`
}

type RegisterForm struct {
	LoginForm
	DisplayName  string `form:"display_name"`
	Organization string `form:"organization"`
}

// Returns a login ctx, but with empty errors and fields.
func getLoginCtx(db db.DBContext, c echo.Context) (*LoginCtx, error) {
	cfg, err := models.GetConfig(db)
	if err != nil {
		return nil, err
	}
	return &LoginCtx{AllowRegistration: cfg.EnableRegistration}, nil
}

// LoginGet implements GET /user/login.
func (g *Group) LoginGet(c echo.Context) error {
	if ok, err := alreadyLoggedIn(g.db, c); err != nil {
		return err
	} else if ok {
		return nil
	}
	ctx, err := getLoginCtx(g.db, c)
	if err != nil {
		return err
	}
	return ctx.Render(c)
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
	ctx, err := getLoginCtx(g.db, c)
	if err != nil {
		return err
	}
	if err := c.Bind(&ctx.LoginForm); err != nil {
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

// RegisterPost implements POST /user/register
func (g *Group) RegisterPost(c echo.Context) error {
	tx, err := g.db.Beginx()
	if err != nil {
		return errors.WithStack(err)
	}
	defer tx.Rollback()

	ctx, err := getLoginCtx(tx, c)
	if err != nil {
		return err
	}

	handle := func() error {
		if !ctx.AllowRegistration {
			return errors.New("Registration is disabled")
		}

		if err := c.Bind(&ctx.RegisterForm); err != nil {
			return errors.WithStack(err)
		}

		// Check if same username exists
		if _, err := models.GetUser(tx, ctx.ID); err == nil {
			return errors.Errorf("Username `%s` already exists", ctx.ID)
		} else if !errors.Is(err, sql.ErrNoRows) {
			return err
		}

		// Hash the password
		password, err := auth.PasswordHash(ctx.Password)
		if err != nil {
			return err
		}

		u := &models.User{
			ID:           ctx.ID,
			Password:     string(password),
			DisplayName:  ctx.DisplayName,
			Organization: ctx.Organization,
		}
		if u.DisplayName == "" {
			u.DisplayName = u.ID
		}
		if err := u.Write(tx); err != nil {
			return err
		}

		// Set the cookie
		if ctx.Remember {
			if err := auth.Store(u, time.Hour*24*30, c); err != nil {
				return err
			}
		} else {
			if err := auth.Store(u, 0, c); err != nil {
				return err
			}
		}
		return errors.WithStack(tx.Commit())
	}
	if err := handle(); err != nil {
		ctx.Error = err
		return ctx.Render(c)
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
