package admin

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/natsukagami/kjudge/db"
	"github.com/natsukagami/kjudge/models"
)

// JobsCtx is a context for rendering jobs.
type JobsCtx struct {
	Jobs []*models.Job

	Tests      map[int]*models.Test
	TestGroups map[int]*models.TestGroup
	Problems   map[int]*models.Problem
}

// Collect a job context.
func getJobsCtx(db db.DBContext) (*JobsCtx, error) {
	jobs, err := models.GetAllJobs(db)
	if err != nil {
		return nil, err
	}
	var testIDs, tgIDs, problemIDs []int
	// collect tests
	for _, job := range jobs {
		if job.TestID.Valid {
			testIDs = append(testIDs, int(job.TestID.Int64))
		}
	}
	tests, err := models.CollectTestsByID(db, testIDs...)
	if err != nil {
		return nil, err
	}
	// collect test groups
	for _, test := range tests {
		tgIDs = append(tgIDs, test.TestGroupID)
	}
	testGroups, err := models.CollectTestGroupsByID(db, tgIDs...)
	if err != nil {
		return nil, err
	}
	// collect problems
	for _, tg := range testGroups {
		problemIDs = append(problemIDs, tg.ProblemID)
	}
	problems, err := models.CollectProblemsByID(db, problemIDs...)
	if err != nil {
		return nil, err
	}
	return &JobsCtx{
		Jobs: jobs,

		Tests:      tests,
		TestGroups: testGroups,
		Problems:   problems,
	}, nil
}

// Render renders the context.
func (j *JobsCtx) Render(c echo.Context) error {
	return c.Render(http.StatusOK, "admin/jobs", j)
}

// JobsGet implements GET "/admin/jobs".
func (g *Group) JobsGet(c echo.Context) error {
	ctx, err := getJobsCtx(g.db)
	if err != nil {
		return err
	}
	return ctx.Render(c)
}
