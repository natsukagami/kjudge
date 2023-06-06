package perf_test

import (
	"fmt"
	"math/rand"
)

const bigInputCode = 
`#include <stdio.h>
int main(){
	int r = 0;
	while (char c = get_char()){
		if ('a' <= c && c <= 'b') r++;
		else break;
	}
	printf("%i", r);
}
`

// 50MB input to compare disk read time. O(1) memory.
// Problem: Given a string, print it's length
func BigInputProblem() *PerfTestSet {
	maxSize := 50000000 // 50MB
	return &PerfTestSet{
		Name:         "INPUT",
		ExpectedTime: 2000, // TODO
		TestGenerator: func(r *rand.Rand) *PerfTest {
			strSize := maxSize - r.Intn(10)
			input := make([]byte, strSize)
			for i := range input {
				input[i] = byte(r.Intn(26) + 'A')
			}
			return &PerfTest{
				Input:  input,
				Output: []byte(fmt.Sprintf("%v", strSize)),
			}
		},
		TestCode: []byte(bigInputCode),
	}
}
