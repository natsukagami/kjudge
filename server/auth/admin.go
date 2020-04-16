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

// The string that will be used as the admin key.
var adminKey string

// The admin key environment variable.
const adminKeyEnv = "ADMIN_KEY"

// The key name for the admin session.
const adminSessionName = "kjudge_admin"

func init() {
	val := os.Getenv(adminKeyEnv)
	if val != "" {
		if len(val) < 6 {
			log.Fatalf("The admin key should be at least 6 characters long.")
		}
		adminKey = val
		return
	}

	log.Println("ADMIN_KEY variable not set. A random key will be generated and displayed.")
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		log.Fatalf("%+v", errors.WithStack(err))
	}
	adminKey = fmt.Sprintf("%x", b)
	log.Printf("The Admin Panel access key is `%s`\n", adminKey)
}

// MustAdmin is a middleware that ensures admin access.
func MustAdmin(h echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		res, err := AuthenticateAdmin(c)
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
func AuthenticateAdmin(c echo.Context) (bool, error) {
	sess, err := session.Get(adminSessionName, c)
	if err != nil {
		return false, errors.Wrapf(RemoveAdmin(c), "handling err %v", err)
	}
	if sess.IsNew {
		return false, nil
	}
	hashedKey, ok := sess.Values["key"].(string)
	if !ok {
		return false, nil
	}

	return CheckPassword(adminKey, hashedKey)
}

// SaveAdmin saves the admin cookie.
func SaveAdmin(key string, c echo.Context) error {
	if key != adminKey {
		return errors.New("Invalid admin key")
	}
	sess, _ := session.Get(adminSessionName, c)
	sess.Options.MaxAge = 0
	hashedKey, err := PasswordHash(adminKey)
	if err != nil {
		return err
	}
	sess.Values["key"] = string(hashedKey)
	return errors.WithStack(sess.Save(c.Request(), c.Response()))
}

// RemoveAdmin removes the admin cookie session.
func RemoveAdmin(c echo.Context) error {
	sess, _ := session.Get(adminSessionName, c)
	sess.Options.MaxAge = -1
	return errors.WithStack(sess.Save(c.Request(), c.Response()))
}
