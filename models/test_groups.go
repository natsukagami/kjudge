package models

import "git.nkagami.me/natsukagami/kjudge/models/verify"

// TestScoringMode determines how the score of each test in the group adds up to the total of the group.
// The schemes are:
// - Sum: Each test has an equal weight and the score of the group is (sum of test scores / # of tests) * (group score)
// - Min: The score of the group = min(score of each test) * (group score)
// - Product: Score of the group = product(score of each test) * (group score)
type TestScoringMode string

// All possible values of TestScoringMode.
const (
	TestScoringModeSum     TestScoringMode = "sum"
	TestScoringModeMin     TestScoringMode = "min"
	TestScoringModeProduct TestScoringMode = "product"
)

func (t TestScoringMode) verify() error {
	return verify.String(string(t), verify.Enum(string(TestScoringModeSum), string(TestScoringModeMin), string(TestScoringModeProduct)))
}

// Verify verifies TestGroup's content.
func (r *TestGroup) Verify() error {
	return verify.All(map[string]error{
		"TimeLimit":   verify.NullInt(r.TimeLimit, verify.IntPositive),
		"MemoryLimit": verify.NullInt(r.MemoryLimit, verify.IntPositive),
		"Name":        verify.Names(r.Name),
	})
}
