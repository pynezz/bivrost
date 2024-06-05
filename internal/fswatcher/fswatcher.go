package fswatcher

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/pynezz/pynezzentials/ansi"
)

type LogReader struct {
	fileName  string
	parentDir string

	linesRead int
	offset    int64 // Track the offset instead of lines read

	file   *os.File
	reader *bufio.Reader
}

const cacheFile string = ".bivrost_fswatcher.cache"

var cache = make(map[string]int64) // In-memory cache for file offsets

func Watch(file string, data chan<- string) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	directory := filepath.Dir(file)
	fileName := filepath.Base(file)

	readCache()

	logReader := &LogReader{
		fileName:  fileName,
		parentDir: directory,
		offset:    cache[fileName],
		// linesRead: readCacheFile(fileName),
	}

	// Create new watcher.
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	// linePos := 0 // Variable to keep track of the last line read position
	// cooldown := 0
	// var m sync.Mutex

	// Wow, this is a really ugly piece of code. I'm sorry.
	go func() {
		ansi.PrintInfo("Watching file:" + fileName + " in path " + directory)
		for {
			select {
			case event, ok := <-watcher.Events:
				fmt.Println("event triggered: " + event.String())
				fmt.Println("event name: " + event.Name)
				if !ok {
					fmt.Println("NOOO")
					return
				}
				if (event.Op&fsnotify.Write == fsnotify.Write) &&
					filepath.Base(event.Name) == logReader.fileName {

					// if cooldown == 0 {
					// 	cooldown = 2
					// m.Unlock()
					// go cool(&cooldown, &m)
					ansi.PrintInfo("modified file:" + event.Name)

					f, err := os.Open(file)
					if err != nil {
						log.Println("error opening file:", err)
						continue
					}

					logReader.file = f
					logReader.reader = bufio.NewReader(f)
					_, err = logReader.file.Seek(logReader.offset, io.SeekStart)
					if err != nil {
						ansi.PrintError("error seeking file: " + err.Error())
						continue
					}
					// Read newly added data
					lines, newOffset, err := logReader.continueRead()
					if err != nil {
						if err != io.EOF {
							ansi.PrintError("error occurred while continuing read: " + err.Error())
						}
						break
					}
					// logReader.linesRead = lastLine
					logReader.offset = newOffset

					for _, line := range lines {
						data <- line // Send new data to channel
						ansi.PrintSuccess("[" + fmt.Sprintf("%d", logReader.linesRead) + "] Read " + line + " from file and inserted into channel.")
					}

					cache[logReader.fileName] = logReader.offset
					writeCache()
					// } else {
					// 	// m.Unlock()
					// }
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					ansi.PrintError("error reading file: " + err.Error())
					return
				}
				log.Println("error: ", err)
			}
		}
	}()

	// Add the file to be watched.
	watcher.Add(directory)

	ansi.PrintInfo("Waiting for SIGINT or SIGTERM... Press Ctrl+C to exit.")
	// Wait for the signal to exit.
	<-c
	ansi.PrintInfo("Filewatcher: Cleaning up...")
	writeCache()
}

func cool(cooldown *int, m *sync.Mutex) {
	ansi.PrintDebug("[FILEWATCHER] 2 second cooldown...")
	time.Sleep(time.Second)
	m.Lock()
	*cooldown = 0
	m.Unlock()
	ansi.PrintDebug("[FILEWATCHER] Cooldown ended.")
}

// If the passed file is a file, we want to watch the parent directory, and look for the file with event.Name()
func ensureParentDir(path string) (string, bool) {
	fd, err := os.Stat(path)
	if err != nil {
		fmt.Printf("an error occured checking the file %s", fd.Name())
	}

	// If the wat
	if fd.IsDir() {
		return path, true
	}

	parent := filepath.Join(path, "..")
	pfd, err := os.Stat(parent)
	if err != nil {
		fmt.Printf("an error occured checking the file %s", fd.Name())
	}

	if pfd.IsDir() {
		return path, false
	}

	return path, false
}

