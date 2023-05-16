// Command migrate performs migration on a given database.
package main

import (
	"flag"
	"io"
	"log"
	"os"

	"github.com/natsukagami/kjudge/db"
	"github.com/pkg/errors"
)

var (
	dbfile = flag.String("file", "kjudge.db", "Path to the database file")
	reset  = flag.Bool("reset", false, "Reset the database")
)

func main() {
	flag.Parse()

	if *reset {
		log.Println("Removing the old database.")
		os.Remove(*dbfile)
	}

	database, err := db.New(*dbfile)
	if err != nil {
		log.Fatalf("%+v", err)
	}

	log.Println("Now reading SQL commands from the standard input.")
	sql, err := io.ReadAll(os.Stdin)
	if err != nil {
		log.Fatalf("%+v", errors.WithStack(err))
	}
	if _, err := database.Exec(string(sql)); err != nil {
		log.Fatalf("%+v", errors.WithStack(err))
	}

	if err := database.Close(); err != nil {
		log.Fatalf("Error closing the database: %+v", err)
	}
}
