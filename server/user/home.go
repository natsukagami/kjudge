package user

import (
	"net/http"

	"git.nkagami.me/natsukagami/kjudge/db"
	"git.nkagami.me/natsukagami/kjudge/models"
	"git.nkagami.me/natsukagami/kjudge/server/auth"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

// HomeCtx is the context to render the /user page.
type HomeCtx struct {
	*AuthCtx

	ChangePasswordError error
	ChangePassword      ChangePasswordForm
}

// Render renders the context.
func (h *HomeCtx) Render(c echo.Context) error {
	status := http.StatusOK
	if h.ChangePasswordError != nil {
		status = http.StatusBadRequest
	}
	return c.Render(status, "user/home", h)
}

func getHomeCtx(db db.DBContext, c echo.Context) (*HomeCtx, error) {
	me, err := Me(db, c)
	if err != nil {
		return nil, err
	}
	return &HomeCtx{AuthCtx: me}, nil

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
		return err
	}
	defer tx.Rollback()

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
			return err
		}
		return nil
	}

	if err := handle(); err != nil {
		ctx.ChangePasswordError = err
		return ctx.Render(c)
	}
	return c.Redirect(http.StatusSeeOther, "/user")
}
