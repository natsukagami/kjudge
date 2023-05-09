-- Add two columns, "max_submissions_count" for the limit of total number 
-- of submissions; "seconds_between_submissions" for the amount of wait 
-- time between two submissions.

ALTER TABLE problems
    ADD COLUMN max_submissions_count INT NOT NULL DEFAULT 0;

ALTER TABLE problems
    ADD COLUMN seconds_between_submissions INT NOT NULL DEFAULT 0;
