package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type APIDeployment struct {
	secretsDir   string
	apiDir       string
	projectRoot  string
	apiDBSecrets APIDbSecrets
}

type APIDbSecrets struct {
	userSecretFile     string
	passwordSecretFile string
	dbSecretFile       string
}

func NewAPIDeployment(secretsDir, apiDir string) *APIDeployment {
	return &APIDeployment{
		secretsDir:  secretsDir,
		apiDir:      apiDir,
		projectRoot: strings.TrimSpace(filepath.Dir(apiDir) + "/api"),
		apiDBSecrets: APIDbSecrets{
			userSecretFile:     filepath.Join(secretsDir, "postgres_user"),
			passwordSecretFile: filepath.Join(secretsDir, "postgres_password"),
			dbSecretFile:       filepath.Join(secretsDir, "postgres_db"),
		},
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
