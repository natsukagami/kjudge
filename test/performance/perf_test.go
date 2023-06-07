package performance

import (
	"fmt"
	"os"
	"testing"
)

func BenchmarkSandboxes(b *testing.B) {
	tmpDir, err := os.MkdirTemp(os.TempDir(), "kjudge_bench")
	if err != nil {
		b.Error(err)
		b.FailNow()
	}
	defer os.RemoveAll(tmpDir)

	for _, testset := range []*PerfTestSet{BigInputProblem(), SpawnTimeProblem()} {
		for _, sandboxName := range []string{"raw", "isolate"} {
			b.Run(fmt.Sprintf("%v %v", testset.Name, sandboxName),
				func(b *testing.B) {RunSingleTest(b, tmpDir, testset, sandboxName)})
		}
	}
}