// func (l *LogReader) continueRead() (lines []string, lastLine int, err error) {
// 	err = io.EOF // default error

// 	scanner := bufio.NewScanner(l.reader)
// 	currentLine := l.linesRead
// 	for scanner.Scan() {
// 		currentLine++
// 		if currentLine > l.linesRead {
// 			lines = append(lines, scanner.Text())
// 		}
// 	}

// 	lastLine = currentLine
// 	writeCacheFile(l.fileName, lastLine)
// 	return lines, lastLine, err
// }

func (l *LogReader) continueRead() (lines []string, newOffset int64, err error) {
	scanner := bufio.NewScanner(l.reader)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	newOffset, err = l.file.Seek(0, io.SeekCurrent)
	return lines, newOffset, err
}

/*
Cached file layout:

<filename> <linesread>
*/
// func writeCacheFile(fileName string, linesRead int) {
// 	var file *os.File
// 	var err error

// 	if !fsansi.FileExists(cacheFile) {
// 		file, err = fsansi.CreateFile(cacheFile)
// 	} else {
// 		file, err = fsansi.GetFile(cacheFile)
// 	}
// 	if err != nil {
// 		ansi.PrintError("error occured while reading or creating cache: " + err.Error())
// 		return
// 	}

// 	sc := bufio.NewScanner(file)

// 	for sc.Scan() {
// 		line := strings.Split(sc.Text(), " ")
// 		if line[0] == fileName {
// 			write := fmt.Sprintf("%s%d", line[0], linesRead)

// 		}
// 	}

// }

func readCache() {
	file, err := os.Open(cacheFile)
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		ansi.PrintError("error opening cache file: " + err.Error())
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, " ")
		if len(parts) == 2 {
			var offset int64
			fmt.Sscanf(parts[1], "%d", &offset)
			cache[parts[0]] = offset
		}
	}
}

func writeCache() {
	file, err := os.OpenFile(cacheFile, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		ansi.PrintError("error opening cache file: " + err.Error())
		return
	}
	defer file.Close()

	file.Seek(0, 0)
	file.Truncate(0)
	for fileName, offset := range cache {
		file.WriteString(fmt.Sprintf("%s %d\n", fileName, offset))
	}
}

func writeCacheFile(fileName string, linesRead int) {
	file, err := os.OpenFile(cacheFile, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		ansi.PrintError("error opening cache file: " + err.Error())
		return
	}
	defer file.Close()

	fmt.Println("writing cache file...")
	var cacheData []string
	scanner := bufio.NewScanner(file)
	found := false
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, " ")
		if parts[0] == fileName {
			cacheData = append(cacheData, fmt.Sprintf("%s %d", fileName, linesRead))
			found = true
			fmt.Printf("appending data: %s\n", cacheData)
		} else {
			cacheData = append(cacheData, line)
		}
	}
	if !found {
		cacheData = append(cacheData, fmt.Sprintf("%s %d", fileName, linesRead))
	}

	file.Seek(0, 0)
	file.Truncate(0)
	for _, line := range cacheData {
		file.WriteString(line + "\n")
	}
}

/*
Cached file layout:

<filename> <linesread>
*/
// func readCacheFile(fileName string) int {
// 	contents, err := fsansi.GetFileContent(fileName)
// 	if err != nil {
// 		ansi.PrintError("error reading cached filecontents: " + err.Error())
// 		return 0
// 	}

// 	sc := bufio.NewScanner()
// 	sc.Scan()

// 	reader := bufio.NewReader(cacheFile)
// 	for {
// 		if line := reader.ReadLine(); strings.Split(line)[0] == fileName {
// 			// We've found the cached read file

// 		}
// 	}

// }
func readCacheFile(fileName string) int {
	file, err := os.Open(cacheFile)
	if err != nil {
		if os.IsNotExist(err) {
			return 0
		}
		ansi.PrintError("error opening cache file: " + err.Error())
		return 0
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, " ")
		if parts[0] == fileName {
			var linesRead int
			fmt.Sscanf(parts[1], "%d", &linesRead)
			return linesRead
		}
	}

	return 0
}
