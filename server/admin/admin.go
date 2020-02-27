// Package admin defines admin routes.
package admin

import (
	"git.nkagami.me/natsukagami/kjudge/db"
	"github.com/labstack/echo/v4"
)

// Group represents a router Group with handling functions.
type Group struct {
	*echo.Group
	db *db.DB
}

// New creates a new group.
func New(db *db.DB) *Group {
}

func (g *Group) 
