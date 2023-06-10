package db

import (
	"database/sql"
	"io/fs"
	"log"
	"path"
	"regexp"
	"sort"

	"github.com/natsukagami/kjudge/embed"
	"github.com/pkg/errors"
)

var versionRegexp = regexp.MustCompile(`(.+)\.sql`)

const (
	assetsSql = "assets/sql"
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
		for len(versions) > 0 && versions[0] <= version {
			versions = versions[1:]
		}
	}

	// Do migrations one by one
	for _, name := range versions {
		sqlFile := path.Join(assetsSql, name+".sql")
		file, err := fs.ReadFile(embed.Content, sqlFile)
		if err != nil {
			return errors.Wrapf(err, "File %s", sqlFile)
		}
		if _, err := db.Exec(string(file)); err != nil {
			return errors.Wrapf(err, "File %s", sqlFile)
		}
		log.Printf("DB migrated to schema: %s", sqlFile)
		version = name
	}

	// Update the schema version
	if _, err := db.Exec("UPDATE version SET version = ?", version); err != nil {
		return errors.WithStack(err)
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
	files, err := fs.ReadDir(embed.Content, assetsSql)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	var names []string
	for _, file := range files {
		matches := versionRegexp.FindAllStringSubmatch(file.Name(), 1)
		if len(matches) == 1 {
			names = append(names, matches[0][1])
		}
	}
	sort.Slice(names, func(i, j int) bool { return names[i] < names[j] })
	return names, nil
}
