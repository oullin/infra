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

func parseDbSecrets(files DbSecretFiles, s *DbSecrets) error {
	var err error
	var username, password, dbname string

	if dbname, err = pkg.GetFileContent(files.dbName); err != nil {
		return fmt.Errorf("Error: Could not parse db name secret file: %v\n", err)
	}

	if username, err = pkg.GetFileContent(files.dbUser); err != nil {
		return fmt.Errorf("Error: Could not parse db user name secret file: %v\n", err)
	}

	if password, err = pkg.GetFileContent(files.dbPass); err != nil {
		return fmt.Errorf("Error: Could not parse db user password secret file: %v\n", err)
	}

	s.dbName = dbname
	s.dbUser = username
	s.dbPass = password

	return nil
}
