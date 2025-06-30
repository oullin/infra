package api

import (
	"fmt"
	"github.com/oullin/infra/pkg"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func NewAPIDeployment(request DeploymentRequest) Deployment {
	apiDir := strings.TrimSpace(request.ApiDir)
	appDir := strings.TrimSpace(filepath.Base(apiDir))
	secretsDir := strings.TrimSpace(request.SecretsDir)

	projectRoot := strings.TrimSpace(filepath.Dir(apiDir) + appDir)

	dbSecretFile := strings.TrimSpace(filepath.Join(secretsDir, "postgres_db"))
	userSecretFile := strings.TrimSpace(filepath.Join(secretsDir, "postgres_user"))
	passwordSecretFile := strings.TrimSpace(filepath.Join(secretsDir, "postgres_password"))

	return Deployment{
		apiDir:      apiDir,
		secretsDir:  secretsDir,
		projectRoot: projectRoot,
		commands: Commands{
			Production: "build:prod",
		},
		apiDBSecrets: DbSecrets{
			dbSecretFile:          dbSecretFile,
			dbSecretContent:       "",
			userSecretFile:        userSecretFile,
			userSecretContent:     "",
			passwordSecretFile:    passwordSecretFile,
			passwordSecretContent: "",
		},
	}
}

func (d *Deployment) Run() error {
	files := []string{
		d.apiDir,
		d.secretsDir,
		d.projectRoot,
		d.apiDBSecrets.dbSecretFile,
		d.apiDBSecrets.userSecretFile,
		d.apiDBSecrets.passwordSecretFile,
	}

	if err := pkg.FilesExist(files); err != nil {
		return err
	}

	if err := d.ReadSecrets(); err != nil {
		return err
	}

	return nil
}

func (d *Deployment) ReadSecrets() error {
	dbSecretFile := d.apiDBSecrets.dbSecretFile
	userSecretFile := d.apiDBSecrets.userSecretFile
	passwordSecretFile := d.apiDBSecrets.passwordSecretFile

	// ---
	if content, err := pkg.GetFileContent(userSecretFile); err != nil {
		return fmt.Errorf("issue reading db user secret file: %v", err)
	} else {
		d.apiDBSecrets.userSecretContent = content
	}

	if content, err := pkg.GetFileContent(dbSecretFile); err != nil {
		return fmt.Errorf("issue reading db name secret file: %v", err)
	} else {
		d.apiDBSecrets.dbSecretContent = content
	}

	if content, err := pkg.GetFileContent(passwordSecretFile); err != nil {
		return fmt.Errorf("issue reading db password secret file: %v", err)
	} else {
		d.apiDBSecrets.passwordSecretContent = content
	}

	return nil
}

func (d *Deployment) Export() {
	projectRoot := d.projectRoot
	dbUser := d.apiDBSecrets.userSecretContent
	dbPassword := d.apiDBSecrets.passwordSecretContent
	dbName := d.apiDBSecrets.dbSecretContent

	makeArgs := []string{
		"-C",
		projectRoot,
		"build:prod",
		fmt.Sprintf("POSTGRES_USER_SECRET_PATH=%s", d.apiDBSecrets.userSecretFile),
		fmt.Sprintf("POSTGRES_PASSWORD_SECRET_PATH=%s", d.apiDBSecrets.passwordSecretFile),
		fmt.Sprintf("POSTGRES_DB_SECRET_PATH=%s", d.apiDBSecrets.dbSecretFile),
		fmt.Sprintf("ENV_DB_USER_NAME=%s", dbUser),
		fmt.Sprintf("ENV_DB_USER_PASSWORD=%s", d.apiDBSecrets.passwordSecretContent),
		fmt.Sprintf("ENV_DB_DATABASE_NAME=%s", dbName),
		fmt.Sprintf("ENV_DB_USER_NAME=%s", dbUser),
		fmt.Sprintf("ENV_DB_USER_PASSWORD=%s", dbPassword),
		fmt.Sprintf("ENV_DB_DATABASE_NAME=%s", dbName),
	}

	cmd := exec.Command("make", makeArgs...)

	// Pass the parent environment to the child process.
	cmd.Env = os.Environ()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("\nError: 'make -C %s build:prod' command failed: %v\n", projectRoot, err)
		os.Exit(1)
	}
}

func Handle() {
	// --- Configuration ---
	//secretsDir := strings.TrimSpace("/home/gocanto/.oullin/secrets")
	//apiDir := strings.TrimSpace("/home/gocanto/Sites/oullin/api")
	secretsDir := strings.TrimSpace("/Users/gus/.oullin/secrets")
	apiDir := strings.TrimSpace("/Users/gus/Sites/oullin/api")

	// Resolves to /home/gocanto/Sites/oullin
	projectRoot := strings.TrimSpace(filepath.Dir(apiDir) + "/api")

	// --- 1. Pre-flight Checks ---
	fmt.Println("--> [1/3] Verifying secret files...")

	userSecretFile := filepath.Join(secretsDir, "postgres_user")
	passwordSecretFile := filepath.Join(secretsDir, "postgres_password")
	dbSecretFile := filepath.Join(secretsDir, "postgres_db")

	checkFile(userSecretFile)
	checkFile(passwordSecretFile)
	checkFile(dbSecretFile)

	fmt.Println("--> Secret files verified successfully.")

	// --- 2. Read Secret Contents ---
	fmt.Println("--> [2/3] Reading secret contents from files...")

	dbUser := readSecretContent(userSecretFile)
	dbPassword := readSecretContent(passwordSecretFile)
	dbName := readSecretContent(dbSecretFile)

	fmt.Println("--> Secret contents loaded.")

	// --- 3. Deployment Execution ---
	fmt.Printf("--> [3/3] Launching Docker Compose services via 'make' from root (%s)...\n", projectRoot)

	// Construct the 'make' command with all secrets passed as arguments.
	// This makes them available as variables within the Makefile.
	makeArgs := []string{
		"-C",
		projectRoot,
		"build:prod",
		fmt.Sprintf("POSTGRES_USER_SECRET_PATH=%s", userSecretFile),
		fmt.Sprintf("POSTGRES_PASSWORD_SECRET_PATH=%s", passwordSecretFile),
		fmt.Sprintf("POSTGRES_DB_SECRET_PATH=%s", dbSecretFile),
		fmt.Sprintf("ENV_DB_USER_NAME=%s", dbUser),
		fmt.Sprintf("ENV_DB_USER_PASSWORD=%s", dbPassword),
		fmt.Sprintf("ENV_DB_DATABASE_NAME=%s", dbName),
	}

	cmd := exec.Command("make", makeArgs...)

	// Pass the parent environment to the child process.
	cmd.Env = os.Environ()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("\nError: 'make -C %s build:prod' command failed: %v\n", projectRoot, err)
		os.Exit(1)
	}

	fmt.Println("")
	fmt.Println("--> Deployment initiated successfully!")
}

// checkFile verifies a file exists, exiting on error.
func checkFile(path string) {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("Error: Secret file not found at: %s\n", path)
		} else {
			fmt.Printf("Error: Could not stat file %s: %v\n", path, err)
		}

		os.Exit(1)
	}
}

// readSecretContent reads the content of a secret file, trims whitespace, and returns it.
func readSecretContent(path string) string {
	content, err := os.ReadFile(path)

	if err != nil {
		fmt.Printf("Error: Failed to read secret file content from %s: %v\n", path, err)
		os.Exit(1)
	}

	// Trim trailing newlines or spaces, which are common in secret files.
	return strings.TrimSpace(string(content))
}
