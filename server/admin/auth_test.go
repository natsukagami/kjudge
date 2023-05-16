package admin_test

import (
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/natsukagami/kjudge/server/auth"
	"github.com/natsukagami/kjudge/test"
)

func TestAdminLogin(t *testing.T) {
	ts := test.NewServer(t)

	run := func(key string) *http.Response {
		form := url.Values{}
		form.Set("key", key)

		return ts.Serve(ts.PostForm(t, "/admin/login", form))
	}

	t.Run("success", func(t *testing.T) {
		resp := run(os.Getenv(auth.AdminKeyEnv))

		if resp.StatusCode >= 400 {
			t.Errorf("Expected OK got %d", resp.StatusCode)
		}
		cookies := resp.Cookies()
		if len(cookies) != 1 {
			t.Errorf("Expected one cookie got %#v", cookies)
		}
		cookie := cookies[0]
		if cookie.Name != auth.AdminSessionName {
			t.Errorf("Expected cookie name `%s` got `%s`", auth.AdminSessionName, cookie.Name)
		}
		if !cookie.HttpOnly {
			t.Error("Cookie is not Http-Only")
		}
		if cookie.SameSite != http.SameSiteStrictMode {
			t.Error("Cookie is not strict same-site")
		}
	})

	t.Run("fail", func(t *testing.T) {
		resp := run("ababababababa")

		if resp.StatusCode < 400 {
			t.Errorf("Expected error got %d", resp.StatusCode)
		}
		cookies := resp.Cookies()
		if len(cookies) != 0 {
			t.Errorf("Expected no cookie got %#v", cookies)
		}
	})
}

func TestAdminLoginCheck(t *testing.T) {
	ts := test.NewServer(t)
	withAdmin := ts.WithAdmin(t)

	t.Run("with admin cookie", func(t *testing.T) {
		resp := ts.Serve(ts.Get(t, "/admin", nil), withAdmin)
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected OK got %d", resp.StatusCode)
		}
	})
	t.Run("without admin cookie", func(t *testing.T) {
		resp := ts.Serve(ts.Get(t, "/admin", nil))
		if resp.StatusCode == http.StatusOK {
			t.Errorf("Expected redirect got %d", resp.StatusCode)
		}
	})
}
