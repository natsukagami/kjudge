package main

import (
	"flag"
	"log"
	"os"
	"os/signal"

	_ "github.com/natsukagami/kjudge"
	"github.com/natsukagami/kjudge/db"
	_ "github.com/natsukagami/kjudge/models"
	"github.com/natsukagami/kjudge/server"
	"github.com/natsukagami/kjudge/worker"
	"github.com/natsukagami/kjudge/worker/isolate"
	"github.com/natsukagami/kjudge/worker/raw"
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

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	log.Println("Starting kjudge. Press Ctrl+C to stop")

	go queue.Start()
	go func() {
		if err := server.Start(*port); err != nil {
			panic(err)
		}
	}()

	<-stop

	log.Println("Shutting down")
}
