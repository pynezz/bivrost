CREATE TABLE nginx_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    time_local TEXT NOT NULL,
    remote_addr TEXT NOT NULL,
    remote_user TEXT NOT NULL,
    request TEXT NOT NULL,
    status INTEGER NOT NULL,
    body_bytes_sent INTEGER NOT NULL,
    request_time REAL NOT NULL,
    http_referrer TEXT,
    http_user_agent TEXT,
    request_body TEXT
);

/* Indexes. These are the columns that will be frequently searched and thus deserve an index for faster search.
The general guideline is to create indexes on columns that are:
 - Part of the WHERE, ORDER BY, or JOIN clauses in your queries.
 - Columns with high selectivity (unique or nearly unique values).
 - Columns that are frequently joined.
Indexes do have some drawbacks, such as slowing down inserts and updates. But we're expecting a lot more reads than writes. */
CREATE INDEX idx_remote_addr ON nginx_logs (remote_addr);   -- Indexing remote_addr is useful for finding requests from a specific IP
CREATE INDEX idx_status ON nginx_logs (status);             -- To be able to search by status code      (such as 302 to find all login redirects)
