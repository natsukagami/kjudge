package server

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"git.nkagami.me/natsukagami/kjudge/db"
	"git.nkagami.me/natsukagami/kjudge/models"
	"git.nkagami.me/natsukagami/kjudge/models/verify"
	"git.nkagami.me/natsukagami/kjudge/server/admin"
	"git.nkagami.me/natsukagami/kjudge/server/contests"
	"git.nkagami.me/natsukagami/kjudge/server/template"
	"git.nkagami.me/natsukagami/kjudge/server/user"
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
	s.echo.HTTPErrorHandler = s.HandleError
	s.echo.Use(session.Middleware(sessions.NewCookieStore(config.SessionKey)))
	s.echo.Use(middleware.Recover())
	s.echo.Use(middleware.Gzip())

	s.SetupProfiling()

	if _, err := admin.New(s.db, s.echo.Group("/admin")); err != nil {
		return nil, err
	}
	if _, err := user.New(s.db, s.echo.Group("/user")); err != nil {
		return nil, err
	}
	contests, err := contests.New(s.db, s.echo.Group("/contests"))
	if err != nil {
		return nil, err
	}
	s.echo.GET("", contests.ConestsGetNearestOngoingContest)
	s.echo.GET("*", StaticFiles)
	s.echo.POST("*", NotFoundHandler)

	return &s, nil
}

// NotFoundHandler handles the "not found" situation. It should be a catch-all for all urls.
func NotFoundHandler(c echo.Context) error {
	return echo.NewHTTPError(http.StatusNotFound, "The page you are looking for does not exist")
}

// HandleError defines an error handler that complies with echo's standards.
func (s *Server) HandleError(err error, c echo.Context) {
	type errCtx struct {
		Code       int
		Message    string
		StatusText string
	}
	// the convention is:
	// - if err is *echo.HTTPError, it is a "normal error" with its own message and everything.
	// - otherwise, it is an unexpected error.

	// if err is verify.Error, it is a "normal error" with statusCode = 400 Bad Request
	if errors.As(err, &verify.Error{}) {
		err = echo.NewHTTPError(http.StatusBadRequest, err)
	}

	if e, ok := err.(*echo.HTTPError); ok {
		// Just handle it gracefully
		c.Render(e.Code, "error", errCtx{Code: e.Code, Message: fmt.Sprint(e.Message), StatusText: http.StatusText(e.Code)})
	} else {
		// internal error: dump it.
		c.Render(http.StatusInternalServerError, "error", errCtx{Code: http.StatusInternalServerError})

		errStr := fmt.Sprintf("An unexpected error has occured: %v\n", err)
		path := filepath.Join(os.TempDir(), fmt.Sprintf("kjudge-%v.txt", time.Now().Format(time.RFC3339)))
		if err := ioutil.WriteFile(path, []byte(fmt.Sprintf("%+v", errors.WithStack(err))), 0644); err != nil {
			errStr += fmt.Sprintf("Cannot log the error down to file: %v", err)
		} else {
			errStr += fmt.Sprintf(`The error has been logged down to file '%s'.
Please check out the open issues and help opening a new one if possible on https://git.nkagami.me/natsukagami/kjudge/issues/new`, path)
		}
		log.Println(errStr)
	}
}

// Start starts the server, listening for requests.
func (s *Server) Start(port int) error {
	return s.echo.Start(fmt.Sprintf(":%d", port))
}

// StartWithTLS starts the server, also tries to get a cert from LetsEncrypt.
func (s *Server) StartWithTLS(address string) error {
	return s.echo.StartAutoTLS(address)
}
