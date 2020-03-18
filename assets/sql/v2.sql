-- Add new column to handle user's failed attempts in unweighted contests
ALTER TABLE problem_results ADD COLUMN failed_attempts INTEGER NOT NULL DEFAULT 0;
