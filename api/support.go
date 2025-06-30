package api

import (
	"fmt"
	"github.com/oullin/infra/pkg"
)

func assertSecretFiles(f DbSecretFiles) error {
	files := []string{f.dbUser, f.dbName, f.dbPass}

	for _, file := range files {
		if err := pkg.FileExists(file); err != nil {
			return fmt.Errorf("The following secret file does not exist: %s\n", file)
		}
	}

	return nil
}

func parseDbSecrets(s *DbSecrets) error {
	var err error

	if s.dbUser, err = pkg.GetFileContent(s.dbUser); err != nil {
		return fmt.Errorf("Error: Could not parse db name secret file: %v\n", err)
	}

	if s.dbName, err = pkg.GetFileContent(s.dbName); err != nil {
		return fmt.Errorf("Error: Could not parse db user name secret file: %v\n", err)
	}

	if s.dbPass, err = pkg.GetFileContent(s.dbPass); err != nil {
		return fmt.Errorf("Error: Could not parse db user password secret file: %v\n", err)
	}

	return nil
}
