package server

import (
	"io"
	"net/http"
	"os"
	stdPath "path"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/natsukagami/kjudge/embed"
	"github.com/pkg/errors"
)

// StaticFiles serves files from the source fileb0x.
// It filters away files that don't end with ".css", ".js" or ".map"
func StaticFiles(c echo.Context) error {
	path := c.Request().URL.Path
	for _, suffix := range []string{".woff2", ".woff", ".css", ".js", ".map", ".png", ".ogg"} {
		if strings.HasSuffix(path, suffix) {
			return serveFile(stdPath.Join("templates", path), c)
		}
	}
	return NotFoundHandler(c)
}

func serveFile(file string, c echo.Context) error {
	f, err := embed.Content.Open(file)
	if err != nil {
		if os.IsNotExist(err) {
			return echo.NewHTTPError(http.StatusNotFound)
		}
		return errors.WithStack(err)
	}
	stat, err := f.Stat()
	if err != nil {
		return errors.WithStack(err)
	}
	http.ServeContent(c.Response(), c.Request(), stat.Name(), stat.ModTime(), f.(io.ReadSeeker))
	return nil
}
