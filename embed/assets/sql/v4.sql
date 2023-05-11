-- Add new column to handle contest's allowance for scoreboard public view
ALTER TABLE contests ADD COLUMN scoreboard_view_status VARCHAR NOT NULL DEFAULT "public";
