// Package test provides some utilities for integration testing at endpoint levels.
package test

import (
	"crypto/rand"
	_ "embed"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/natsukagami/kjudge/db"
	"github.com/natsukagami/kjudge/server"
	"github.com/natsukagami/kjudge/server/auth"
	"github.com/pkg/errors"
)

// the embedded test data.
//
//go:embed data.sql
var testData string

// Content of the kjudge.db containing the full test database.
var databaseContent []byte

func init() {
	tmpDir, err := os.MkdirTemp(os.TempDir(), "kjudge_test")
	if err != nil {
		log.Panic("cannot create temp dir:", err)
	}
	defer os.RemoveAll(tmpDir)
	dbFile := filepath.Join(tmpDir, "kjudge.db")
	tmpDb, err := db.New(dbFile)
	if err != nil {
		log.Panic("cannot create temp database:", errors.WithStack(err))
	}
	if _, err := tmpDb.Exec(testData); err != nil {
		log.Panic("Cannot import test data:", err)
	}
	if err := tmpDb.Close(); err != nil {
		log.Panic("cannot create temp database:", errors.WithStack(err))
	}
	dbFileContent, err := os.ReadFile(dbFile)
	if err != nil {
		log.Panic("cannot read temp database:", errors.WithStack(err))
	}
	databaseContent = dbFileContent
}

// TestServer wraps Server and adds some fancy stuff.
type TestServer struct {
	*server.Server
	DB *db.DB
}

// NewDB creates and populates a test database.
func NewDB(t *testing.T) *db.DB {
	tmpDbFile := filepath.Join(t.TempDir(), "kjudge.db")
	if err := os.WriteFile(tmpDbFile, databaseContent, 0644); err != nil {
		t.Fatal("cannot create temp database:", errors.WithStack(err))
	}
	tmpDb, err := db.New(tmpDbFile)
	if err != nil {
		t.Fatal(errors.WithStack(err))
	}

	return tmpDb
}

// NewServer creates a new kjudge server running on a test database.
func NewServer(t *testing.T) *TestServer {
	// generate an admin key
	adminKey := make([]byte, 32)
	if _, err := rand.Read(adminKey); err != nil {
		t.Fatal("generating admin key:", err)
	}
	t.Setenv(auth.AdminKeyEnv, fmt.Sprintf("%x", adminKey))

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

// Get fires a new GET request with URL queries.
func (ts *TestServer) Get(t *testing.T, path string, queries url.Values) *http.Request {
	if len(queries) > 0 {
		path = path + "?" + queries.Encode()
	}
	req := httptest.NewRequest(http.MethodGet, path, nil)

	return req
}

// Serve serves a HTTP request and returns its response.
func (ts *TestServer) Serve(req *http.Request, opts ...ReqOpt) *http.Response {
	for _, opt := range opts {
		opt(req)
	}
	rec := httptest.NewRecorder()
	ts.ServeHTTP(rec, req)
	return rec.Result()
}

// WithMisaka logs in as Misaka for the next request.
func (ts *TestServer) WithMisaka(t *testing.T) ReqOpt {
	// Perform the log in.
	form := url.Values{}
	form.Set("id", "misaka")
	form.Set("password", "misaka")
	resp := ts.Serve(ts.PostForm(t, "/user/login", form))

	if resp.StatusCode >= 400 {
		t.Fatalf("Cannot login as misaka: got code %d", resp.StatusCode)
	}
	cookies := resp.Cookies()
	if len(cookies) != 1 {
		t.Fatalf("Cannot login as misaka: expect one cookie, got %#v", cookies)
	}

	return func(req *http.Request) {
		req.AddCookie(cookies[0])
	}
}

// WithAdmin logs in with the admin panel cookie for the next request.
func (ts *TestServer) WithAdmin(t *testing.T) ReqOpt {
	// Perform the log in.
	form := url.Values{}
	form.Set("key", os.Getenv(auth.AdminKeyEnv))
	resp := ts.Serve(ts.PostForm(t, "/admin/login", form))

	if resp.StatusCode >= 400 {
		t.Fatalf("Cannot login to admin panel: got code %d", resp.StatusCode)
	}
	cookies := resp.Cookies()
	if len(cookies) != 1 {
		t.Fatalf("Cannot login to admin panel: expect one cookie, got %#v", cookies)
	}

	return func(req *http.Request) {
		req.AddCookie(cookies[0])
	}
}

// ReqOpt is an option for sending requests.
type ReqOpt func(*http.Request)
