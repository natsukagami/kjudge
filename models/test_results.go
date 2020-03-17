package models

import "git.nkagami.me/natsukagami/kjudge/models/verify"

// Verify verifies that the TestResult is a legit one.
func (r *TestResult) Verify() error {
	return verify.All(map[string]error{
		"MemoryUsed":  verify.IntMin(0)(r.MemoryUsed),
		"RunningTime": verify.IntMin(0)(r.RunningTime),
		"Score":       verify.Float(r.Score, verify.FloatRange(0, 1)),
		"Verdict":     verify.StringNonEmpty(r.Verdict),
	})
}
