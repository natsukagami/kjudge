package models

import (
	"log"
	"os/exec"

	"github.com/jmoiron/sqlx"
	"github.com/natsukagami/kjudge/db"
	"github.com/natsukagami/kjudge/models/verify"
	"github.com/pkg/errors"
)

// Language represents the language of the submission.
// The available values depend on the machine the judge is run on.
type Language string

const (
	LanguageCpp  Language = "g++"
	LanguagePas  Language = "fpc"
	LanguageJava Language = "javac"
	LanguagePy2  Language = "python2"
	LanguagePy3  Language = "python3"
	LanguageGo   Language = "go"
	LanguageRust Language = "rustc"
)

const (
	VerdictCompileError = "Compile Error"
	VerdictScored       = "Scored"
	VerdictAccepted     = "Accepted"
	VerdictIsInQueue 	= "..."
)

var availableLanguages []string

// LanguageByExt returns a language based on the file extension.
func LanguageByExt(ext string) (Language, error) {
	switch ext {
	case ".cpp", ".cc":
		return LanguageCpp, nil
	case ".pas":
		return LanguagePas, nil
	case ".java":
		return LanguageJava, nil
	case ".py2":
		return LanguagePy2, nil
	case ".py3", ".py":
		return LanguagePy3, nil
	case ".go":
		return LanguageGo, nil
	case ".rs":
		return LanguageRust, nil
	default:
		return "", errors.New("unknown language")
	}
}

func init() {
	for _, l := range []Language{LanguageCpp, LanguagePas, LanguageJava, LanguagePy2, LanguagePy3, LanguageGo, LanguageRust} {
		ok := false
		for _, versionArg := range []string{"--version", "-version", "version", "-iW"} {
			if exec.Command(string(l), versionArg).Run() == nil {
				availableLanguages = append(availableLanguages, string(l))
				ok = true
				break
			}
		}
		if !ok {
			log.Printf("\"%s\" seems to be unavailable on the system. Declining all submissions with the language...", l)
		}
	}
}

func (l Language) verify() error {
	return verify.String(string(l), verify.Enum(availableLanguages...))
}

// Verify verifies Submission's content.
func (r *Submission) Verify() error {
	if (r.Penalty.Valid || r.Score.Valid) && (!r.Penalty.Valid || !r.Score.Valid) {
		return errors.New("penalty and score must be both null")
	}
	if r.Source == nil {
		return errors.New("source must not be null")
	}
	m := map[string]error{
		"Language": r.Language.verify(),
		"Verdict":  verify.StringNonEmpty(r.Verdict),
	}
	if r.Penalty.Valid {
		m["Penalty"] = verify.NullInt(r.Penalty, verify.IntMin(0))
		m["Score"] = verify.Float(r.Score.Float64, verify.FloatMin(0))
	}
	return verify.All(m)
}

// GetProblemsSubmissions returns the submissions that belong to a list of problems.
func GetProblemsSubmissions(db db.DBContext, problemID ...int) ([]*Submission, error) {
	if len(problemID) == 0 {
		return nil, nil
	}
	query, args, err := sqlx.In("SELECT * FROM submissions WHERE problem_id IN (?)"+querySubmissionOrderBy, problemID)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	var result []*Submission
	if err := db.Select(&result, query, args...); err != nil {
		return nil, errors.WithStack(err)
	}
	return result, nil
}

// GetUserProblemSubmissions gets all submissions belonging to an user on a (list of) problem(s).
func GetUserProblemSubmissions(db db.DBContext, userID string, problemID ...int) ([]*Submission, error) {
	if len(problemID) == 0 {
		return nil, nil
	}
	query, args, err := sqlx.In("SELECT * FROM submissions WHERE problem_id IN (?) AND user_id = ?"+querySubmissionOrderBy, problemID, userID)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	var result []*Submission
	if err := db.Select(&result, query, args...); err != nil {
		return nil, errors.WithStack(err)
	}
	return result, nil
}
