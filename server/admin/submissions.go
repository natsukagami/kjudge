package admin

import (
	"net/http"

	"git.nkagami.me/natsukagami/kjudge/db"
	"git.nkagami.me/natsukagami/kjudge/models"
	"github.com/labstack/echo/v4"
)

// ContestWithProblems is a contest with embedded problems.
type ContestWithProblems struct {
	*models.Contest
	Problems []*models.Problem
}

// SubmissionsCtx is a context for *listing submissions*.
type SubmissionsCtx struct {
	User  *models.User // or
	Users map[string]*models.User

	Problem     *models.Problem         // or
	ProblemsMap map[int]*models.Problem // AND one of
	Contest     *ContestWithProblems
	Contests    []*ContestWithProblems

	Submissions []*models.Submission
}

// SubmissionsAll returns a context of all submissions.
func SubmissionsAll(db db.DBContext) (*SubmissionsCtx, error) {
	var s SubmissionsCtx
	if err := s.fillUsers(db); err != nil {
		return nil, err
	}
	if err := s.fillContests(db); err != nil {
		return nil, err
	}
	ss, err := models.GetAllSubmissions(db)
	if err != nil {
		return nil, err
	}
	s.Submissions = ss
	return &s, nil
}

// SubmissionsBy applies a filter to the submission list:
// - either with an user (if u != nil)
// - or with a contest... (if c != nil)
// - or with a single problem (if p != nil, overrides contest).
func SubmissionsBy(db db.DBContext, u *models.User, c *models.Contest, p *models.Problem) (*SubmissionsCtx, error) {
	var s SubmissionsCtx
	var err error
	if u != nil {
		s.User = u
		if p != nil {
			s.Problem = p
			s.Submissions, err = models.GetUserProblemSubmissions(db, u.ID, p.ID)
		} else if c != nil {
			if err := s.fillContest(db, c); err != nil {
				return nil, err
			}
			var ps []int
			for _, p := range s.Contest.Problems {
				ps = append(ps, p.ID)
			}
			s.Submissions, err = models.GetUserProblemSubmissions(db, u.ID, ps...)
		} else {
			if err := s.fillContests(db); err != nil {
				return nil, err
			}
			s.Submissions, err = models.GetUserSubmissions(db, u.ID)
		}
	} else {
		users, err := models.GetAllUsers(db)
		if err != nil {
			return nil, err
		}
		s.Users = make(map[string]*models.User)
		for _, u := range users {
			s.Users[u.ID] = u
		}
		if p != nil {
			s.Problem = p
			s.Submissions, err = models.GetProblemSubmissions(db, p.ID)
		} else if c != nil {
			if err := s.fillContest(db, c); err != nil {
				return nil, err
			}
			var ps []int
			for _, p := range s.Contest.Problems {
				ps = append(ps, p.ID)
			}
			s.Submissions, err = models.GetProblemsSubmissions(db, ps...)
		} else {
			return SubmissionsAll(db)
		}
	}
	return &s, err
}

func (s *SubmissionsCtx) fillUsers(db db.DBContext) error {
	us, err := models.GetAllUsers(db)
	if err != nil {
		return err
	}
	users := make(map[string]*models.User)
	for _, u := range us {
		users[u.ID] = u
	}
	s.Users = users
	return nil
}

func (s *SubmissionsCtx) fillContest(db db.DBContext, c *models.Contest) error {
	ps, err := models.GetContestProblems(db, c.ID)
	if err != nil {
		return err
	}
	problems := make(map[int]*models.Problem)
	for _, p := range ps {
		problems[p.ID] = p
	}
	s.Contest = &ContestWithProblems{Contest: c, Problems: ps}
	s.ProblemsMap = problems
	return nil
}

func (s *SubmissionsCtx) fillContests(db db.DBContext) error {
	cs, err := models.GetAllContests(db)
	if err != nil {
		return err
	}
	contestsMap := make(map[int]*ContestWithProblems)
	problemsMap := make(map[int]*models.Problem)
	for _, c := range cs {
		contestsMap[c.ID] = &ContestWithProblems{Contest: c}
	}

	ps, err := models.GetAllProblems(db)
	if err != nil {
		return err
	}
	for _, p := range ps {
		problemsMap[p.ID] = p
		if c, ok := contestsMap[p.ContestID]; ok {
			c.Problems = append(c.Problems, p)
		}
	}

	s.Contests = nil
	for _, c := range contestsMap {
		s.Contests = append(s.Contests, c)
	}

	s.ProblemsMap = problemsMap

	return nil
}

// SubmissionsGet implements GET /admin/submissions
func (g *Group) SubmissionsGet(c echo.Context) error {
	subs, err := SubmissionsAll(g.db)
	if err != nil {
		return err
	}
	return c.Render(http.StatusOK, "admin/submissions", subs)
}
