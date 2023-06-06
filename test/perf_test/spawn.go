package perf_test

import (
	"fmt"
	"math/rand"
)

const spawnTimeCode = `#include <stdio.h>
int main(){
	int a; scanf("%i", &a);
	printf("%i", a*2);
}
`

// O(1) problem to compare sandbox spawn time.
// Problem: Input one number, then output the double of that number
func SpawnTimeProblem() *PerfTestSet {
	// maxValue * 2 must not cause integer overflow
	maxValue := 1 << 30
	return &PerfTestSet{
		Name:         "SPAWN",
		ExpectedTime: 1,
		CapTime:      100,
		TestGenerator: func(r *rand.Rand) []byte {
			value := r.Intn(maxValue)
			return []byte(fmt.Sprintf("%v", value))
		},
		TestCode: []byte(spawnTimeCode), // TODO
	}
}
