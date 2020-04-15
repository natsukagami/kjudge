package auth

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/natsukagami/kjudge/db"
	"github.com/natsukagami/kjudge/models"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

const sessionName = "kjudge_user"

// MustAuth returns the middleware that redirects to /login if authentication is not found.
func MustAuth(db *db.DB) echo.MiddlewareFunc {
	return func(h echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			u, err := Authenticate(db, c)
			if err != nil {
				return err
			}
			if u == nil {
				// Redirect to login
				return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/user/login?last=%s", url.QueryEscape(c.Request().URL.EscapedPath())))
			}
			return h(c)
		}
	}
}

// Authenticate tries to resolve an authentication from the context.
// Might return a nil user with a nil error.
func Authenticate(db db.DBContext, c echo.Context) (*models.User, error) {
	// Search the cache
	if user, ok := c.Get(sessionName).(*models.User); ok {
		return user, nil
	}

	sess, err := session.Get(sessionName, c)
	if err != nil {
		return nil, Remove(c)
	}
	if sess.IsNew {
		return nil, nil
	}
	username, ok := sess.Values["username"].(string)
	if !ok {
		return nil, Remove(c)
	}

	user, err := models.GetUser(db, username)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, Remove(c)
	} else if err != nil {
		return nil, err
	}

	// Save to cache
	c.Set(sessionName, user)

	return user, nil
}

// Store stores the user as a cookie.
func Store(u *models.User, timeout time.Duration, c echo.Context) error {
	sess, err := session.Get(sessionName, c)
	if err != nil {
		return errors.WithStack(err)
	}
	sess.Values["username"] = u.ID
	if timeout != 0 {
		sess.Options.MaxAge = int(timeout / time.Second)
	}
	return errors.WithStack(sess.Save(c.Request(), c.Response()))
}

// Remove removes the authentication cookie.
func Remove(c echo.Context) error {
	sess, _ := session.Get(sessionName, c)
	sess.Options.MaxAge = -1
	return errors.WithStack(sess.Save(c.Request(), c.Response()))
}
