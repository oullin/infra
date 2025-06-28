package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// ==============================================================================
// Oullin Production Deployment Program (v2)
//
// This program is a Go equivalent of the original deployment shell script.
// It's designed to be run on the production VPS to prepare the environment
// and launch the Docker Compose services.
//
// It prepares the environment by verifying credentials from a secure,
// non-repository location and then passes their paths to the 'make' command.
//
// Author: Gus (Original Script), Gemini (Go Translation)
// ==============================================================================

func main() {
	// --- Configuration ---
	// The absolute path where secrets are securely stored on the VPS.
	secretsDir := "/home/gocanto/.oullin/secrets"
	// The path where the deployment-related files are. We'll derive the root from this.
	apiDir := "/home/gocanto/Sites/oullin/api"
	// Define the project's root directory, assuming it's one level above apiDir.
	projectRoot := filepath.Dir(apiDir) // This will resolve to /home/gocanto/Sites/oullin

	// --- Pre-flight Checks ---
	fmt.Println("--> [1/3] Verifying secret files...")

	// Check if the secrets directory exists
	if info, err := os.Stat(secretsDir); err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("Error: Secrets directory not found at %s\n", secretsDir)
		} else {
			fmt.Printf("Error: Could not stat secrets directory: %v\n", err)
		}
		os.Exit(1)
	} else if !info.IsDir() {
		fmt.Printf("Error: Path %s is not a directory.\n", secretsDir)
		os.Exit(1)
	}

	// Define the full paths to the individual secret files
	userSecretFile := filepath.Join(secretsDir, "postgres_user")
	passwordSecretFile := filepath.Join(secretsDir, "postgres_password")
	dbSecretFile := filepath.Join(secretsDir, "postgres_db")

	// Check that each required secret file exists individually for clearer error reporting.
	checkFile(userSecretFile)
	checkFile(passwordSecretFile)
	checkFile(dbSecretFile)

	fmt.Println("--> Secret files verified successfully.")

	// --- Environment Preparation ---
	fmt.Println("--> [2/3] Preparing environment for 'make' command...")

	// These environment variables will be passed to the child process (make)
	envVars := []string{
		"POSTGRES_USER_SECRET_PATH=" + userSecretFile,
		"POSTGRES_PASSWORD_SECRET_PATH=" + passwordSecretFile,
		"POSTGRES_DB_SECRET_PATH=" + dbSecretFile,
	}

	fmt.Println("--> Environment variables are set for the deployment process.")

	// --- Deployment Execution ---
	fmt.Printf("--> [3/3] Launching Docker Compose services via 'make' from root (%s)...\n", projectRoot)

	// Prepare the command to be executed.
	// The '-C' flag tells 'make' to change to the specified directory first.
	// This ensures it finds the Makefile at the project root.
	cmd := exec.Command("make", "-C", projectRoot+"/api", "build:prod")

	// Pass the current environment plus our new variables to the command.
	cmd.Env = append(os.Environ(), envVars...)

	// Connect the command's output and error streams to the main process
	// so we can see the output in the console.
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run the command
	if err := cmd.Run(); err != nil {
		fmt.Printf("\nError: 'make -C %s build:prod' command failed: %v\n", projectRoot, err)
		os.Exit(1)
	}

	fmt.Println("")
	fmt.Println("--> Deployment initiated successfully!")
}

// checkFile is a helper function to verify a file exists.
// It prints a specific error message and exits if the file is not found.
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
