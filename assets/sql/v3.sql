-- Add column "created_at" for Job
ALTER TABLE jobs ADD COLUMN created_at DATETIME NOT NULL DEFAULT '2020-03-29 21:49:17';
