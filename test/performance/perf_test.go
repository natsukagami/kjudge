package performance

import (
	"fmt"
	"log"
	"testing"
)

var testList = []*PerfTestSet{BigInputProblem(), SpawnTimeProblem()}
var sandboxList = []string{"raw", "isolate"}

func BenchmarkSandboxes(b *testing.B) {
	log.Println("creating test DB")

	ctx, err := NewBenchmarkContext(b.TempDir())
	if err != nil {
		b.Fatal(err)
	}
	defer ctx.Close()

	for _, testset := range testList {
		log.Printf("creating problem %v", testset.Name)
		if err := ctx.writeProblem(testset); err != nil {
			b.Fatal(err)
		}
	}

	for _, testset := range testList {
		for _, sandboxName := range sandboxList {
			testName := fmt.Sprintf("%v %v", testset.Name, sandboxName)
			log.Printf("running %v", testName)
			b.Run(testName, func(b *testing.B) { RunSingleTest(b, ctx, testset, sandboxName) })
		}
	}
}
