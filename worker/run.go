package worker

import (
	"database/sql"
	"log"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/natsukagami/kjudge/models"
	"github.com/pkg/errors"
)

// The filename of the "compare" binary.
const CompareFilename = "compare"

// RunContext is the context needed to run a test.
type RunContext struct {
	DB        *sqlx.Tx
	Sub       *models.Submission
	Problem   *models.Problem
	TestGroup *models.TestGroup
	Test      *models.Test
}

// TimeLimit returns the time limit of the context, in time.Duration.
func (r *RunContext) TimeLimit() time.Duration {
	if r.TestGroup.TimeLimit.Valid {
		return time.Duration(r.TestGroup.TimeLimit.Int64) * time.Millisecond
	}
	return time.Duration(r.Problem.TimeLimit) * time.Millisecond
}

// MemoryLimit returns the memory limit of the context, in Kilobytes.
func (r *RunContext) MemoryLimit() int {
	if r.TestGroup.MemoryLimit.Valid {
		return int(r.TestGroup.MemoryLimit.Int64)
	}
	return r.Problem.MemoryLimit
}

// RunnCommand returns the run command (command, args list) for the language.
func RunCommand(l models.Language) (string, []string, error) {
	switch l {
	case models.LanguageJava:
		return "/usr/bin/java", []string{"-Donline_judge=true", "-Dkjudge=true", "-Smx512M", "-Xss64M", "-cp", "code", "Main"}, nil
	case models.LanguagePy2:
		return "/usr/bin/python2", []string{"-S", "code"}, nil
	case models.LanguagePy3:
		return "/usr/bin/python3", []string{"-S", "code"}, nil
	case models.LanguageCpp, models.LanguageGo, models.LanguageRust, models.LanguagePas:
		return "code", nil, nil
	default:
		return "", nil, errors.New("unknown language")
	}
}

// CompiledSource returns the CompiledSource. Returns false when the submission hasn't been compiled.
// Returns nil if the submission failed to compile.
func (r *RunContext) CompiledSource() (bool, []byte) {
	if r.Sub.CompilerOutput == nil {
		return false, nil
	}
	return true, r.Sub.CompiledSource
}

// RunInput creates a SandboxInput for running the submission's source.
func (r *RunContext) RunInput(source []byte) (*SandboxInput, error) {
	command, args, err := RunCommand(r.Sub.Language)
	if err != nil {
		return nil, err
	}
	return &SandboxInput{
		Command:     command,
		Args:        args,
		Files:       nil,
		TimeLimit:   r.TimeLimit(),
		MemoryLimit: r.MemoryLimit(),

		CompiledSource: source,
		Input:          r.Test.Input,
	}, nil
}

// CompareInput creates a SandboxInput for running the comparator.
// Also returns whether we have diff-based or comparator-based input.
func (r *RunContext) CompareInput(submissionOutput []byte) (input *SandboxInput, useComparator bool, err error) {
	file, err := models.GetFileWithName(r.DB, r.Problem.ID, "compare")
	if errors.Is(err, sql.ErrNoRows) {
		// Use a simple diff
		return &SandboxInput{
			Command:     "/usr/bin/diff",
			Args:        []string{"-wqts", "output", "expected"},
			Files:       map[string][]byte{"output": submissionOutput, "expected": r.Test.Output},
			TimeLimit:   time.Second,
			MemoryLimit: 262144, // 256MBs
		}, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	// Use the given comparator.
	return &SandboxInput{
		Command:     "code",
		Args:        []string{"input", "expected", "output"},
		Files:       map[string][]byte{"input": r.Test.Input, "expected": r.Test.Output, "output": submissionOutput},
		TimeLimit:   20 * time.Second,
		MemoryLimit: (2 << 20), // 1 GB

		CompiledSource: file.Content,
	}, true, nil
}

// Run runs a RunContext.
func Run(sandbox Sandbox, r *RunContext) error {
	compiled, source := r.CompiledSource()
	if !compiled {
		// Add a compilation job and re-add ourselves.
		log.Printf("[WORKER] Submission %v not compiled, creating Compile job.\n", r.Sub.ID)
		return models.BatchInsertJobs(r.DB, models.NewJobCompile(r.Sub.ID), models.NewJobRun(r.Sub.ID, r.Test.ID))
	}
	if source == nil {
		log.Printf("[WORKER] Not running a submission that failed to compile.\n")
		return nil
	}

	log.Printf("[WORKER] Running submission %v on [test `%v`, group `%v`]\n", r.Sub.ID, r.Test.Name, r.TestGroup.Name)

	// First, use the sandbox to run the submission itself.
	input, err := r.RunInput(source)
	if err != nil {
		return err
	}
	output, err := sandbox.Run(input)
	if err != nil {
		return errors.WithStack(err)
	}
	result := parseSandboxOutput(output, r)

	if !output.Success {
		result.Verdict = "Runtime Error"
		if output.ErrorMessage != "" {
			result.Verdict = output.ErrorMessage
		}
		// If running the source did not succeed, we stop here and be happy with the test result.
		return result.Write(r.DB)
	}

	// Attempt to run the comparator
	input, useComparator, err := r.CompareInput(output.Stdout)
	if err != nil {
		return err
	}
	output, err = sandbox.Run(input)
	if err != nil {
		return err
	}
	if err := parseComparatorOutput(output, result, useComparator); err != nil {
		return err
	}

	log.Printf("[WORKER] Done running submission %v on [test `%v`, group `%v`]: %.1f (t = %v, m = %v)\n",
		r.Sub.ID, r.Test.Name, r.TestGroup.Name, result.Score, result.RunningTime, result.MemoryUsed)

	return result.Write(r.DB)
}

// Parse the comparator's output and reflect it into `result`.
func parseComparatorOutput(s *SandboxOutput, result *models.TestResult, useComparator bool) error {
	if useComparator {
		// Paste the comparator's output to result
		result.Verdict = strings.TrimSpace(string(s.Stderr))
		if result.Verdict == "" {
			result.Verdict = "Compare returns no output."
		}
		score, err := strconv.ParseFloat(strings.TrimSpace(string(s.Stdout)), 64)
		if err != nil {
			return errors.WithStack(err)
		}
		if math.IsInf(score, 0) || math.IsNaN(score) || score < 0 || score > 1 {
			return errors.Errorf("invalid comparator score %f", score)
		}
		result.Score = score
	} else {
		// Cute message from diff
		result.Verdict = strings.TrimSpace(string(s.Stdout))
		if result.Verdict == "" {
			result.Verdict = "Diff failed"
		}
		if !s.Success {
			result.Score = 0.0
		}
	}
	return nil
}

// Parse the sandbox output into a TestResult.
func parseSandboxOutput(s *SandboxOutput, r *RunContext) *models.TestResult {
	score := 1.0
	if !s.Success {
		score = 0.0
	}
	return &models.TestResult{
		MemoryUsed:   s.MemoryUsed,
		RunningTime:  int(s.RunningTime / time.Millisecond),
		Score:        score,
		SubmissionID: r.Sub.ID,
		TestID:       r.Test.ID,
		Verdict:      s.ErrorMessage,
	}
}
