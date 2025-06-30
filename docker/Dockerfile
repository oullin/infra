ARG GO_VERSION=1.24

# --- The builder image can be overridden. This allows to swap images for
#     local compilations defaults to the lean Alpine image for production.
ARG BUILDER_IMAGE=golang:${GO_VERSION}-alpine

# The final image base.
ARG FINAL_IMAGE=alpine:latest

# Build-time variables for your application.
ARG INFRA_VERSION="0.0.0.0"
ARG BUILD_TAGS="infra,deployer,gus,oullin"
ARG BINARY_NAME=deployment
ARG INFRA_USER=infrauser
ARG INFRA_GROUP=infragroup

# --- Target OS and Architecture for cross-compilation.
#     Defaults to a standard Linux build.
ARG GOOS=linux
ARG GOARCH=amd64

# Derived arguments
ARG INFRA_HOME=/home/${INFRA_USER}
ARG BUILD_DIR=/app

# --- Build Stage
#     This stage compiles the Go application.
#     It uses the BUILDER_IMAGE argument, allowing for different build environments.
FROM ${BUILDER_IMAGE} AS builder

# Forward arguments into this stage.
ARG BUILD_DIR
ARG BINARY_NAME
ARG INFRA_VERSION
ARG BUILD_TAGS
ARG GOOS
ARG GOARCH

# --- Ensure tzdata and ca-certificates are installed in the builder.
#     This works for both Alpine (apk) and Debian-based (apt) builders,
#     making the build process more robust and portable.
RUN if [ -f /etc/alpine-release ]; then \
    apk add --no-cache tzdata ca-certificates; \
else \
    apt-get update && apt-get install -y tzdata ca-certificates; \
fi

# Sets the working directory.
WORKDIR ${BUILD_DIR}

# Copy and download dependencies first to leverage Docker layer caching.
COPY ../go.mod go.sum ./
RUN go mod download

# Copy the rest of the application source code.
COPY .. .

# --- Compile the Go application.
#     This now uses the GOOS and GOARCH args passed in.
RUN GOOS=${GOOS} GOARCH=${GOARCH} CGO_ENABLED=0 go build -tags "${BUILD_TAGS}" -o ${BUILD_DIR}/${BINARY_NAME} -ldflags="-s -w -X main.Version=${INFRA_VERSION}" .

# --- Final Stage
#     Forward arguments needed for the final stage.
FROM ${FINAL_IMAGE}

ARG INFRA_USER
ARG INFRA_GROUP
ARG INFRA_HOME
ARG BINARY_NAME

# Create a dedicated, non-root user and group.
RUN addgroup -S ${INFRA_GROUP} && adduser -S ${INFRA_USER} -G ${INFRA_GROUP}

# Set the working directory.
WORKDIR ${INFRA_HOME}

# Create a bin directory for the executable.
RUN mkdir -p ${INFRA_HOME}/bin

# --- Copy the compiled binary from the builder stage using its explicit path.
#     We know the path in the builder is /app, so we use it directly here.
COPY --from=builder /app/${BINARY_NAME} ${INFRA_HOME}/bin/

# --- Copy required system data like timezone info.
#     Note: ca-certificates might not exist on the builder if it's minimal (like alpine).
#           This is safer when using different builders (like bookworm).
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy the .env file.
COPY .env .

# Set ownership for the application user.
RUN chown -R ${INFRA_USER}:${INFRA_GROUP} ${INFRA_HOME}

# Switch to the non-root user.
USER ${INFRA_USER}

# Define the default command to run when the container starts.
CMD ["./bin/${BINARY_NAME}"]
