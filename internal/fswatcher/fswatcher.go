package fswatcher

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
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

	// Wow, this is a really ugly piece of code. I'm sorry.
	go func() {
		util.PrintInfo("Watching file:" + fileName + " in path " + directory)
		for {
			select {
			case event, ok := <-watcher.Events:
				fmt.Println("event triggered: " + event.String())
				fmt.Println("by: " + event.Op.String())
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write && cooldown == 0 {
					cooldown = 1
					go cool(&cooldown)
					util.PrintInfo("modified file:" + event.Name)
					// Open the file at each write event

					f, err := os.Open(file)
					if err != nil {
						log.Println("error opening file:", err)
					}

					scanner := bufio.NewScanner(f)

					// Read newly added data
					for i := 0; i < linePos; i++ {
						scanner.Scan() // Skip the lines that have already been read
					}

					for scanner.Scan() {
						data <- scanner.Text() // Send new data to channel
						util.PrintSuccess("[" + fmt.Sprintf("%d", linePos) + "] Read " + scanner.Text() + " from file and inserted into channel.")
						linePos++
					}
					f.Close()
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

func cool(i *int) {
	time.Sleep(time.Second)
	*i = 0
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
