package main

import (
	"flag"
	"log"
	"os"
	"os/signal"

	_ "git.nkagami.me/natsukagami/kjudge"
	"git.nkagami.me/natsukagami/kjudge/db"
	_ "git.nkagami.me/natsukagami/kjudge/models"
	"git.nkagami.me/natsukagami/kjudge/worker"
	"git.nkagami.me/natsukagami/kjudge/worker/isolate"
	"git.nkagami.me/natsukagami/kjudge/worker/raw"
)

var (
	dbfile  = flag.String("file", "kjudge.db", "Path to the database file")
	sandboxImpl = flag.String("sandbox", "isolate", "The sandbox implementation to be used (isolate, raw). If anything other than 'raw' is given, isolate is used.")
)

func main() {
	flag.Parse()

	db, err := db.New(*dbfile)
	if err != nil {
		log.Fatalf("%+v", err)
	}

	var sandbox worker.Sandbox
    if *sandboxImpl == "raw" {
        log.Println("'raw' sandbox selected. WE ARE NOT RESPONSIBLE FOR ANY BREAKAGE CAUSED BY FOREIGN CODE.")
        sandbox = &raw.Sandbox{}
    } else {
        sandbox = isolate.New()
    }

    // Start the queue
	queue := worker.Queue { Sandbox: sandbox, DB: db }
	go queue.Start()

    stop := make(chan os.Signal)
    signal.Notify(stop, os.Interrupt)

    log.Println("Starting kjudge. Press Ctrl+C to stop")
    <-stop

    log.Println("Shutting down")
}
