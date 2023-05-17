package admin_test

import (
	"bytes"
	_ "embed"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/natsukagami/kjudge/models"
	"github.com/natsukagami/kjudge/test"
)

var (
	//go:embed test_assets/tests.zip
	testsZip []byte

	testsLen = 10
)

func TestUploadMultipleTests(t *testing.T) {
	testGroup := 7
	path := fmt.Sprintf("/admin/test_groups/%d/upload_multiple", testGroup)

	makeReq := func(t *testing.T, input, output string, override string) *http.Request {
		buf := bytes.NewBuffer(nil)
		body := multipart.NewWriter(buf)

		if err := body.WriteField("input", input); err != nil {
			t.Fatalf("writing field `input`: %v", err)
		}
		if err := body.WriteField("output", output); err != nil {
			t.Fatalf("writing field `output`: %v", err)
		}
		file, err := body.CreateFormFile("file", "tests.zip")
		if err != nil {
			t.Fatalf("writing field `file`: %v", err)
		}
		if _, err := file.Write(testsZip); err != nil {
			t.Fatalf("writing file: %v", err)
		}
		if err := body.WriteField("override", override); err != nil {
			t.Fatalf("writing field `override`: %v", err)
		}
		if err := body.Close(); err != nil {
			t.Fatalf("writing tail: %v", err)
		}

		req := httptest.NewRequest(http.MethodPost, path, buf)
		req.Header.Add(echo.HeaderContentType, body.FormDataContentType())

		return req
	}

	t.Run("fail no admin", func(t *testing.T) {
		ts := test.NewServer(t)

		// get the tests before
		testsBefore, err := models.GetTestGroupTests(ts.DB, testGroup)
		if err != nil {
			t.Fatal("fetching old tests", err)
		}

		req := makeReq(t, "?.in", "?.out", "false")
		resp := ts.Serve(req)

		if resp.StatusCode != http.StatusTemporaryRedirect {
			t.Errorf("expected redirect got %d", resp.StatusCode)
		}

		testsAfter, err := models.GetTestGroupTests(ts.DB, testGroup)
		if err != nil {
			t.Fatal("fetching new tests", err)
		}
		if len(testsAfter) != len(testsBefore) {
			t.Errorf("expected %d tests, got %d", len(testsBefore), len(testsAfter))
		}
		for i, test := range testsAfter {
			if testsBefore[i].ID != test.ID { // but different IDs
				t.Errorf("test %d: unexpected different ID", i)
			}
		}
	})

	t.Run("success override", func(t *testing.T) {
		ts := test.NewServer(t)
		withAdmin := ts.WithAdmin(t)

		// get the tests before
		testsBefore, err := models.GetTestGroupTests(ts.DB, testGroup)
		if err != nil {
			t.Fatal("fetching old tests", err)
		}

		req := makeReq(t, "?.in", "?.out", "true")
		resp := ts.Serve(req, withAdmin)

		if resp.StatusCode >= 400 {
			t.Errorf("expected success got %d", resp.StatusCode)
		}

		testsAfter, err := models.GetTestGroupTests(ts.DB, testGroup)
		if err != nil {
			t.Fatal("fetching new tests", err)
		}
		if len(testsAfter) != testsLen {
			t.Errorf("expected %d tests, got %d", testsLen, len(testsAfter))
		}
		for i, test := range testsAfter {
			if testsBefore[i].Name != test.Name { // they have the same names
				t.Errorf("test %d: expected name `%s` got `%s`", i, testsBefore[i].Name, test.Name)
			}
			if testsBefore[i].ID == test.ID { // but different IDs
				t.Errorf("test %d: unexpected same ID", i)
			}

			if string(test.Input) != fmt.Sprintf("%d\n", i+1) {
				t.Errorf("test %d: unexpected inputs: %v", i, test.Input)
			}
			if string(test.Output) != fmt.Sprintf("%d\n", 3*(i+1)) {
				t.Errorf("test %d: unexpected inputs: %v", i, test.Input)
			}
		}
	})

	t.Run("fail no-override", func(t *testing.T) {
		ts := test.NewServer(t)
		withAdmin := ts.WithAdmin(t)

		// get the tests before
		testsBefore, err := models.GetTestGroupTests(ts.DB, testGroup)
		if err != nil {
			t.Fatal("fetching old tests", err)
		}

		req := makeReq(t, "?.in", "?.out", "false")
		resp := ts.Serve(req, withAdmin)

		if resp.StatusCode < 400 {
			t.Errorf("expected failure got %d", resp.StatusCode)
		}

		testsAfter, err := models.GetTestGroupTests(ts.DB, testGroup)
		if err != nil {
			t.Fatal("fetching new tests", err)
		}
		if len(testsAfter) != len(testsBefore) {
			t.Errorf("expected %d tests, got %d", len(testsBefore), len(testsAfter))
		}
		for i, test := range testsAfter {
			if testsBefore[i].ID != test.ID { // but different IDs
				t.Errorf("test %d: unexpected different ID", i)
			}
		}
	})
}
