/* sqlite3 database */
-- Path: db/migrations/004_results.sql

--	synTrafficFields := []string{"description", "source", "first_timestamp", "last_timestamp", "count", "status", "recommendation"}
-- for detecting nmap scans
CREATE TABLE IF NOT EXISTS syntraffic (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    source TEXT,
    description TEXT,
    first_timestamp TIMESTAMP,
    last_timestamp TIMESTAMP,
    count INTEGER,
    status TEXT,
    recommendation TEXT
);

-- 	attackTypeLogFields := []string{"source", "description", "count", "severity", "threshold", "first_timestamp", "last_timestamp", "status", "recommendation", "request_path", "user_agent", "payload"}
CREATE TABLE IF NOT EXISTS attack_type (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    source TEXT,
    description TEXT,
    count INTEGER,
    severity TEXT,
    threshold TEXT,
    first_timestamp TIMESTAMP,
    last_timestamp TIMESTAMP,
    status TEXT,
    recommendation TEXT,
    request_path TEXT,
    user_agent TEXT,
    payload TEXT
);
