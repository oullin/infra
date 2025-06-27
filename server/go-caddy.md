## Tutorial: Building a Production-Ready Go & Caddy Stack for CI/CD

This tutorial provides a comprehensive, step-by-step guide to creating a secure, configurable, and multienvironment 
deployment for a Go application. We will use Caddy as a reverse proxy and Docker Compose for orchestration. This process 
documents the debugging journey and results in a system that is perfectly primed for integration into a CI/CD pipeline.

### Prerequisites

Before you begin, ensure you have the following installed:
* **Docker and Docker Compose:** The core containerization tools.
* **Go (1.24 or later):** For managing local dependencies.
* A text editor or IDE (like VSCode).

---

### Step 1: Project Structure

First, create the following directory and file structure. This organization keeps our services clean and separated, 
which is crucial for maintainability.

```
.
├── caddy/
│   ├── Caddyfile.local
│   ├── Caddyfile.prod
│   └── Dockerfile
├── database/
│   └── (Your database scripts and config)
├── docker/
│   └── dockerfile-api
├── boost/
│   └── (Your local Go package files)
├── main.go
├── go.mod
├── go.sum
├── .env
└── docker-compose.yml
```

---

### Step 2: The Go Application & Local Dependencies

The core of our system is the Go application. A key challenge in containerizing Go projects is managing internal packages correctly.

#### The `go.mod` `replace` Directive

When your `main.go` file imports a local package (like our `boost` package), the Go compiler needs to be told where to find it. Without instruction, it will try to download it from the internet, leading to build failures.

The `replace` directive in `go.mod` solves this. It's a standard and simple way to alias an import path to a local directory.

**Action:** Add the following line to the bottom of your `go.mod` file.

```go
// In your go.mod file
replace [github.com/oullin/boost](https://github.com/oullin/boost) => ./boost
```

This tells the Go compiler, "When you see an import for `github.com/oullin/boost`, use the local `./boost` directory instead of going online."

---

### Step 3: Containerizing the Services

We will create separate, configurable Dockerfiles for our API and Caddy services.

#### 1. The API Dockerfile (`./docker/dockerfile-api`)

We use a multi-stage Dockerfile to create a small, secure final image. This file is made highly configurable using `ARG` variables, a critical feature for CI/CD pipelines where you'll want to pass in version numbers and tags dynamically.

**Key Concepts:**
* **Multi-stage Build:** The `builder` stage compiles the application. The final stage copies *only* the compiled binary and necessary assets, resulting in a minimal and secure image.
* **`ARG`:** Build-time variables that make the Dockerfile reusable and configurable from `docker-compose.yml` or your CI/CD runner.
* **Layer Caching:** We `COPY go.mod go.sum` and run `go mod download` first. This layer is only rebuilt if your dependencies change, speeding up later builds.

**Action:** Create the `./docker/dockerfile-api` file with the following content.

```dockerfile
# Filename: ./docker/dockerfile-api
ARG GO_VERSION=1.24
ARG ALPINE_VERSION=latest
ARG APP_VERSION="0.0.0-dev"
ARG BUILD_TAGS="posts,expirence,profile,projects,social,talks,gus,gocanto"
ARG BINARY_NAME=server
ARG APP_HOST_PORT=8080
ARG APP_USER=appuser
ARG APP_GROUP=appgroup
ARG APP_HOME=/home/${APP_USER}
ARG BUILD_DIR=/app
ARG STORAGE_DIR=storage
ARG LOGS_DIR=logs
ARG MEDIA_DIR=media
ARG FIXTURES_DIR=fixture

# --- Build Stage ---
FROM golang:${GO_VERSION}-alpine AS builder
ARG BUILD_DIR
ARG BINARY_NAME
ARG APP_VERSION
ARG BUILD_TAGS
RUN apk add --no-cache tzdata
WORKDIR ${BUILD_DIR}
COPY go.mod go.sum ./
RUN go mod download
COPY .. .
RUN CGO_ENABLED=0 go build -tags "${BUILD_TAGS}" -o ${BUILD_DIR}/${BINARY_NAME} -ldflags="-s -w -X main.Version=${APP_VERSION}" .

# --- Final Stage ---
FROM alpine:${ALPINE_VERSION}
ARG APP_USER
ARG APP_GROUP
ARG APP_HOME
ARG BUILD_DIR
ARG BINARY_NAME
ARG APP_HOST_PORT
ARG STORAGE_DIR
ARG LOGS_DIR
ARG MEDIA_DIR
ARG FIXTURES_DIR
RUN addgroup -S ${APP_GROUP} && adduser -S ${APP_USER} -G ${APP_GROUP}
WORKDIR ${APP_HOME}
RUN mkdir -p ${STORAGE_DIR}/${LOGS_DIR} ${STORAGE_DIR}/${MEDIA_DIR}
COPY ${STORAGE_DIR}/${FIXTURES_DIR} ./${STORAGE_DIR}/${FIXTURES_DIR}/
COPY --from=builder ${BUILD_DIR}/${BINARY_NAME} .
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY .env .
RUN chown -R ${APP_USER}:${APP_GROUP} ${APP_HOME}
USER ${APP_USER}
EXPOSE ${APP_HOST_PORT}
CMD ["./${BINARY_NAME}"]
```

