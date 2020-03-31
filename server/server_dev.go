// +build !production

package server

import (
	"net/http"
	"net/http/pprof"

	"github.com/labstack/echo/v4"
)

// SetupProfiling sets up profiling for the server in development mode.
func (s *Server) SetupProfiling() {
	// Turn on debugging mode
	s.echo.Debug = true

	s.echo.Any("/debug/pprof/*", echo.WrapHandler(http.HandlerFunc(pprof.Index)))
	s.echo.Any("/debug/pprof/cmdline", echo.WrapHandler(http.HandlerFunc(pprof.Cmdline)))
	s.echo.Any("/debug/pprof/profile", echo.WrapHandler(http.HandlerFunc(pprof.Profile)))
	s.echo.Any("/debug/pprof/symbol", echo.WrapHandler(http.HandlerFunc(pprof.Symbol)))
	s.echo.Any("/debug/pprof/trace", echo.WrapHandler(http.HandlerFunc(pprof.Trace)))
}
