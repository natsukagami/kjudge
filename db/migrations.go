package db

import (
	"database/sql"
	"fmt"
	"log"
	"sort"

	"git.nkagami.me/natsukagami/kjudge/static"
	"github.com/pkg/errors"
)

// Attempt to migrate to a newer version of the schema, if any.
func (db *DB) migrate() error {
	// First we always attempt to create a version table.
	if _, err := db.Exec("CREATE TABLE IF NOT EXISTS version (version VARCHAR NOT NULL);"); err != nil {
		return errors.WithStack(err)
	}
	// Now get the schema version of the DB
	version, err := db.getSchemaVersion()
	if err != nil {
		return err
	}

	versions, err := getSchemaFiles()
	if err != nil {
		return err
	}

	if version != "" {
		// Filter away the versions that are already migrated
		sqlFileString := fmt.Sprintf("sql/%s.sql", version)
		for len(versions) > 0 && versions[0] <= sqlFileString {
			versions = versions[1:]
		}
	}

	// Do migrations one by one
	for _, path := range versions {
		file, err := static.ReadFile(path)
		if err != nil {
			return errors.Wrapf(err, "File %s", path)
		}
		if _, err := db.Exec(string(file)); err != nil {
			return errors.Wrapf(err, "File %s", path)
		}
		log.Printf("DB migrated to schema: %s", path)
	}
	return nil
}

// Gets the schema version of the DB.
func (db *DB) getSchemaVersion() (string, error) {
	row := db.QueryRow("SELECT version FROM version")
	var version string
	if err := row.Scan(&version); err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", errors.WithStack(err)
	}
	return version, nil
}

// Collect the schema files from the static.
func getSchemaFiles() ([]string, error) {
	files, err := static.WalkDirs("sql", false)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	sort.Slice(files, func(i, j int) bool { return files[i] < files[j] })
	return files, nil
}
