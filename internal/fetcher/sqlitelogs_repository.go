package fetcher

import (
	"database/sql"
	"errors"

	"github.com/pynezz/bivrost/internal/util"
)

var (
	ErrDuplicate    = errors.New("record already exists")
	ErrNotExists    = errors.New("row not exists")
	ErrUpdateFailed = errors.New("update failed")
	ErrDeleteFailed = errors.New("delete failed")
)

const (
	TableNginxLogs    = "nginx_logs"
	TableSshLogs      = "ssh_logs"
	TableAuthLogs     = "auth_logs"
	TableErrorLogs    = "error_logs"
	TableTLSErrorLogs = "tls_error_logs"
)

type SQLiteLogsRepository struct {
	db *sql.DB
}

func NewSQLiteLogs(db *sql.DB) *SQLiteLogsRepository {
	return &SQLiteLogsRepository{
		db: db,
	}
}

func (r *SQLiteLogsRepository) Migrate() error {
	query := `
CREATE TABLE IF NOT EXISTS nginx_logs (
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

CREATE INDEX IF NOT EXISTS idx_time_local ON nginx_logs (time_local);
CREATE INDEX IF NOT EXISTS idx_remote_addr ON nginx_logs (remote_addr);
CREATE INDEX IF NOT EXISTS idx_status ON nginx_logs (status);
CREATE INDEX IF NOT EXISTS idx_request_time ON nginx_logs (request_time);
CREATE INDEX IF NOT EXISTS idx_http_path ON nginx_logs (request);
    `

	// CREATE TABLE IF NOT EXISTS ssh_logs (
	// 	id INTEGER PRIMARY KEY AUTOINCREMENT,
	// 	time_local TEXT NOT NULL,
	// 	remote_addr TEXT NOT NULL,
	// 	remote_port INTEGER NOT NULL,
	// 	user TEXT NOT NULL,
	// 	status TEXT NOT NULL,
	// 	command TEXT NOT NULL
	// );

	// CREATE INDEX idx_time_local_ssh ON ssh_logs (time_local);
	// CREATE INDEX idx_remote_addr_ssh ON ssh_logs (remote_addr);
	// CREATE INDEX idx_status_ssh ON ssh_logs (status);
	// CREATE INDEX idx_user_ssh ON ssh_logs (user);

	// CREATE TABLE IF NOT EXISTS auth_logs (
	// 	id INTEGER PRIMARY KEY AUTOINCREMENT,
	// 	time_local TEXT NOT NULL,
	// 	remote_addr TEXT NOT NULL,
	// 	remote_port INTEGER NOT NULL,
	// 	user TEXT NOT NULL,
	// 	status TEXT NOT NULL,
	// 	command TEXT NOT NULL
	// );

	// CREATE INDEX idx_time_local_auth ON auth_logs (time_local);
	// CREATE INDEX idx_remote_addr_auth ON auth_logs (remote_addr);

	// CREATE TABLE IF NOT EXISTS error_logs (
	// 	id INTEGER PRIMARY KEY AUTOINCREMENT,
	// 	time_local TEXT NOT NULL,
	// 	remote_addr TEXT NOT NULL,
	// 	remote_port INTEGER NOT NULL,
	// 	status TEXT NOT NULL,
	// 	message TEXT NOT NULL
	// );

	// CREATE INDEX idx_time_local_error ON error_logs (time_local);
	// CREATE INDEX idx_remote_addr_error ON error_logs (remote_addr);

	// CREATE TABLE IF NOT EXISTS tls_error_logs (
	// 	id INTEGER PRIMARY KEY AUTOINCREMENT,
	// 	time_local TEXT NOT NULL,
	// 	remote_addr TEXT NOT NULL,
	// 	remote_port INTEGER NOT NULL,
	// 	status TEXT NOT NULL,
	// 	message TEXT NOT NULL
	// );

	// CREATE INDEX idx_time_local_tls_error ON tls_error_logs (time_local);
	// CREATE INDEX idx_remote_addr_tls_error ON tls_error_logs (remote_addr);

	// CREATE TABLE IF NOT EXISTS command_logs (
	// 	id INTEGER PRIMARY KEY AUTOINCREMENT,
	// 	time_local TEXT NOT NULL,
	// 	remote_addr TEXT,
	// 	user TEXT NOT NULL,
	// 	command TEXT NOT NULL,
	// 	status TEXT NOT NULL,
	// 	message TEXT NOT NULL,
	// );

	// CREATE INDEX idx_time_local_command ON command_logs (time_local);
	// CREATE INDEX idx_user_command ON command_logs (user);

	_, err := r.db.Exec(query)
	return err
}

