package perf_test

import (
	"fmt"
	"math/rand"
)
const spawnTimeCode =
`#include <stdio.h>
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
		TestGenerator: func(r *rand.Rand) *PerfTest {
			value := r.Intn(maxValue)
			return &PerfTest{
				Input:  []byte(fmt.Sprintf("%v", value)),
				Output: []byte(fmt.Sprintf("%v", value*2)),
			}
		},
		TestCode: []byte(spawnTimeCode), // TODO
	}
}
