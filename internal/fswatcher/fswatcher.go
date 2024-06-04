package fswatcher

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/pynezz/bivrost/internal/util"
)

func Watch(file string, data chan<- string) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	directory := filepath.Dir(file)
	fileName := filepath.Base(file)

	// Create new watcher.
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	linePos := 0 // Variable to keep track of the last line read position
	cooldown := 0
	// var m sync.Mutex

	// Wow, this is a really ugly piece of code. I'm sorry.
	go func() {
		util.PrintInfo("Watching file:" + fileName + " in path " + directory)
		for {
			select {
			case event, ok := <-watcher.Events:
				fmt.Println("event triggered: " + event.String())
				fmt.Println("by: " + event.Op.String())
				if !ok {
					fmt.Println("NOOO")
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					// m.Lock()
					if cooldown == 0 {
						cooldown = 2
						// m.Unlock()
						// go cool(&cooldown, &m)
						util.PrintInfo("modified file:" + event.Name)

						f, err := os.Open(file)
						if err != nil {
							log.Println("error opening file:", err)
							continue
						}
						defer f.Close()

						reader := bufio.NewReader(f)

						// Skip already read lines
						// for i := 0; i < linePos; i++ {
						// 	_, _, err := reader.ReadLine()
						// 	if err != nil {
						// 		util.PrintError("encountered an error while reading line: " + err.Error())
						// 		break
						// 	}
						// }

						// Read newly added data
						for {
							line, _, err := reader.ReadLine()
							if err != nil {
								break
							}
							data <- string(line) // Send new data to channel
							// util.PrintSuccess("[" + fmt.Sprintf("%d", linePos) + "] Read " + string(line) + " from file and inserted into channel.")
							linePos++
						}
					} else {
						// m.Unlock()
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					util.PrintError("error reading file: " + err.Error())
					return
				}
				log.Println("error: ", err)
			}
		}
	}()

	// Add the file to be watched.
	watcher.Add(directory)

	util.PrintInfo("Waiting for SIGINT or SIGTERM... Press Ctrl+C to exit.")
	// Wait for the signal to exit.
	<-c
	util.PrintInfo("Filewatcher: Cleaning up...")
}

func cool(cooldown *int, m *sync.Mutex) {
	util.PrintDebug("[FILEWATCHER] 2 second cooldown...")
	time.Sleep(time.Second)
	m.Lock()
	*cooldown = 0
	m.Unlock()
	util.PrintDebug("[FILEWATCHER] Cooldown ended.")
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
