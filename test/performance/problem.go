// Package performance provides performance testing
package performance

import (
	"github.com/natsukagami/kjudge/db"
	"github.com/natsukagami/kjudge/models"
)

func generateProblem(db db.DBContext, contestID int, displayName string, name string) *models.Problem{
	return &models.Problem {
		ContestID:					contestID,
		DisplayName:				displayName,
		ID:							0,			// Auto generate ID
		MaxSubmissionsCount:		0,            // No limit
		MemoryLimit:				1 << 20,      // 1GB
		Name:						name,
		PenaltyPolicy:				"none",
		ScoringMode:				"best",
		SecondsBetweenSubmissions:	0,
		TimeLimit:					5,
	}
}

// Generates O(1), multitest problem to compare sandbox spawn time
func spawnTimeProblem(db db.DBContext, contestID int) error {
	generateProblem()
}

func GenerateContest() {

}
