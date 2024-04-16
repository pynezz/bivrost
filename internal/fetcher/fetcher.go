/* fetcher is the package that fetches the data from the given location */

package fetcher

import (
	"bufio"
	"os"
)

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
