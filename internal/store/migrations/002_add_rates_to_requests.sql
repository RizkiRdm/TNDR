ALTER TABLE requests ADD COLUMN prompt_rate REAL DEFAULT 0.0;
ALTER TABLE requests ADD COLUMN completion_rate REAL DEFAULT 0.0;
