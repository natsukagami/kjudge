package main

import (
	"flag"
	"log"
	"os"
	"os/signal"

	_ "git.nkagami.me/natsukagami/kjudge"
	"git.nkagami.me/natsukagami/kjudge/db"
	_ "git.nkagami.me/natsukagami/kjudge/models"
	"git.nkagami.me/natsukagami/kjudge/server"
	"git.nkagami.me/natsukagami/kjudge/worker"
	"git.nkagami.me/natsukagami/kjudge/worker/isolate"
	"git.nkagami.me/natsukagami/kjudge/worker/raw"
)

var (
	dbfile      = flag.String("file", "kjudge.db", "Path to the database file")
	sandboxImpl = flag.String("sandbox", "isolate", "The sandbox implementation to be used (isolate, raw). If anything other than 'raw' is given, isolate is used.")
	port        = flag.Int("port", 8088, "The port for the server to listen on")
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
	queue := worker.Queue{Sandbox: sandbox, DB: db}

	// Build the server
	server, err := server.New(db)
	if err != nil {
		log.Fatalf("%+v", err)
	}

	stop := make(chan os.Signal)
	signal.Notify(stop, os.Interrupt)

	log.Println("Starting kjudge. Press Ctrl+C to stop")

	go queue.Start()
	go server.Start(*port)

	<-stop

	log.Println("Shutting down")
}
