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

func getUsers(db db.DBContext, c echo.Context) (*UsersCtx, error) {
	users, err := models.GetAllUsers(db)
	if err != nil {
    	println(err)
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
