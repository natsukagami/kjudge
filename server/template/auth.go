package template

import "github.com/natsukagami/kjudge/models"

// Me should implement a context with a logged in user.
type Me interface {
	GetMe() *models.User
}

// Checks if the given context implements Me and returns a non-nil Me.
func loggedIn(ctx interface{}) bool {
	v, ok := ctx.(Me)
	return ok && v.GetMe() != nil
}
