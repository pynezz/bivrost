/* PATH: db/migrations/001_create_users_table.sql */

/* SQLite initialization script for the users table */
CREATE TABLE users (
    UserID INTEGER PRIMARY KEY AUTOINCREMENT,
    DisplayName TEXT UNIQUE NOT NULL,
    CreatedAt TEXT DEFAULT (datetime('now')),
    UpdatedAt TEXT DEFAULT (datetime('now')),
    LastLogin TEXT,
    Role TEXT CHECK(Role IN ('admin', 'user')) DEFAULT 'user',
    FirstName TEXT,
    ProfileImageURL TEXT,
    SessionId TEXT,
    AuthMethodID INTEGER   /* This is a foreign key to the auth_methods table,
                              but we need to add it later due to the auth tables being created after this one */
);
