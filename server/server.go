package server

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"git.nkagami.me/natsukagami/kjudge/db"
	"git.nkagami.me/natsukagami/kjudge/models"
	"git.nkagami.me/natsukagami/kjudge/server/admin"
	"git.nkagami.me/natsukagami/kjudge/server/template"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/pkg/errors"
)

// Server this the root entry of the server.
type Server struct {
	db   *db.DB
	echo *echo.Echo
}

func (s *Server) HTTPErrorHandler(err error, c echo.Context) {
	var e *echo.HTTPError
	if errors.As(err, &e) {
		s.echo.DefaultHTTPErrorHandler(err, c)
	} else {
		log.Printf("%+v", err)
		c.JSON(http.StatusInternalServerError, "Internal Server Error")
	}
}

// New creates a new server.
func New(db *db.DB) (*Server, error) {
	s := Server{
		db:   db,
		echo: echo.New(),
	}

	// Load the configuration
	config, err := models.GetConfig(db)
	if errors.Is(err, sql.ErrNoRows) {
		config, err = models.GenerateConfig()
	}
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if err := config.Write(db); err != nil {
		return nil, err
	}

	// Perform linking for Echo here
	// ...
	s.echo.HideBanner = true
	s.echo.Renderer = template.Renderer{}
	s.echo.HTTPErrorHandler = s.HTTPErrorHandler
	s.echo.Use(session.Middleware(sessions.NewCookieStore(config.SessionKey)))
	s.echo.Use(middleware.Recover())
	s.echo.Use(middleware.Gzip())

	admin.New(s.echo.Group("/admin"), s.db)
	s.echo.GET("*", StaticFiles)

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
