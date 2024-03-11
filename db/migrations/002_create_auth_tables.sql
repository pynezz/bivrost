/* SQLite initialization script for the webauthn_auth and password_auth tables */

/* PATH: db/migrations/002_create_auth_tables.sql */

CREATE TABLE auth_methods (
    AuthMethodID INTEGER PRIMARY KEY,
    Description TEXT
);

/* Populate auth_methods with initial data */
INSERT INTO auth_methods (AuthMethodID, Description)
VALUES
    (1, 'Password'),
    (2, 'WebAuthn');

CREATE TABLE user_sessions (
    SessionID TEXT PRIMARY KEY,
    UserID INTEGER NOT NULL,
    Token TEXT NOT NULL, /* The user session is a JWT Token */
    FOREIGN KEY (UserID) REFERENCES users(UserID) ON DELETE CASCADE
);

CREATE TABLE webauthn_auth (
    CredentialID TEXT PRIMARY KEY,
    UserID INTEGER NOT NULL,
    PublicKey TEXT NOT NULL,
    UserHandle TEXT NOT NULL,
    SignatureCounter INTEGER NOT NULL,
    CreatedAt TEXT DEFAULT (datetime('now')),
    FOREIGN KEY (UserID) REFERENCES users(UserID) ON DELETE CASCADE
);

CREATE TABLE password_auth (
    UserID INTEGER PRIMARY KEY,
    Enabled BOOLEAN DEFAULT 1, -- SQLite uses 1 for TRUE
    PasswordHash TEXT NOT NULL,
    FOREIGN KEY (UserID) REFERENCES users(UserID) ON DELETE CASCADE
);
