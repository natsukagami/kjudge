package main

import (
	"flag"
	"log"

	_ "git.nkagami.me/natsukagami/kjudge"
	_ "git.nkagami.me/natsukagami/kjudge/models"
	"git.nkagami.me/natsukagami/kjudge/db"
)

var (
	dbfile = flag.String("file", "kjudge.db", "Path to the database file")
)

func main() {
	flag.Parse()

	_, err := db.New(*dbfile)
	if err != nil {
		log.Fatalf("%+v", err)
	}
}
