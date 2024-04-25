package database

import (
	"bufio"
	"fmt"
	"os"

	"github.com/pynezz/bivrost/internal/fsutil"
	"github.com/pynezz/bivrost/internal/util"
)

func (s *DataStore[T]) NewTestWriter(logfile string) (*bufio.Scanner, *os.File, error) {
	if logfile == "" {
		logfile = "./nginx_logentries_1k.log"
	}

	if logfile == "10k" {
		logfile = "./nginx_logentries_10k.log"
	}

	if logfile == "1k" {
		logfile = "./nginx_logentries_1k.log"
	}

	if !fsutil.FileExists(logfile) {
		return nil, nil, fmt.Errorf("file %s does not exist", logfile)
	}
	util.PrintInfo("Testing database write from " + logfile + "...")

	// logentries, err := os.Open("/home/siem/bluelogs/soc/standard.log")
	logentries, err := os.Open(logfile)
	if err != nil {
		return nil, nil, err
	}

	scanner := bufio.NewScanner(logentries)

	return scanner, logentries, nil
}
