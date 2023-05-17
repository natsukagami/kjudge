package server

import (
	"context"
	"log"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/pkg/errors"
)

// ServeHTTPRootCA starts a HTTP server running on `address` serving the .pem file at
// rootCA, serving the root CA from "/ca". It rejects all other requests.
func (s *Server) ServeHTTPRootCA(address, rootCA string) error {
	if stat, err := os.Stat(rootCA); err != nil {
		return errors.WithStack(err)
	} else if stat.IsDir() {
		return errors.Errorf("cannot use rootCA: %s is a directory", rootCA)
	}
	server := echo.New()
	if s.verbose {
		server.Use(middleware.Logger())
	}
	server.GET("/ca", func(c echo.Context) error {
		return c.Attachment(rootCA, "root.pem")
	})
	server.HideBanner = true
	server.HidePort = true
	defer func() { _ = server.Shutdown(context.Background()) }()
	log.Printf("Root certificate is being served on '%s', obtain the Root CA on /ca", address)
	return errors.WithStack(server.Start(address))
}
