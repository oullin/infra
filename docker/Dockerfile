# API Docker File.

# --- Build Arguments for Configuration
#     Defines variables that make this Dockerfile more configurable and reusable.
#     These can be overridden during the build process (e.g., via docker-compose).
ARG GO_VERSION=1.24
ARG ALPINE_VERSION=latest

ARG INFRA_VERSION="0.0.0.0"
ARG BUILD_TAGS="infra,deployer,gus,oullin"

ARG BINARY_NAME=infra
ARG INFRA_USER=infrauser
ARG INFRA_GROUP=infragroup

# Derived arguments
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
RUN apk add --no-cache tzdata

# Sets the primary working directory for this stage of the build.
WORKDIR ${BUILD_DIR}

# Copies the Go module definition files into the builder.
COPY ../go.mod go.sum ./

# Downloads the application's external dependencies.
RUN go mod download

# Copies the rest of the application's source code into the builder.
COPY .. .

# Compiles the Go application into a single, statically-linked binary.
RUN CGO_ENABLED=0 go build -tags "${BUILD_TAGS}" -o ${BUILD_DIR}/${BINARY_NAME} -ldflags="-s -w -X main.Version=${INFRA_VERSION}" .

# --- Final Stage
#     This is the final, production-ready stage of the build.
#     It creates a minimal environment to hold the compiled artifact.
FROM alpine:${ALPINE_VERSION}

# Forwards build-time arguments into this final stage so they can be referenced.
ARG INFRA_USER
ARG INFRA_GROUP
ARG INFRA_HOME
ARG BINARY_NAME

# Creates a dedicated, non-root user and group for the application.
RUN addgroup -S ${INFRA_GROUP} && adduser -S ${INFRA_USER} -G ${INFRA_GROUP}

# Sets the working directory for the final container.
WORKDIR ${INFRA_HOME}

# --- MODIFICATION: Create a /bin directory for the binary ---
RUN mkdir -p ${INFRA_HOME}/bin

# Copies the compiled application binary from the 'builder' stage into the new /bin directory.
COPY --from=builder /app/${BINARY_NAME} ${INFRA_HOME}/bin/

# Copies the timezone database from the 'builder' stage.
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copies the .env file into the container.
COPY .env .

# Recursively sets the ownership of all files in the application's home directory.
RUN chown -R ${INFRA_USER}:${INFRA_GROUP} ${INFRA_HOME}

# Switches the context of the container to run as the non-root user.
USER ${INFRA_USER}

# --- Defines the default command to run the binary from the /bin folder.
#     This will run if someone executes `docker run` on the image.
CMD ["./bin/${BINARY_NAME}"]
