// Package integration contains integration tests.
package integration

import (
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/natsukagami/kjudge/server/auth"
	"github.com/natsukagami/kjudge/test"
)

func TestLogin(t *testing.T) {
	svr := test.NewServer(t)

	run := func(username, password string) *http.Response {
		form := url.Values{}
		form.Set("id", username)
		form.Set("password", password)
		return svr.Serve(svr.PostForm(t, "/user/login", form))
	}

	t.Run("success", func(t *testing.T) {
		res := run("misaka", "misaka")
		if res.StatusCode >= 400 {
			t.Errorf("Expected OK got %d", res.StatusCode)
		}
		cookies := res.Cookies()
		if len(cookies) != 1 {
			t.Errorf("Expected one cookie, got %#v", cookies)
		}
		cookie := cookies[0]
		if cookie.Name != auth.SessionName {
			t.Errorf("Expected %s cookie name, got %s", auth.SessionName, cookie.Name)
		}
		if !cookie.HttpOnly {
			t.Errorf("Cookie is not HttpOnly")
		}
		if cookie.Expires.Before(time.Now()) {
			t.Errorf("Cookie expired at %v", cookie.Expires)
		}
	})

	t.Run("wrong password", func(t *testing.T) {
		res := run("misaka", "misaka2")
		if res.StatusCode < 400 {
			t.Errorf("Expected err got %d", res.StatusCode)
		}
		cookies := res.Cookies()
		if len(cookies) != 0 {
			t.Errorf("Expected no cookie, got %#v", cookies)
		}
	})
}