func (r *SQLiteLogsRepository) Create(log NginxLog) (*NginxLog, error) {
	res, err := r.db.Exec(
		`INSERT INTO nginx_logs (
			time_local, remote_addr, remote_user, request, status, body_bytes_sent, request_time, http_referrer, http_user_agent, request_body
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		log.TimeLocal, log.RemoteAddr, log.RemoteUser,
		log.Request, log.Status, log.BodyBytesSent,
		log.RequestTime, log.HttpReferer,
		log.HttpUserAgent, log.RequestBody,
	)

	if err != nil {
		return nil, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	log.ID = id

	return &log, nil

}

func (r *SQLiteLogsRepository) All() (*NginxLogsList, error) {
	rows, err := r.db.Query(`SELECT * FROM nginx_logs`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs NginxLogsList // Just a slice of logs ([]NginxLog)
	for rows.Next() {
		var log NginxLog
		err := rows.Scan(
			&log.ID, &log.TimeLocal, &log.RemoteAddr, &log.RemoteUser,
			&log.Request, &log.Status, &log.BodyBytesSent,
			&log.RequestTime, &log.HttpReferer, &log.HttpUserAgent,
			&log.RequestBody,
		)
		if err != nil {
			return nil, err
		}

		logs.Logs = append(logs.Logs, log)
	}

	return &logs, nil
}

func (r *SQLiteLogsRepository) GetByIP(ip string) (NginxLogsList, error) {
	var logs NginxLogsList                                                        // Just a slice of logs ([]NginxLog)
	rows, err := r.db.Query(`SELECT * FROM nginx_logs WHERE remote_addr = ?`, ip) // ERROR: Query error: no such table: nginx_logs
	if err != nil {
		util.PrintError("Query error: " + err.Error())
		return logs, err
	}

	for rows.Next() {
		var log NginxLog

		err := rows.Scan(
			&log.ID, &log.TimeLocal, &log.RemoteAddr, &log.RemoteUser,
			&log.Request, &log.Status, &log.BodyBytesSent,
			&log.RequestTime, &log.HttpReferer, &log.HttpUserAgent,
			&log.RequestBody,
		)
		if err != nil {
			util.PrintError("Scan error: " + err.Error())
			return logs, err
		}

		logs.Logs = append(logs.Logs, log)
	}

	return logs, nil
}

func (r *SQLiteLogsRepository) Update(id int64, updated NginxLog) (*NginxLog, error) {
	res, err := r.db.Exec(
		`UPDATE nginx_logs SET
			time_local=?, remote_addr=?, remote_user=?, request=?, status=?, body_bytes_sent=?, request_time=?, http_referrer=?, http_user_agent=?, request_body=?
		WHERE id = ?`,
		updated.TimeLocal, updated.RemoteAddr, updated.RemoteUser,
		updated.Request, updated.Status, updated.BodyBytesSent,
		updated.RequestTime, updated.HttpReferer,
		updated.HttpUserAgent, updated.RequestBody,
		id,
	)

	if err != nil {
		return nil, err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}

	if rows == 0 {
		return nil, ErrUpdateFailed
	}

	updated.ID = id

	return &updated, nil
}

func (r *SQLiteLogsRepository) Delete(id int64) error {
	res, err := r.db.Exec(`DELETE FROM nginx_logs WHERE id = ?`, id)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrDeleteFailed
	}

	return nil
}
