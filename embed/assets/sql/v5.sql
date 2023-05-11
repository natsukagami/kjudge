BEGIN TRANSACTION;
    -- Add columns for the users table.
    ALTER TABLE users ADD COLUMN display_name VARCHAR NOT NULL DEFAULT "";
    -- Set the default display names to id.
    UPDATE users SET display_name = id;
    -- Add "organization" column
    ALTER TABLE users ADD COLUMN organization VARCHAR NOT NULL DEFAULT "";

    -- Add customization option
    ALTER TABLE config ADD COLUMN enable_user_customization INTEGER NOT NULL DEFAULT 1;
COMMIT;
