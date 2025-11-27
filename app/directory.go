package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

func createDirectory(path string) bool {
	_, err := os.Stat(path)

	if err == nil {
		return true
	}

	if errors.Is(err, fs.ErrNotExist) {
		err := os.Mkdir(path, 0755)
		if err != nil {
			fmt.Sprintf("Failed to create directory: %s", err)
			return false
		}
	}

	fmt.Sprintf("Failed to find directory: %s", err)
	return false
}

func getFileContent(name string) ([]byte, error) {
	path := filepath.Join(args.directory, name)

	file, err := os.ReadFile(path)

	if err != nil {
		return nil, err
	}

	return file, nil
}
