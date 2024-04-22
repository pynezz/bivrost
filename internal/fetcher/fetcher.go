/* fetcher is the package that fetches the data from the given location */

package fetcher

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"database/sql"

	"github.com/pynezz/bivrost/internal/util"
)

const (
	LogsDB = "logs"
)

var dbs = []string{
	"logs",
}

// FetchFS fetches data from the file system on the given path
// Path must be absolute (e.g. /path/to/file)
func FetchFS(path string) (string, error) {
	var data string

	// Open the file for reading
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		data += scanner.Text() + "\n"
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return data, nil
}

const (
	nginx_log_test_001 = `{"time_local":"22/Apr/2024:17:56:07 +0000","remote_addr":"43.163.232.152","remote_user":"","request":"GET /viwwwsogou?op=8&query=%E7%A8%8F%E5%BB%BA%09%E9%BE%90%E1%B7%A2 HTTP/1.1","status": "400","body_bytes_sent":"248","request_time":"0.000","http_referrer":"","http_user_agent":"Mozilla/5.0 (Windows NT 6.1; Trident/7.0; rv:11.0) like Gecko","request_body":""}`
	nginx_log_test_002 = `{"time_local":"22/Apr/2024:16:53:00 +0000","remote_addr":"91.90.40.176","remote_user":"","request":"HEAD /(/302.php HTTP/1.1","status": "404","body_bytes_sent":"0","request_time":"0.037","http_referrer":"","http_user_agent":"DirBuster-1.0-RC1 (http://www.owasp.org/index.php/Category:OWASP_DirBuster_Project)","request_body":""}`
	nginx_log_test_003 = `{"time_local":"22/Apr/2024:13:39:49 +0000","remote_addr":"91.90.40.176","remote_user":"","request":"POST /login HTTP/1.1","status": "302","body_bytes_sent":"0","request_time":"0.010","http_referrer":"http://164.92.132.240/","http_user_agent":"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36","request_body":"username=admin&password=password_1234"}`
)

// ReadDB reads the data from the given database
// Enter the database name
func ReadDB(database string) (*sql.DB, error) {
	// Let's start by removing the database extension for fault tolerance
	// (eg. "logs.db" is passed instead of "logs")
	if !finddb(stripExt(database)) { // Then we add it back
		return nil, fmt.Errorf("database not found")
	}

	db, err := sql.Open("sqlite3", database)
	if err != nil {
		return nil, err
	}

	nginxlogRepository := NewSQLiteLogs(db)

	testIP := "91.90.40.176"

	readLogs, err := nginxlogRepository.GetByIP(testIP)
	if err != nil {
		util.PrintError("error fetching log by IP: " + err.Error())
	}
	if len(readLogs.Logs) > 0 {
		util.PrintSuccess("Logs fetched by IP:" + string(len(readLogs.Logs)))
		for _, log := range readLogs.Logs {
			util.PrintSuccess("Log ID: " + string(log.ID) + "\nLog timestamp: " + log.TimeLocal)
		}
	} else {
		util.PrintError("No logs found for IP: " + testIP)
	}

	// Let's create a new log
	log1, err := parseNginxLog(nginx_log_test_001)
	if err != nil {
		return nil, err
	}

	log2, err := parseNginxLog(nginx_log_test_002)
	if err != nil {
		return nil, err
	}

	log3, err := parseNginxLog(nginx_log_test_003)
	if err != nil {
		return nil, err
	}

	nginxlogRepository.Create(log1)
	nginxlogRepository.Create(log2)
	nginxlogRepository.Create(log3)

	return db, nil
}

/*
log_format standard_json escape=json

	'{'
	  '"time_local":"$time_local",'
	  '"remote_addr":"$remote_addr",'
	  '"remote_user":"$remote_user",'
	  '"request":"$request",'
	  '"status": "$status",'
	  '"body_bytes_sent":"$body_bytes_sent",'
	  '"request_time":"$request_time",'
	  '"http_referrer":"$http_referer",'
	  '"http_user_agent":"$http_user_agent",'
	  '"request_body":"$request_body"'
	'}';
*/
func parseNginxLog(log string) (NginxLog, error) { // Returning a copy for performance reasons
	// Remove the enclosing curly braces from the log
	// log = strings.TrimPrefix(log, "{")
	// log = strings.TrimSuffix(log, "}")

	print("Log to parse: " + log + "\n")

	var nginxLog NginxLog
	err := json.NewDecoder(strings.NewReader(log)).Decode(&nginxLog)
	if err != nil {
		return NginxLog{}, err
	}

	return nginxLog, nil
}

func finddb(database string) bool {
	for _, db := range dbs {
		if db == database {
			return true
		}
	}
	return false
}

func stripExt(s string) string {

	if len(strings.Split(s, ".")) > 0 {
		util.PrintItalic("Extension found in the database name. Stripping the extension: " + strings.Split(s, ".")[0])
		return strings.Split(s, ".")[0]
	}
	util.PrintItalic("No extension found in the database name. Returning the original string: " + s)
	return s
}
