package models

import (
	"fmt"
	"strings"

	"git.nkagami.me/natsukagami/kjudge/db"
	"git.nkagami.me/natsukagami/kjudge/models/verify"
	"github.com/pkg/errors"
)

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

// WriteTests writes the given set of tests into the Database.
// If override is set, all tests in the test group gets deleted first.
func (r *TestGroup) WriteTests(db db.DBContext, tests []*Test, override bool) error {
	for _, test := range tests {
		test.TestGroupID = r.ID
		if err := test.Verify(); err != nil {
			return errors.Wrapf(err, "test `%s`", test.Name)
		}
	}
	if override {
		if _, err := db.Exec("DELETE FROM tests WHERE test_group_id = ?", r.ID); err != nil {
			return errors.WithStack(err)
		}
	}
	var (
		terms []string
		vars  []interface{}
	)
	for _, test := range tests {
		terms = append(terms, "(?, ?, ?, ?)")
		vars = append(vars, test.Name, test.TestGroupID, test.Input, test.Output)
	}
	res, err := db.Exec(fmt.Sprintf("INSERT INTO tests(name, test_group_id, input, output) VALUES %s", strings.Join(terms, ", ")), vars...)
	if err != nil {
		return errors.WithStack(err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return errors.WithStack(err)
	}
	for i, test := range tests {
		test.ID = int(id) - len(tests) + i + 1
	}
	return nil
}
