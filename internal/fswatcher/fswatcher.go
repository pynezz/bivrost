package fswatcher

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/fsnotify/fsnotify"
	"github.com/pynezz/bivrost/internal/util"
)

func Watch(file string, data chan<- string) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// Create new watcher.
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	linePos := 0 // Variable to keep track of the last line read position
	f, err := os.Open(file)
	if err != nil {
		log.Println("error opening file:", err)
	}

	scanner := bufio.NewScanner(f)

	// Wow, this is a really ugly piece of code. I'm sorry.
	go func() {
		util.PrintInfo("Watching file:" + file + "...")
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					util.PrintInfo("modified file:" + event.Name)
					// Open the file at each write event

					// Read newly added data
					// scanner := bufio.NewScanner(f)
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
					util.PrintError("error reading file:" + err.Error())
					return
				}
				log.Println("error:", err)
			}
		}

	}()

	// Add the file to be watched.
	err = watcher.Add(file)
	if err != nil {
		log.Fatal(err)
	}

	util.PrintInfo("Waiting for SIGINT or SIGTERM... Press Ctrl+C to exit.")
	// Wait for the signal to exit.
	<-c
	util.PrintInfo("Filewatcher: Cleaning up...")
}
