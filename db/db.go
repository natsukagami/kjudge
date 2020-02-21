package db

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
)

// DB is an implementation of the underlying DB.
type DB struct {
	*sqlx.DB
}

// New creates a new DB object from the given filename.
func New(filename string) (*DB, error) {
	sqlxdb, err := sqlx.Open("sqlite3", fmt.Sprintf("%s?_fk=1&mode=rw", filename))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	db := &DB{
		DB: sqlxdb,
	}
	// Perform migrations, if needed.
	if err := db.migrate(); err != nil {
		return nil, err
	}

	return db, nil
}
