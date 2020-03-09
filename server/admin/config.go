package admin

import (
	"net/http"

	"git.nkagami.me/natsukagami/kjudge/models"
	"github.com/labstack/echo/v4"
)

// ToggleEnableRegistration implements POST /admin/config/toggle_enable_registration
func (g *Group) ToggleEnableRegistration(c echo.Context) error {
	config, err := models.GetConfig(g.db)
	if err != nil {
		return err
	}
	config.EnableRegistration = !config.EnableRegistration
	if err := config.Write(g.db); err != nil {
		return err
	}
	return c.Redirect(http.StatusSeeOther, "/admin/users")
}
