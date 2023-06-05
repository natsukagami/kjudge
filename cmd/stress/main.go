// Compares provided sandboxes
package main

import (
	"flag"
	"log"

	"github.com/natsukagami/kjudge/db"
	"github.com/natsukagami/kjudge/worker"
)

var (
	dbfile       = flag.String("file", "kjudge-test.db", "Path to database file.")
	outputfile   = flag.String("output", "kjudge-test.json", "Path to output file.")
	sandboxImpls = flag.Args()
	batchCount   = flag.Int("count", 10, "Number of iterations.")
	batchSize    = flag.Int("count", 30, "Size of testset.")
)

func main() {
	flag.Parse()

	db, err := db.New(*dbfile)
	if err != nil {
		log.Fatalf("%+v", err)
	}
	defer db.Close()

	for i := 1; i <= *batchCount; i++ {
		log.Printf("Running testset number %v", i)
		for _, sbxName := range sandboxImpls {
			sandbox, err := worker.NewSandbox(sbxName)
			if err != nil {
				log.Printf("%v", err)
			}
		}
	}
}
