package database

import (
	"bufio"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/pynezz/bivrost/internal/database/models"
)

func ReadNginxLogs(scanner *bufio.Scanner, lines chan<- string, wg *sync.WaitGroup) {
	readLines := 0
	fmt.Print("Read lines: ")
	for scanner.Scan() {
		readLines++
		// Print the readLines count by overwriting the previous line:
		fmt.Printf("\r%d", readLines)
		lines <- scanner.Text()
	}
	fmt.Println()

	close(lines)
}

func ParseNginxLog(log string) (models.NginxLog, error) { // Returning a copy for performance reasons
	// Remove the enclosing curly braces from the log
	// log = strings.TrimPrefix(log, "{")
	// log = strings.TrimSuffix(log, "}")
	if log[0] != '{' && log[len(log)-1] != '}' {
		return models.NginxLog{}, EnvironError // Skip the log
	}

	// print("Log to parse: " + log + "\n")

	var nginxLog models.NginxLog
	err := json.NewDecoder(strings.NewReader(log)).Decode(&nginxLog)
	if err != nil {
		return models.NginxLog{}, err
	}

	return nginxLog, nil
}

// ParseBufferedNginxLog parses a channel of log lines and sends the parsed logs to another channel
func ParseBufferedNginxLog(lines <-chan string, logs chan<- models.NginxLog, wg *sync.WaitGroup) {
	count := 0
	for line := range lines {
		log, err := ParseNginxLog(line)
		if err != nil {
			fmt.Println("Failed to parse log:", line, err)
			continue
		}
		logs <- log
		count++
	}
	fmt.Print("Total logs parsed:")
	fmt.Printf("\r%d\n", count)
	// wg.Done()
	close(logs)
}
