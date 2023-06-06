package perf_test

import "math/rand"

const bigInputCode = `#include <stdio.h>
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
		CapTime:      5000,
		Count:    10,
		Generator: func(r *rand.Rand) []byte {
			strSize := maxSize - r.Intn(10)
			input := make([]byte, strSize)
			for i := range input {
				input[i] = byte(r.Intn(26) + 'A')
			}
			return input
		},
		Solution: []byte(bigInputCode),
	}
}
