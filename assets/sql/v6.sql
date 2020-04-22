BEGIN TRANSACTION;

-- An "Announcements" table.
CREATE TABLE announcements (
    id INTEGER NOT NULL PRIMARY KEY,
    contest_id INTEGER NOT NULL,
    problem_id INTEGER,
    content BLOB NOT NULL,
    created_at DATETIME NOT NULL,

    FOREIGN KEY (contest_id) REFERENCES contests(id) ON DELETE CASCADE,
    FOREIGN KEY (problem_id) REFERENCES problems(id) ON DELETE CASCADE
);
CREATE INDEX announcements_by_contest ON announcements(contest_id ASC, id DESC);

-- Clarifications table.
CREATE TABLE clarifications (
    id INTEGER NOT NULL PRIMARY KEY,
    user_id VARCHAR NOT NULL,
    contest_id INTEGER NOT NULL,
    problem_id INTEGER, 

    content BLOB NOT NULL,
    updated_at DATETIME NOT NULL, -- Updated when responded

    response BLOB,

    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (contest_id) REFERENCES contests(id) ON DELETE CASCADE,
    FOREIGN KEY (problem_id) REFERENCES problems(id) ON DELETE CASCADE
);

CREATE INDEX clarifications_by_user ON clarifications(contest_id ASC, user_id ASC, id DESC);

COMMIT;
