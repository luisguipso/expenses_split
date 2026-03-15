ALTER TABLE households ADD COLUMN split_mode VARCHAR(20) NOT NULL DEFAULT 'salary';
ALTER TABLE household_members ADD COLUMN split_percentage INT NOT NULL DEFAULT 0;
