package pkg

import (
	"fmt"
	"os"
	"strings"
)

func Trim(seed string) string {
	return strings.TrimSpace(seed)
}

func FilesExist(files []string) error {
	for _, file := range files {
		if err := FileExists(file); err != nil {
			return err
		}
	}

	return nil
}

func FileExists(path string) error {
	info, err := os.Stat(path)

	if err != nil {
		return fmt.Errorf("Could not stat file %s: %v\n", path, err)
	}

	if os.IsNotExist(err) {
		return fmt.Errorf("Error: File not found at: %s\n", path)
	}

	if info.IsDir() {
		return fmt.Errorf("Error: %s is a directory\n", path)
	}

	return nil
}

func GetFileContent(path string) (string, error) {
	if err := FileExists(path); err != nil {
		return "", fmt.Errorf("Error: File not found: %s\n", path)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("Error: Failed to read file content from %s: %v\n", path, err)
	}

	return strings.TrimSpace(string(content)), nil
}
