# API Docker File.

# --- Build Arguments for Configuration ---
# Defines variables that make this Dockerfile more configurable and reusable.
# These can be overridden during the build process (e.g., via docker-compose).
ARG GO_VERSION=1.24
ARG ALPINE_VERSION=latest

ARG INFRA_VERSION="0.0.0-dev"
ARG BUILD_TAGS="infra,deployer,gus,oullin"

ARG BINARY_NAME=infra
ARG INFRA_USER=infrauser
ARG INFRA_GROUP=infragroup
ARG INFRA_HOME=/home/${INFRA_USER}

ARG BUILD_DIR=/app
ARG LOGS_DIR=logs

# --- Build Stage ---
#     This stage, named 'builder', is responsible for compiling the Go application.
#     It uses a Go-specific base image that includes the necessary toolchain.
FROM golang:${GO_VERSION}-alpine AS builder

# Forwards build-time arguments into this specific stage so they can be referenced.
ARG BUILD_DIR
ARG BINARY_NAME
ARG INFRA_VERSION
ARG BUILD_TAGS

# Installs the timezone database package into the builder image.
# This is necessary because the base 'alpine' image is minimal and does not
# include this data by default. It is copied to the final image later.
RUN apk add --no-cache tzdata

# Sets the primary working directory for this stage of the build.
WORKDIR ${BUILD_DIR}

# Copies the Go module definition files into the builder.
# This is done first to leverage Docker's layer caching. The subsequent
# 'go mod download' step will only be re-run if these files have changed.
COPY ../go.mod go.sum ./

# Downloads the application's external dependencies as specified in the go.mod file.
# The 'replace' directive in go.mod is used to resolve local packages, so this
# command will not attempt to download them from the internet.
RUN go mod download

# Copies the rest of the application's source code into the builder.
# This includes the main package and any local packages like 'boost'.
COPY .. .

# Compiles the Go application into a single, statically-linked binary.
# -tags: Applies build constraints, allowing for conditional compilation.
# -o: Specifies the output path and name for the compiled binary.
# -ldflags: Provides flags to the linker.
#   -s: Strips the symbol table, reducing binary size.
#   -w: Strips DWARF debugging information, further reducing size.
#   -X: Injects a value into a string variable at build time. Here, it sets
#       the application's version by targeting the 'Version' variable in the 'main' package.
RUN CGO_ENABLED=0 go build -tags "${BUILD_TAGS}" -o ${BUILD_DIR}/${BINARY_NAME} -ldflags="-s -w -X main.Version=${INFRA_VERSION}" .

# --- Final Stage
#     This is the final, production-ready stage of the build.
#     It uses a minimal 'alpine' base image to create a small and secure container.
FROM alpine:${ALPINE_VERSION}

# Forwards build-time arguments into this final stage so they can be referenced.
ARG INFRA_USER
ARG INFRA_GROUP
ARG INFRA_HOME
ARG BUILD_DIR
ARG BINARY_NAME
ARG LOGS_DIR

# Creates a dedicated, non-root user and group for the application.
# Running the application as a non-root user is a critical security best practice.
RUN addgroup -S ${INFRA_GROUP} && adduser -S ${INFRA_USER} -G ${INFRA_GROUP}

# Sets the working directory for the final container.
WORKDIR ${INFRA_HOME}

# Copies the compiled application binary from the 'builder' stage.
# This is the core of the multi-stage build pattern, ensuring the final image
# contains only the compiled application and not the Go toolchain or source code.
COPY --from=builder ${BUILD_DIR}/${BINARY_NAME} .

# Copies the timezone database from the 'builder' stage.
# This ensures that time-related functions in the application work correctly.
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copies the .env file into the container.
# This allows the application to load its configuration from environment variables.
# For this to work, '.env' must not be in the .dockerignore file.
COPY .env .

# Recursively sets the ownership of all files in the application's home directory.
# This ensures the non-root application user has the correct permissions to execute the binary
# and write to the storage directories.
RUN chown -R ${INFRA_USER}:${INFRA_GROUP} ${INFRA_HOME}

# Switches the context of the container to run as the non-root user.
# Any subsequent commands (like the CMD) will be executed by this user.
USER ${INFRA_USER}

# Exposes the application's port from the container.
# This does not publish the port; it serves as documentation for which port to map.
EXPOSE ${INFRA_HOST_PORT}

# Defines the default command to execute when the container starts.
# This runs the compiled application binary.
CMD ["./${BINARY_NAME}"]