-- Version of the schema
CREATE TABLE IF NOT EXISTS version (
	version VARCHAR NOT NULL
);

INSERT INTO version VALUES ("v1");

-- Contests 
CREATE TABLE contests (
	id INTEGER PRIMARY KEY NOT NULL,
	name VARCHAR NOT NULL,
	start_time VARCHAR NOT NULL,
	end_time VARCHAR NOT NULL, 
	contest_type VARCHAR NOT NULL
);

-- Problems
CREATE TABLE problems (
	id INTEGER PRIMARY KEY NOT NULL,
	contest_id INTEGER NOT NULL,
	name VARCHAR NOT NULL,
	display_name VARCHAR NOT NULL,
	time_limit INTEGER NOT NULL,
	memory_limit INTEGER NOT NULL,
	scoring_mode VARCHAR NOT NULL,
    penalty_policy VARCHAR NOT NULL,

	FOREIGN KEY(contest_id) REFERENCES contests(id) ON DELETE CASCADE
);

CREATE TABLE test_groups (
    id INTEGER PRIMARY KEY NOT NULL,
    problem_id INTEGER NOT NULL,
    name VARCHAR NOT NULL DEFAULT "main",
    time_limit INTEGER DEFAULT NULL,
    memory_limit INTEGER DEFAULT NULL,
    score REAL NOT NULL,
    scoring_mode VARCHAR NOT NULL,

    UNIQUE(problem_id, name),
    FOREIGN KEY(problem_id) REFERENCES problems(id) ON DELETE CASCADE
);

-- Tests
CREATE TABLE tests (
    id INTEGER PRIMARY KEY NOT NULL,
	test_group_id INTEGER NOT NULL,
	name VARCHAR NOT NULL,
	input BLOB NOT NULL,
	output BLOB NOT NULL,

    UNIQUE(test_group_id, name),
	FOREIGN KEY(test_group_id) REFERENCES test_groups(id) ON DELETE CASCADE
);

-- Users
CREATE TABLE users (
	id VARCHAR PRIMARY KEY NOT NULL,
	password VARCHAR NOT NULL
);

-- Submissions
CREATE TABLE submissions (
	id INTEGER PRIMARY KEY NOT NULL,
	problem_id INTEGER NOT NULL,
	user_id VARCHAR NOT NULL,
	submitted_at VARCHAR NOT NULL,
	-- Source file information
	language VARCHAR NOT NULL,
	source BLOB NOT NULL,
	-- Judge cache
	compiled_source BLOB DEFAULT NULL,
    compiler_output BLOB DEFAULT NULL,
	-- Results
	verdict VARCHAR NOT NULL DEFAULT "...",
	score   REAL DEFAULT NULL,
	penalty REAL DEFAULT NULL,
	
	FOREIGN KEY(problem_id) REFERENCES problems(id) ON DELETE CASCADE,
	FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Test results 
CREATE TABLE test_results (
  submission_id INTEGER NOT NULL,
  test_id INTEGER NOT NULL,
  verdict VARCHAR NOT NULL,
  score   REAL NOT NULL,
  running_time INTEGER NOT NULL,
  memory_used INTEGER NOT NULL
);
	
-- Problem results
CREATE TABLE problem_results (
	user_id VARCHAR NOT NULL,
	problem_id INTEGER NOT NULL,
	-- meaningful values
    solved INTEGER NOT NULL DEFAULT 0,
	score REAL NOT NULL DEFAULT 0,
	penalty REAL NOT NULL DEFAULT 0,
	best_submission_id INTEGER DEFAULT NULL,
	
	PRIMARY KEY (user_id, problem_id),
	FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
	FOREIGN KEY (problem_id) REFERENCES problems(id) ON DELETE CASCADE,
    FOREIGN KEY (best_submission_id) REFERENCES submissions(id) ON DELETE CASCADE
);

-- Internal jobs 
CREATE TABLE jobs (
  id INTEGER PRIMARY KEY NOT NULL,
  priority INTEGER NOT NULL,
  type VARCHAR NOT NULL,
  user_id INTEGER DEFAULT NULL,
  problem_id INTEGER DEFAULT NULL,
  submission_id INTEGER DEFAULT NULL,
  test_id INTEGER DEFAULT NULL,
  
  FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE,
  FOREIGN KEY(problem_id) REFERENCES problems(id) ON DELETE CASCADE,
  FOREIGN KEY(submission_id) REFERENCES submissions(id) ON DELETE CASCADE,
  FOREIGN KEY(test_id) REFERENCES tests(id) ON DELETE CASCADE
);

CREATE INDEX jobs_by_priority ON jobs (priority DESC, id ASC);
CREATE INDEX jobs_by_type ON jobs (type); -- Just to run SELECT count(id) FROM jobs GROUP BY type;

-- Files for use within a problem (mostly graders)
CREATE TABLE files (
    id INTEGER PRIMARY KEY NOT NULL,
    problem_id INTEGER NOT NULL,
    filename VARCHAR NOT NULL,
    content BLOB NOT NULL,

    FOREIGN KEY (problem_id) REFERENCES problems(id)
);
