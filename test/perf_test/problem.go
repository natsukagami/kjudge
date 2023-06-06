// Package perf_test provides performance testing
package perf_test

import "math/rand"

type PerfTest struct {
	Input  []byte
	Output []byte
}

type PerfTestSet struct {
	Name          string
	ExpectedTime  int // Expected running time of each test in ms
	TestGenerator func(*rand.Rand) *PerfTest
	TestCode      []byte // Solution to tested problem
}

func GenerateContest() {

}
