package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"path/filepath"

	_ "github.com/natsukagami/kjudge"
	"github.com/natsukagami/kjudge/db"
	_ "github.com/natsukagami/kjudge/models"
	"github.com/natsukagami/kjudge/server"
	"github.com/natsukagami/kjudge/worker"
)

var (
	dbfile      = flag.String("file", "kjudge.db", "Path to the database file.")
	sandboxImpl = flag.String("sandbox", "isolate", "The sandbox implementation to be used (isolate, raw). Defaults to isolate.")
	port        = flag.Int("port", 8088, "The port for the server to listen on.")
	verbose     = flag.Bool("verbose", false, "Log every http requests")

	httpsDir = flag.String("https", "", "Path to the directory where the HTTPS private key (kjudge.key) and certificate (kjudge.crt) is located. If omitted or empty, HTTPS is disabled.")
)

func main() {
	flag.Parse()

	db, err := db.New(*dbfile)
	if err != nil {
		log.Fatalf("%+v", err)
	}
	defer db.Close()

	sandbox, err := worker.NewSandbox(*sandboxImpl)
	if err != nil {
		log.Fatalf("%v", err)
	}

	opts := []server.Opt{}
	if *verbose {
		opts = append(opts, server.Verbose())
	}

	// Start the queue
	queue := worker.Queue{Sandbox: sandbox, DB: db}

	// Build the server
	server, err := server.New(db, opts...)
	if err != nil {
		log.Fatalf("%+v", err)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	log.Println("Starting kjudge. Press Ctrl+C to stop")

	go queue.Start()
	go startServer(server)

	received_signal := <-stop

	log.Printf("Shutting down on receiving %s", received_signal)
}

func startServer(server *server.Server) {
	var err error
	if *httpsDir == "" {
		// No HTTPS
		err = server.Start(*port)
	} else {
		// Start a HTTP server to host the root CA.
		if rootCAPort, ok := os.LookupEnv("ROOT_CA_PORT"); ok {
			go func() {
				if err := server.ServeHTTPRootCA(":"+rootCAPort, filepath.Join(*httpsDir, "root.pem")); err != nil {
					panic(err)
				}
			}()
		}
		err = server.StartWithSSL(*port, filepath.Join(*httpsDir, "kjudge.key"), filepath.Join(*httpsDir, "kjudge.crt"))
	}
	if err != nil {
		panic(err)
	}
}
