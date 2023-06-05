// Package performance provides performance testing
package performance

import (
	"fmt"
	"math/rand"

	"github.com/natsukagami/kjudge/db"
	"github.com/natsukagami/kjudge/models"
)

func generateProblem(db db.DBContext, contestID int, displayName string, name string) *models.Problem {
	return &models.Problem{
		ContestID:                 contestID,
		DisplayName:               displayName,
		ID:                        0,       // Auto generate ID
		MaxSubmissionsCount:       0,       // No limit
		MemoryLimit:               1 << 20, // 1GB
		Name:                      name,
		PenaltyPolicy:             "none",
		ScoringMode:               "best",
		SecondsBetweenSubmissions: 0,
		TimeLimit:                 5,
	}
}

type PerfTest struct {
	Input  []byte
	Output []byte
}

type PerfTestSet struct {
	Name          string
	TestCount     int
	TestGenerator func(*rand.Rand) *PerfTest
}

// Generates O(1), multitest problem to compare sandbox spawn time.
// The problem used is input one number, then output double of that number
func spawnTimeProblem() *PerfTestSet {
	return &PerfTestSet{
		Name:      "SPAWN",
		TestCount: 1000,
		TestGenerator: func(r *rand.Rand) *PerfTest {
			value := r.Int31() / 2
			return &PerfTest{
				Input:  []byte(fmt.Sprintf("%v", value)),
				Output: []byte(fmt.Sprintf("%v", value*2)),
			}
		},
	}
}

func GenerateContest() {

}
