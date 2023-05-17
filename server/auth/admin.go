package auth

import (
	"crypto/rand"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

// AdminAuth hosts the authentication module for admin.
type AdminAuth struct {
	// The string that will be used as the admin key.
	hashedAdminKey []byte
	// Random session key, will change every restart. Guarantees re-login of admin panel.
	sessionKey string
}

// AdminKeyEnv is the admin key environment variable to look for.
const AdminKeyEnv = "ADMIN_KEY"

const (
	// AdminSessionName is the key name for the admin session cookie.
	AdminSessionName     = "kjudge_admin"
	adminSessionKeyField = "key"
)

// NewAdmin creates a new AdminAuth.
func NewAdmin() (*AdminAuth, error) {
	val := os.Getenv(AdminKeyEnv)
	if val != "" {
		if len(val) < 6 {
			return nil, errors.New("The admin key should be at least 6 characters long.")
		}
	} else {
		log.Printf("%s variable not set. A random key will be generated and displayed.", AdminKeyEnv)
		b := make([]byte, 16)
		if _, err := rand.Read(b); err != nil {
			log.Fatalf("%+v", errors.WithStack(err))
		}
		val = fmt.Sprintf("%x", b)
		log.Printf("The Admin Panel access key is `%s`\n", val)
	}
	// store the hashed version of it in memory
	adminKey, err := PasswordHash(val)
	if err != nil {
		return nil, errors.Wrap(err, "error generating hashed key")
	}
	// generate a session key
	sessionKey := make([]byte, 64)
	if _, err := rand.Read(sessionKey); err != nil {
		return nil, errors.Wrap(err, "error generating session key")
	}
	return &AdminAuth{hashedAdminKey: adminKey, sessionKey: fmt.Sprintf("%x", sessionKey)}, nil
}

// MustAdmin is a middleware that ensures admin access.
func (au *AdminAuth) MustAdmin(h echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		res, err := au.AuthenticateAdmin(c)
		if err != nil {
			return nil
		}
		if !res {
			_ = RemoveAdmin(c)
			return c.Redirect(http.StatusTemporaryRedirect, "/admin/login?last="+url.QueryEscape(c.Request().URL.String()))
		}
		return h(c)
	}
}

// AuthenticateAdmin returns whether the context has admin panel access.
func (au *AdminAuth) AuthenticateAdmin(c echo.Context) (bool, error) {
	sess, err := session.Get(AdminSessionName, c)
	if err != nil {
		return false, errors.Wrapf(RemoveAdmin(c), "handling err %v", err)
	}
	if sess.IsNew {
		return false, nil
	}
	key, ok := sess.Values[adminSessionKeyField].(string)
	if !ok {
		return false, nil
	}

	return key == au.sessionKey, nil
}

// SaveAdmin saves the admin cookie.
func (au *AdminAuth) SaveAdmin(key string, c echo.Context) error {
	if ok, err := CheckPassword(key, string(au.hashedAdminKey)); err != nil {
		return err
	} else if !ok {
		return errors.New("Invalid admin key")
	}
	sess, _ := session.Get(AdminSessionName, c)
	sess.Options.MaxAge = 0
	sess.Options.HttpOnly = true
	sess.Options.SameSite = http.SameSiteStrictMode
	sess.Values[adminSessionKeyField] = au.sessionKey
	return errors.WithStack(sess.Save(c.Request(), c.Response()))
}

// RemoveAdmin removes the admin cookie session.
func RemoveAdmin(c echo.Context) error {
	sess, _ := session.Get(AdminSessionName, c)
	sess.Options.MaxAge = -1
	return errors.WithStack(sess.Save(c.Request(), c.Response()))
}
