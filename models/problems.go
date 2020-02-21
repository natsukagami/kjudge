package models

import "git.nkagami.me/natsukagami/kjudge/models/verify"

// ScoringMode dictates how the best submission is chosen.
// There are:
// - Best: The submission with the highest score is chosen. If on a tie, choose the one with largest penalty.
// - Once: The first (successfully compiled) submission is the best one.
// - Last: The last submission is the best one.
// - Decay: The last submission is the best one. The score is modified by the number of submissions before it (0.1 * count), and the
// time passed (0.7 * time passed in %), to a minimum of 0.3 times the original.
type ScoringMode string

// Defined values for ScoringMode.
const (
	ScoringModeBest  ScoringMode = "best"
	ScoringModeOnce  ScoringMode = "once"
	ScoringModeLast  ScoringMode = "last"
	ScoringModeDecay ScoringMode = "decay"
)

func (s ScoringMode) verify() error {
	return verify.String(string(s), verify.Enum(string(ScoringModeBest), string(ScoringModeOnce), string(ScoringModeLast), string(ScoringModeDecay)))
}

// PenaltyPolicy dictates how the penalty is calculated.
// There are:
// - None: no penalty at all (IOI - like).
// - SubmitTime: minutes passed from the start of the contest, of the submission time.
// - ICPC: SubmitTime + 20 * (# of past submissions)
type PenaltyPolicy string

// Defined values for PenaltyPolicy.
const (
	PenaltyPolicyNone       PenaltyPolicy = "none"
	PenaltyPolicySubmitTime PenaltyPolicy = "submit_time"
	PenaltyPolicyICPC       PenaltyPolicy = "icpc"
)

func (p PenaltyPolicy) verify() error {
	return verify.String(string(p), verify.Enum(string(PenaltyPolicyNone), string(PenaltyPolicySubmitTime), string(PenaltyPolicyICPC)))
}

// Verify verifies a Problem's content.
func (r *Problem) Verify() error {
	return verify.All(map[string]error{
		"DisplayName":   verify.Names(r.DisplayName),
		"Name":          verify.Names(r.Name),
		"ScoringMode":   r.ScoringMode.verify(),
		"PenaltyPolicy": r.PenaltyPolicy.verify(),
		"TimeLimit":     verify.IntPositive(r.TimeLimit),
		"MemoryLimit":   verify.IntPositive(r.MemoryLimit),
	})
}
