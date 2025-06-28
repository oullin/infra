# Go & Docker Deployment and Permissions

This document provides a complete guide for automating the deployment of a Go application using Docker Compose on a 
production VPS. It details a secure and robust process that began as a simple shell script and evolved into a 
sophisticated Go-based deployment program.

This guide captures the key learnings and final solutions from a detailed debugging process, making it a practical manual 
for spinning up new environments or for internal documentation.

---

## 1. The Initial Goal: From Shell Script to Go Program

The project started with a simple goal: to replace a shell-based deployment script with a more robust and portable Go program.

### The Original `deploy.sh` Script

The initial deployment was handled by a `bash` script that performed pre-flight checks for secret files, exported their paths as environment variables, and then ran a `make` command.

```bash
#!/bin/bash
set -e
SECRETS_DIR="/home/gocanto/.oullin/secrets"
API_DIR="/home/gocanto/Sites/oullin/api"

echo "--> [1/3] Verifying secret files..."
# ... checks for secret files ...

echo "--> [2/3] Exporting secret paths..."
export POSTGRES_USER_SECRET_PATH="$SECRETS_DIR/postgres_user"
export POSTGRES_PASSWORD_SECRET_PATH="$SECRETS_DIR/postgres_password"
export POSTGRES_DB_SECRET_PATH="$SECRETS_DIR/postgres_db"

echo "--> [3/3] Launching Docker Compose services..."
cd $API_DIR || exit 1
make build:prod
```

### The Go Solution: `deployment.go`

To make this process more secure and self-contained, we translated this logic into a Go program. This eliminates shell 
dependencies and creates a single, compilable binary that can be run on any similar VPS.

The final, working version of the deployment program reads the content of the secret files and passes all necessary 
variables directly to the `make` command as arguments. This is a secure approach that keeps all secrets in memory and 
avoids intermediate files.

**Final `deployment.go`**
---

## 2. The Core Challenge & Solution: Securely Connecting the Pieces

The most significant challenge was getting the credentials securely from the Go program into the running `api` container. 
The final, robust solution involves a three-part harmony between the Go program, the `Makefile`, and `docker-compose.yml`.

### Step 2.1: The Go Deployer (Covered Above)

The `deployment.go` script acts as the entry point, reading secrets and passing them as arguments to `make`.

### Step 2.2: The Makefile Bridge

The `Makefile` is the critical bridge. It receives the variables from the Go program and uses them to construct the final `docker-compose` command.

**Key Learning:** Simply exporting variables with `.EXPORT_ALL_VARIABLES` was not consistently reliable. The most robust 
method is to prefix the variables directly to the `docker-compose` command.

**Final `Makefile` rule (e.g., in `config/makefile/build.mk`):**

### Step 2.3: The Docker Compose Configuration

The `docker-compose.yml` file for the `api` service must be configured to accept the environment variables passed in by 
the `Makefile`.

**Final** `api` service configuration in **`docker-compose.yml`:**
---

## 3. The Debugging Journey: Key Learnings

The path to the final solution revealed several critical insights.

### Learning 1: The True Source of Environment Variables

The most crucial breakthrough was realizing that the application code itself dictates the names of the environment variables it uses.

**Problem:** We initially tried variables like `DB_USER` and `POSTGRES_USER`, but authentication kept failing.
**Discovery:** By inspecting the application's `factory.go` file, we found the exact names it was looking for.

```go
// Snippet from factory.go
db := env.DBEnvironment{
   UserName:     env.GetEnvVar("ENV_DB_USER_NAME"),
   UserPassword: env.GetEnvVar("ENV_DB_USER_PASSWORD"),
   DatabaseName: env.GetEnvVar("ENV_DB_DATABASE_NAME"),
   // ...
}
```

**Lesson:** When debugging credential issues, **always** verify the variable names in the application source code.

### Learning 2: Stale Data in Docker Volumes

After fixing the variable names, the application was finally sending the *correct username* but still failing authentication.

**Problem:** `password authentication failed for user "oullin_gocanto_db"`
**Discovery:** The persistent Docker volume (`oullin_db_data`) was holding a stale database initialized with an *old, incorrect password*.
**Lesson:** When credentials change, the database's persistent volume must be reset to allow it to re-initialize with the new secrets.

**Solution:**

1. Stop all services: `docker-compose down`
2. Find the exact volume name: `docker volume ls`
3. Remove the stale volume (this deletes all DB data): `docker volume rm <project_name>_oullin_db_data`

---

## 4. The Final End-to-End Workflow

Here is the complete, step-by-step process to deploy the application on a new VPS using this system.

1. **Place Secrets:**
    * On the VPS, create the secrets directory: `mkdir -p /home/gocanto/.oullin/secrets`
    * Create the three secret files with the correct content (no trailing newlines):
        * `/home/gocanto/.oullin/secrets/postgres_user`
        * `/home/gocanto/.oullin/secrets/postgres_password`
        * `/home/gocanto/.oullin/secrets/postgres_db`
2. **Configure Project:**
    * Ensure the `deployment.go`, `Makefile`, and `docker-compose.yml` files in your project match the final versions provided in this guide.
3. **Run Deployment:**
    * Navigate to your project directory.
    * Execute the Go program. This single command handles the entire deployment process.
      ```bash
      go run deployment.go
      ```
4. **Verify:**
    * Check that all containers are running and healthy.
      ```bash
      docker ps
      ```
    * Check the logs of the `api` container to confirm a successful database connection.
      ```bash
      docker logs <api_container_name_or_id>
      ```

By following this guide, you can consistently and securely deploy your application, leveraging the power of Go for 
automation and Docker for containerization.
