// Package test provides some utilities for integration testing at endpoint levels.
package test

import (
	_ "embed"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/natsukagami/kjudge/db"
	"github.com/natsukagami/kjudge/server"
	"github.com/pkg/errors"
)

// the embedded test data.
//
//go:embed data.sql
var testData string

// TestServer wraps Server and adds some fancy stuff.
type TestServer struct {
	*server.Server
	DB *db.DB
}

// NewDB creates and populates a test database.
func NewDB(t *testing.T) *db.DB {
	tmpDbFile := filepath.Join(t.TempDir(), "kjudge.db")
	tmpDb, err := db.New(tmpDbFile)
	if err != nil {
		t.Fatal(errors.WithStack(err))
	}

	if _, err := tmpDb.Exec(testData); err != nil {
		t.Fatal("Cannot import test data: ", err)
	}
	return tmpDb
}

// NewServer creates a new kjudge server running on a test database.
func NewServer(t *testing.T) *TestServer {
	db := NewDB(t)
	s, err := server.New(db)
	if err != nil {
		t.Fatal(err)
	}
	return &TestServer{Server: s, DB: db}
}

// PostForm fires a new POST request with a form body.
func (ts *TestServer) PostForm(t *testing.T, path string, body url.Values) *http.Request {
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)

	return req
}

// Serve serves a HTTP request and returns its response.
func (ts *TestServer) Serve(req *http.Request) *http.Response {
	rec := httptest.NewRecorder()
	ts.ServeHTTP(rec, req)
	return rec.Result()
}
