// filesystem handling
package fs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pynezz/bivrost/internal/util"
)

// FileExists checks if a file exists
func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

// DirExists checks if a directory exists
func DirExists(dirname string) bool {
	info, err := os.Stat(dirname)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

// GetFiles returns a list of files in a directory
func GetFiles(dirname string) ([]string, error) {
	files := []string{}
	err := filepath.Walk(dirname, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

// GetDirs returns a list of directories in a directory
func GetDirs(dirname string) ([]string, error) {
	dirs := []string{}
	err := filepath.Walk(dirname, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			dirs = append(dirs, path)
		}
		return nil
	})
	return dirs, err
}

// GetFilesWithExtension returns a list of files in a directory with a specific extension
func GetFilesWithExtension(dirname string, extension string) ([]string, error) {
	files := []string{}
	err := filepath.Walk(dirname, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && strings.HasSuffix(info.Name(), extension) {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func GetFile(filename string) (*os.File, error) {
	file := filename
	if !FileExists(file) {
		errMsg := fmt.Sprintf("File %s does not exist", file)
		return nil, util.Errorf(errMsg)
	}
	return os.Open(file)
}

func CreateFile(filename string) (*os.File, error) {
	return os.Create(filename)
}

// TODO: Should probably account for different filetypes
func GetFileContent(filename string) (string, error) {
	content, err := os.ReadFile(filename)
	return string(content), err
}
