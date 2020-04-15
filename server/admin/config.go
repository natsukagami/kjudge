package admin

import (
	"net/http"

	"github.com/natsukagami/kjudge/models"
	"github.com/natsukagami/kjudge/server/httperr"
	"github.com/labstack/echo/v4"
)

// ConfigTogglePost implements POST /admin/config/toggle
func (g *Group) ConfigTogglePost(c echo.Context) error {
	config, err := models.GetConfig(g.db)
	if err != nil {
		return err
	}
	switch c.FormValue("key") {
	case "enable_registration":
		config.EnableRegistration = !config.EnableRegistration
	case "enable_user_customization":
		config.EnableUserCustomization = !config.EnableUserCustomization
	default:
		return httperr.BadRequestf("Unknown key: `%s`", c.FormValue("key"))
	}
	if err := config.Write(g.db); err != nil {
		return err
	}
	return c.Redirect(http.StatusSeeOther, "/admin/users")
}
