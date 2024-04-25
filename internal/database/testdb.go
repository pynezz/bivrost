package database

import (
	"bufio"
	"fmt"
	"os"

	"github.com/pynezz/bivrost/internal/fetcher"
	"github.com/pynezz/bivrost/internal/fsutil"
	"github.com/pynezz/bivrost/internal/util"
	"gorm.io/gorm"
)

func (s *DataStore[T]) TestWrite(logfile string) error {
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
		return fmt.Errorf("file %s does not exist", logfile)
	}
	util.PrintInfo("Testing database write from " + logfile + "...")

	timestamp := util.UnixNanoTimestamp()
	var finalTime int64

	// logentries, err := os.Open("/home/siem/bluelogs/soc/standard.log")
	logentries, err := os.Open(logfile)
	if err != nil {
		return err
	}
	defer logentries.Close()

	st, err := NewDataStore[fetcher.NginxLog]("logs", gorm.Config{})
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(logentries)
	for scanner.Scan() {
		log, err := fetcher.ParseNginxLog(scanner.Text())
		if err != nil {
			if err.Error() == "log is an environment variable" {
				continue
			} else {
				return err
			}
		}

		if err := st.InsertLog(log); err != nil {
			return err
		}
	}
	finalTime = util.UnixNanoTimestamp()
	elapsed := finalTime - timestamp
	util.PrintSuccess(fmt.Sprintf("Created 10k logs\n > %d µsec", elapsed/1000))
	util.PrintSuccess(fmt.Sprintf(" > %d msec", elapsed/1000000))
	util.PrintSuccess(fmt.Sprintf(" > %d sec", elapsed/1000000000))
	util.PrintSuccess(fmt.Sprintf(" > %d min", elapsed/1000000000/60))

	return nil
}
