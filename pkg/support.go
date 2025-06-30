package pkg

import (
	"fmt"
	"os"
	"strings"
)

func FilesExist(files []string) error {
	for _, file := range files {
		if err := FileExists(file); err != nil {
			return err
		}
	}

	return nil
}

func FileExists(path string) error {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("Error: File not found at: %s\n", path)
		} else {
			return fmt.Errorf("Error: Could not stat file %s: %v\n", path, err)
		}
	}

	return nil
}

func GetFileContent(path string) (string, error) {
	content, err := os.ReadFile(path)

	if err != nil {
		return "", fmt.Errorf("Error: Failed to read file content from %s: %v\n", path, err)
	}

	return strings.TrimSpace(
		string(content),
	), nil
}
