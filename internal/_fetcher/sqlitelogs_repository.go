package fetcher

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"

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

// Generic repository interface implementation (from nginxlogs.go)
type SQLiteRepository[T any] struct {
	db     *sql.DB
	table  string
	fields []string

	insert func() (*sql.Stmt, error)
}

//	func NewSQLiteLogs(db *sql.DB) *SQLiteLogsRepository {
//		return &SQLiteLogsRepository{
//			db: db,
//		}
//	}

// NewSQLiteRepository creates a new SQLiteRepository.
func NewSQLiteRepository[T any](db *sql.DB, table string, fields []string) *SQLiteRepository[T] {
	return &SQLiteRepository[T]{
		db:     db,
		table:  table,
		fields: fields,

		insert: func() (*sql.Stmt, error) {
			insert := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", table, strings.Join(fields, ", "), strings.TrimRight(strings.Repeat("?, ", len(fields)), ", "))
			return db.Prepare(insert)
		},
	}
}

func (r *SQLiteRepository[T]) Create(args ...any) error {
	// Implement the Create method
	stmt, err := r.insert()
	if err != nil {
		return err
	}

	res, err := stmt.Exec(args)
	if err != nil {
		return err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return err
	}

	fmt.Printf("inserted entry to %s with id: %d\n", r.table, id)

	return nil
}

// func (r *SQLiteLogsRepository) Create(log NginxLog) (*NginxLog, error) {
// 	res, err := r.db.Exec(
// 		`INSERT INTO nginx_logs (
// 			time_local, remote_addr, remote_user, request, status, body_bytes_sent, request_time, http_referrer, http_user_agent, request_body
// 		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
// 		log.TimeLocal, log.RemoteAddr, log.RemoteUser,
// 		log.Request, log.Status, log.BodyBytesSent,
// 		log.RequestTime, log.HttpReferer,
// 		log.HttpUserAgent, log.RequestBody,
// 	)

// 	if err != nil {
// 		return nil, err
// 	}

// 	id, err := res.LastInsertId()
// 	if err != nil {
// 		return nil, err
// 	}

// 	log.ID = id

// 	return &log, nil
// }

func (r *SQLiteRepository[T]) All() ([]T, error) {
	rows, err := r.db.Query(`SELECT * FROM ?`, r.table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	logs := make([]T, 128)

	for rows.Next() {
		var log T
		err := scanRowIntoStruct(rows, &log)
		if err != nil {
			return nil, err
		}

		logs = append(logs, log)
	}

	return logs, nil
}

// func (r *SQLiteLogsRepository) All() (*NginxLogsList, error) {
// 	rows, err := r.db.Query(`SELECT * FROM nginx_logs`)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer rows.Close()

// 	var logs NginxLogsList // Just a slice of logs ([]NginxLog)
// 	for rows.Next() {
// 		var log NginxLog
// 		err := rows.Scan(
// 			&log.ID, &log.TimeLocal, &log.RemoteAddr, &log.RemoteUser,
// 			&log.Request, &log.Status, &log.BodyBytesSent,
// 			&log.RequestTime, &log.HttpReferer, &log.HttpUserAgent,
// 			&log.RequestBody,
// 		)
// 		if err != nil {
// 			return nil, err
// 		}

// 		logs.Logs = append(logs.Logs, log)
// 	}

// 	return &logs, nil
// }

func (r *SQLiteRepository[T]) GetByID(id int64) (*T, error) {

	return nil, nil
}

func (r *SQLiteRepository[T]) Update(id int64, updated T) (*T, error) {

	return nil, nil
}

func (r *SQLiteRepository[T]) Delete(id int64) error {

	return nil
}

func setIDField[T any](entry *T, id int64) {
	v := reflect.ValueOf(entry).Elem()
	idField := v.FieldByName("ID")
	if idField.IsValid() && idField.CanSet() {
		idField.SetInt(id)
	}
}

func scanRowIntoStruct(rows *sql.Rows, dest interface{}) error {
	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))

	for i := range values {
		valuePtrs[i] = &values[i]
	}

	if err := rows.Scan(valuePtrs...); err != nil {
		return err
	}

	rv := reflect.ValueOf(dest).Elem()
	for i := 0; i < rv.NumField(); i++ {
		field := rv.Field(i)
		if field.CanSet() {
			field.Set(reflect.ValueOf(values[i]))
		}
	}

	return nil
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

CREATE INDEX IF NOT EXISTS idx_remote_addr ON nginx_logs (remote_addr);
CREATE INDEX IF NOT EXISTS idx_status ON nginx_logs (status);
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

// func (n *NginxLogRepository) Create(log NginxLog) (*NginxLog, error) {
// 	res, err := n.db.Exec(
// 		`INSERT INTO nginx_logs (
// 			time_local, remote_addr, remote_user, request, status, body_bytes_sent, request_time, http_referrer, http_user_agent, request_body
// 		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
// 		log.TimeLocal, log.RemoteAddr, log.RemoteUser,
// 		log.Request, log.Status, log.BodyBytesSent,
// 		log.RequestTime, log.HttpReferer,
// 		log.HttpUserAgent, log.RequestBody,
// 	)

// 	if err != nil {
// 		return nil, err
// 	}

// 	id, err := res.LastInsertId()
// 	if err != nil {
// 		return nil, err
// 	}

// 	log.ID = id

// 	return &log, nil
// }

// func (r *ThreatTypeRepository) Create(log ThreatTypeLog, table string) (*ThreatTypeLog, error) {
// 	query := fmt.Sprintf("INSERT INTO %s (source, description, time_local, user_agent, payload) VALUES (?, ?, ?, ?, ?)", table)
// 	res, err := r.db.Exec(query, log.Source, log.Description, log.TimeLocal, log.UserAgent, log.Payload)

// 	if err != nil {
// 		return nil, err
// 	}

// 	id, err := res.LastInsertId()
// 	if err != nil {
// 		return nil, err
// 	}

// 	log.ID = id

// 	return &log, nil

// }

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
