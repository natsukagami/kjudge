package admin

import (
	"net/http"

	"git.nkagami.me/natsukagami/kjudge/db"
	"git.nkagami.me/natsukagami/kjudge/models"
	"github.com/labstack/echo/v4"
)

// UsersCtx provides a context for rendering /admin/users.
type UsersCtx struct {
	Users []*models.User

	FormError error
	Form      UserForm
}

// UserForm is a form for adding or editing an user.
type UserForm struct {
	ID       string `form:"id"`
	Password string `form:"password"`
	Hidden   bool   `form:"hidden"`
}

// Bind binds the form's values to the model.
func (f *UserForm) Bind(u *models.User) {
	u.ID = f.ID
	u.Password = f.Password
	u.Hidden = f.Hidden
}

func getUsers(db db.DBContext, c echo.Context) (*UsersCtx, error) {
	users, err := models.GetAllUsers(db)
	if err != nil {
		return nil, err
	}
	return &UsersCtx{Users: users}, nil
}

// UsersGet implements GET /admin/users.
func (g *Group) UsersGet(c echo.Context) error {
	ctx, err := getUsers(g.db, c)
	if err != nil {
		return err
	}
	return c.Render(http.StatusOK, "admin/users", ctx)
}

func (g *Group) UsersAdd(c echo.Context) error {
	var form UserForm
	if err := c.Bind(&form); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	var u models.User
	form.Bind(&u)
	if err := u.Write(g.db); err != nil {
		ctx, ctxErr := getUsers(g.db, c)
		if ctxErr != nil {
			return ctxErr
		}
		ctx.FormError = err
		return c.Render(http.StatusBadRequest, "admin/users", ctx)
	}
	return c.Redirect(http.StatusSeeOther, "/admin/users")
}