#### 2. The Caddy Dockerfile (`./caddy/Dockerfile`)

For production stability, we pin Caddy to a specific version using an `ARG`.

**Action:** Create the `./caddy/Dockerfile`.

```dockerfile
ARG CADDY_VERSION=2.10.0
FROM caddy:${CADDY_VERSION}
COPY Caddyfile /etc/caddy/Caddyfile
```

---

### Step 4: The Caddy Reverse Proxy Configuration

We need two different configurations for Caddy to handle local development versus a live production environment.

* **`Caddyfile.local`:** For local development, this file is configured for simple, unencrypted HTTP traffic to avoid certificate warnings.
* **`Caddyfile.prod`:** For production, this file is configured to use your domain name and automatically handle HTTPS certificate provisioning via Let's Encrypt. It also includes important security headers.

**Action (`Caddyfile.local`):**
```caddy
{
    auto_https off
}
:80 {
    reverse_proxy api:8080
}
```

**Action (`Caddyfile.prod`):**
```caddy
your-domain.com {
    encode gzip zstd
    header {
       Referrer-Policy "strict-origin-when-cross-origin"
       Strict-Transport-Security "max-age=31536000;"
       X-Frame-Options "SAMEORIGIN"
       X-Content-Type-Options "nosniff"
    }
    reverse_proxy api:8080
}
```

---

### Step 5: The `docker-compose.yml` Orchestrator

This file is the heart of our system, defining how all the services (`api`, `api-db`, `caddy`) are built, configured, and networked together.

#### 1. Networking and Service Discovery

Docker Compose creates private networks that allow containers to communicate. We use two:
* `caddy_net`: For Caddy to talk to the API.
* `oullin_net`: For the API to talk to the database.

Inside these networks, containers can find each other using their **service name** as a hostname (e.g., the `api` service connects to the database at `api-db:5432`).

#### 2. Runtime Configuration & Overrides

This is the most critical concept for CI/CD.
* **`env_file:`**: This block tells a service to load its base configuration from your local `.env` file.
* **`environment:`**: This block **overrides** any values from the `env_file`. This is how we solve the `localhost` problem for both the database and the web server host.

#### 3. Profiles: Switching Between Environments

To manage our separate `local` and `prod` Caddy configurations, we use **Docker Compose Profiles**.
* We define two Caddy services: `caddy_local` and `caddy_prod`.
* Each is assigned to a profile (`profiles: ["local"]` or `profiles: ["prod"]`).
* Each service mounts the correct `Caddyfile`.

**Action:** Create the `docker-compose.yml` file.

