package user

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/natsukagami/kjudge/db"
	"github.com/natsukagami/kjudge/models"
	"github.com/natsukagami/kjudge/server/auth"
	"github.com/natsukagami/kjudge/server/httperr"
	"github.com/pkg/errors"
)

// HomeCtx is the context to render the /user page.
type HomeCtx struct {
	*AuthCtx

	EnableUserCustomization bool
	CustomizeError          error
	Customize               CustomizeForm

	ChangePasswordError error
	ChangePassword      ChangePasswordForm
}

// CustomizeForm is a form for an user to change their name or organization.
type CustomizeForm struct {
	DisplayName  string `form:"display_name"`
	Organization string `form:"organization"`
}

// Load a default CustomizeForm from an user.
func userCustomizeForm(u *models.User) CustomizeForm {
	return CustomizeForm{
		DisplayName:  u.DisplayName,
		Organization: u.Organization,
	}
}

// Bind binds the form to an user.
func (c *CustomizeForm) Bind(u *models.User) {
	u.DisplayName = strings.TrimSpace(c.DisplayName)
	u.Organization = strings.TrimSpace(c.Organization)
}

// Render renders the context.
func (h *HomeCtx) Render(c echo.Context) error {
	status := http.StatusOK
	if h.CustomizeError != nil || h.ChangePasswordError != nil {
		status = http.StatusBadRequest
	}
	return c.Render(status, "user/home", h)
}

func getHomeCtx(db db.DBContext, c echo.Context) (*HomeCtx, error) {
	me, err := Me(db, c)
	if err != nil {
		return nil, err
	}
	config, err := models.GetConfig(db)
	if err != nil {
		return nil, err
	}
	return &HomeCtx{
		AuthCtx:                 me,
		EnableUserCustomization: config.EnableUserCustomization,
		Customize:               userCustomizeForm(me.Me),
	}, nil
}

// ChangePasswordForm is the form handling a password change request.
type ChangePasswordForm struct {
	CurrentPassword string `form:"current_password"`
	NewPassword     string `form:"new_password"`
}

// Bind binds the content of the form into the user.
func (c *ChangePasswordForm) Bind(u *models.User) error {
	if ok, err := auth.CheckPassword(c.CurrentPassword, u.Password); err != nil {
		return err
	} else if !ok {
		return errors.New("Wrong current password")
	}
	pwd, err := auth.PasswordHash(c.NewPassword)
	if err != nil {
		return err
	}
	u.Password = string(pwd)
	return nil
}

// HomeGet implements "/user".
func (g *Group) HomeGet(c echo.Context) error {
	ctx, err := getHomeCtx(g.db, c)
	if err != nil {
		return err
	}
	return ctx.Render(c)
}

// ChangePassword implements POST /user/change_password.
func (g *Group) ChangePassword(c echo.Context) error {
	tx, err := g.db.Beginx()
	if err != nil {
		return errors.WithStack(err)
	}
	defer db.Rollback(tx)

	ctx, err := getHomeCtx(tx, c)
	if err != nil {
		return err
	}

	handle := func() error {
		if err := c.Bind(&ctx.ChangePassword); err != nil {
			return err
		}
		if err := ctx.ChangePassword.Bind(ctx.Me); err != nil {
			return err
		}
		if err := ctx.Me.Write(tx); err != nil {
			return err
		}
		if err := tx.Commit(); err != nil {
			return errors.WithStack(err)
		}
		return nil
	}

	if err := handle(); err != nil {
		ctx.ChangePasswordError = err
		return ctx.Render(c)
	}
	return c.Redirect(http.StatusSeeOther, "/user")
}

// CustomizePost implements POST /user/customize.
func (g *Group) CustomizePost(c echo.Context) error {
	tx, err := g.db.Beginx()
	if err != nil {
		return errors.WithStack(err)
	}
	defer db.Rollback(tx)
	ctx, err := getHomeCtx(tx, c)
	if err != nil {
		return err
	}
	if !ctx.EnableUserCustomization {
		return httperr.BadRequestf("User customization has been disabled by the contest organizer.")
	}
	if err := c.Bind(&ctx.Customize); err != nil {
		return httperr.BindFail(err)
	}
	u := *ctx.Me
	ctx.Customize.Bind(&u)
	if err := u.Write(tx); err != nil {
		ctx.CustomizeError = err
		return ctx.Render(c)
	}
	if err := tx.Commit(); err != nil {
		return errors.WithStack(err)
	}
	return c.Redirect(http.StatusSeeOther, "/user")
}
