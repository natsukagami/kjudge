package admin

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/natsukagami/kjudge/models"
	"github.com/natsukagami/kjudge/models/verify"
	"github.com/natsukagami/kjudge/server/auth"
	"github.com/natsukagami/kjudge/server/httperr"
	"github.com/pkg/errors"
)

var csvHeaders = []string{"Username", "Display Name", "Organization", "Password", "Hidden"}

const csvType = "text/csv"

func writeCSVToBytes(records ...[]string) ([]byte, error) {
	var b bytes.Buffer
	writer := csv.NewWriter(&b)
	if err := writer.Write(csvHeaders); err != nil {
		return nil, errors.WithStack(err)
	}
	if err := writer.WriteAll(records); err != nil {
		return nil, errors.WithStack(err)
	}
	return b.Bytes(), nil
}

// BatchUsersEmptyGet implements GET /admin/batch_users/empty.
func (g *Group) BatchUsersEmptyGet(c echo.Context) error {
	b, err := writeCSVToBytes()
	if err != nil {
		return err
	}
	c.Response().Header().Add("Content-Disposition", `attachment; filename="users.csv"`)
	return c.Blob(http.StatusOK, csvType, b)
}

// BatchUsersGenerateForm is the form for generating a CSV users file.
type BatchUsersGenerateForm struct {
	Prefix string `form:"prefix"`
	Count  int    `form:"count"`
	Offset int    `form:"offset"`
}

// Verify verifies the content of the form.
func (f *BatchUsersGenerateForm) Verify() error {
	return verify.All(map[string]error{
		"Prefix": verify.String(f.Prefix, verify.StringNonEmpty, verify.StringMaxLength(32-3)),
		"Count":  verify.Int(f.Count, verify.IntMin(1), verify.IntMax(999)),
		"Offset": verify.Int(f.Offset, verify.IntMin(0), verify.IntMax(999)),
	})
}

// BatchUsersGenerateGet implements GET /admin/batch_users/generate
func (g *Group) BatchUsersGenerateGet(c echo.Context) error {
	var form BatchUsersGenerateForm
	if err := c.Bind(&form); err != nil {
		return httperr.BindFail(err)
	}
	if err := form.Verify(); err != nil {
		return err
	}
	records := [][]string{}
	passwords, err := auth.GeneratePassword(form.Count)
	if err != nil {
		return err
	}
	for i := 0; i < form.Count; i++ {
		id := form.Offset + 1 + i
		userID := fmt.Sprintf("%s%03d", form.Prefix, id)
		records = append(records, []string{userID, userID, "", passwords[i], "0"})
	}
	b, err := writeCSVToBytes(records...)
	if err != nil {
		return err
	}
	c.Response().Header().Add("Content-Disposition", `attachment; filename="users.csv"`)
	return c.Blob(http.StatusOK, "text/csv", b)
}

// BatchUsersPost implements POST /admin/batch_users.
func (g *Group) BatchUsersPost(c echo.Context) error {
	mp, err := c.MultipartForm()
	if err != nil {
		return httperr.BindFail(err)
	}
	if count := len(mp.File["file"]); count != 1 {
		return httperr.BadRequestf("Expected 1 file, got %d", count)
	}
	file, err := mp.File["file"][0].Open()
	if err != nil {
		return httperr.BindFail(err)
	}
	defer file.Close()

	users, rows, err := readCSVFile(file)
	if err != nil {
		return err
	}

	tx, err := g.db.Beginx()
	if err != nil {
		return errors.WithStack(err)
	}
	defer tx.Rollback()

	if err := models.BatchAddUsers(tx, c.FormValue("reset") == "true", users...); err != nil {
		return err
	}

	// Write the CSV records
	b, err := writeCSVToBytes(rows...)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return errors.WithStack(err)
	}

	c.Response().Header().Add("Content-Disposition", `attachment; filename="users.csv"`)
	return c.Blob(http.StatusOK, "text/csv", b)
}

func readCSVFile(source io.Reader) (users []*models.User, rows [][]string, err error) {
	reader := csv.NewReader(source)
	reader.FieldsPerRecord = len(csvHeaders)
	// Read the header
	header, err := reader.Read()
	if err != nil {
		return nil, nil, err
	}
	for i, head := range csvHeaders {
		if header[i] != head {
			return nil, nil, httperr.BadRequestf("Invalid CSV file: Headers don't match: Expected %s, got %s", head, header[i])
		}
	}
	// Read the records
	rows, err = reader.ReadAll()
	if err != nil {
		return nil, nil, httperr.BindFail(err)
	}
	for _, row := range rows {
		for i, cell := range row {
			row[i] = strings.TrimSpace(cell)
		}
		u := &models.User{
			ID:           row[0],
			DisplayName:  row[1],
			Organization: row[2],
			Password:     row[3],
			Hidden:       row[4] == "true" || row[4] == "1",
		}
		// Fill the holes
		if u.DisplayName == "" {
			row[1] = u.ID
			u.DisplayName = u.ID
		}
		if u.Password == "" {
			pwd, err := auth.GeneratePassword(1)
			if err != nil {
				return nil, nil, err
			}
			row[3] = pwd[0]
			u.Password = pwd[0]
		}
		row[4] = "0"
		if u.Hidden {
			row[4] = "1"
		}

		hashed, err := auth.PasswordHash(u.Password)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "User %s", u.ID)
		}
		u.Password = string(hashed)
		users = append(users, u)
	}
	return users, rows, nil
}
