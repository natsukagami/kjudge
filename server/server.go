package server

import (
	"fmt"

	"git.nkagami.me/natsukagami/kjudge/db"
	"github.com/labstack/echo/v4"
)

// Server this the root entry of the server.
type Server struct {
	db   *db.DB
	echo *echo.Echo
}

// New creates a new server.
func New(db *db.DB) (*Server, error) {
	s := Server{
		db:   db,
		echo: echo.New(),
	}

	// Perform linking for Echo here
	// ...
	s.echo.HideBanner = true

	return &s, nil
}

// Start starts the server, listening for requests.
func (s *Server) Start(port int) error {
	return s.echo.Start(fmt.Sprintf(":%d", port))
}

// StartWithTLS starts the server, also tries to get a cert from LetsEncrypt.
func (s *Server) StartWithTLS(address string) error {
	return s.echo.StartAutoTLS(address)
}