```yaml
version: '3.8'

volumes:
    caddy_data:
    caddy_config:

networks:
    caddy_net:
        name: caddy_net
        driver: bridge
    oullin_net:
        name: oullin_net
        driver: bridge

services:
    caddy_prod:
        build:
            context: ./caddy
            dockerfile: Dockerfile
            args:
                - CADDY_VERSION=2.10.0
        # This service will only run when the 'prod' profile is active.
        profiles: ["prod"]
        container_name: oullin_proxy_prod
        restart: unless-stopped
        depends_on:
            - api
        ports:
            - "80:80"
            - "443:443"
            - "443:443/udp" # Required for HTTP/3
        volumes:
            - caddy_data:/data
            - caddy_config:/config
            - ./caddy/Caddyfile.prod:/etc/caddy/Caddyfile
        networks:
            - caddy_net

    caddy_local:
        build:
            context: ./caddy
            dockerfile: Dockerfile
            args:
                - CADDY_VERSION=latest
        # This service will only run when the 'local' profile is active.
        profiles: ["local"]
        container_name: oullin_local_proxy
        restart: unless-stopped
        depends_on:
            - api
        ports:
            - "8080:80"
            - "8443:443"
        volumes:
            - caddy_data:/data
            - caddy_config:/config
            - ./caddy/Caddyfile.local:/etc/caddy/Caddyfile
        networks:
            - caddy_net

    api:
        env_file:
            - .env
        environment:
            # This ensures the API connects to the correct database container.
            ENV_DB_HOST: api-db
            # This ensures the Go web server listens for connections from other
            # containers (like Caddy), not just from within itself.
            ENV_HTTP_HOST: 0.0.0.0
        build:
            context: .
            dockerfile: ./docker/dockerfile-api
            args:
                - APP_VERSION=v1.0.0-release
                - APP_HOST_PORT=${ENV_HTTP_PORT}
                - APP_USER=${ENV_DOCKER_USER}
                - APP_GROUP=${ENV_DOCKER_USER_GROUP}
        container_name: oullin_api
        restart: unless-stopped
        depends_on:
            api-db:
                condition: service_healthy
        expose:
            - ${ENV_HTTP_PORT}
        networks:
            - caddy_net
            - oullin_net

    api-db:
        restart: unless-stopped
        image: postgres:17.4
        container_name: oullin_db
        env_file:
            - .env
        networks:
            - oullin_net
        environment:
            # --- Postgres CLI env vars.
            PGUSER: ${ENV_DB_USER_NAME}
            PGDATABASE: ${ENV_DB_DATABASE_NAME}
            PGPASSWORD: ${ENV_DB_USER_PASSWORD}
            # --- Docker postgres-image env vars.
            POSTGRES_USER: ${ENV_DB_USER_NAME}
            POSTGRES_DB: ${ENV_DB_DATABASE_NAME}
            POSTGRES_PASSWORD: ${ENV_DB_USER_PASSWORD}
        ports:
            - "${ENV_DB_PORT}:${ENV_DB_PORT}"
        volumes:
            - ./database/infra/ssl/server.crt:/etc/ssl/certs/server.crt
            - ./database/infra/ssl/server.key:/etc/ssl/private/server.key
            - ./database/infra/data:/var/lib/postgresql/data
            - ./database/infra/config/postgresql.conf:/etc/postgresql/postgresql.conf
            - ./database/infra/docker-entrypoint-initdb.d:/docker-entrypoint-initdb.d
        logging:
            driver: "json-file"
            options:
                max-file: "20"
                max-size: "10M"
        command: >
            sh -c "chown postgres:postgres /etc/ssl/private/server.key && chmod 600 /etc/ssl/private/server.key && exec docker-entrypoint.sh postgres"
        healthcheck:
            interval: 10s
            timeout: 5s
            retries: 5
            test: [
                "CMD-SHELL",
                "pg_isready",
                "--username=${ENV_DB_USER_NAME}",
                "--dbname=${ENV_DB_DATABASE_NAME}",
                "--host=api-db",
                "--port=${ENV_DB_PORT}",
                "--version"
            ]
```

---

### Step 6: Running the Application

With all the files in place, you can now run your entire stack with a single command.

#### To Run for Local Development:

This command activates the "local" profile, which starts the `caddy_local` service.
```bash
docker compose --profile local up --build -d
```
Your API will be accessible at **`http://localhost:8080`**.

#### To Run for Production:

This command activates the "prod" profile, which starts the `caddy_prod` service. Ensure your domain's DNS is pointing to your server's IP address.
```bash
docker compose --profile prod up --build -d
```
Your API will be accessible at **`https://your-domain.com`**.

---

### Step 7: Publishing to a Container Registry for CI/CD

The final step before fully automating is to publish your built images to a registry like the GitHub Container Registry (`ghcr.io`).

1.  **Create a GitHub Personal Access Token (PAT):**
    * Go to GitHub **Settings** > **Developer settings** > **Personal access tokens** > **Tokens (classic)**.
    * Generate a new token with the **`write:packages`** scope.
    * Copy the token immediately.

2.  **Log in from your Terminal:**
    ```bash
    docker login ghcr.io -u YOUR_GITHUB_USERNAME
    ```
    Use the PAT as your password.

3.  **Tag Your Images:**
    * Find your local image names with `docker images`.
    * Tag them for the registry:
        ```bash
        docker tag <local_api_image> ghcr.io/YOUR_GITHUB_USERNAME/oullin_api:latest
        docker tag <local_caddy_image> ghcr.io/YOUR_GITHUB_USERNAME/oullin_proxy:latest
        ```

4.  **Push the Images:**
    ```bash
    docker push ghcr.io/YOUR_GITHUB_USERNAME/oullin_api:latest
    docker push ghcr.io/YOUR_GITHUB_USERNAME/oullin_proxy:latest
    ```

Your CI/CD pipeline will automate these tagging and pushing steps, and your production server will then pull these versioned 
images directly from the registry.

You've successfully built a scalable, secure, and professional foundation for your application.
