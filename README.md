# Oullin Infrastructure

## Introduction

This project is designed to streamline the deployment of the Oullin applications by leveraging Docker and Make for a consistent and automated build and deployment process. The infrastructure is set up to handle servers, certificates, and documentation.

## Intent

The primary goal of this project is to automate the deployment of Go applications. It's designed to take the hassle out of manual deployments by providing a scriptable and repeatable process. This is achieved through a combination of Go code that orchestrates the deployment, Docker to create a consistent build environment, and Make to execute the build and deployment commands.

## Tutorial

This tutorial will guide you through the process of using the Oullin Infrastructure project to deploy your Go application.

### 1. Configuration

Before you can deploy the application, you need to set up your environment variables and secret files.

* **.env file**: The project uses a `.env` file to manage environment-specific configurations. You'll need to create a `.env` file in the root of the project and populate it with the necessary variables, such as `API_SECRETS_DIRECTORY` and `API_DIRECTORY`.
* **Secret Files**: The deployment process relies on secret files for sensitive information like database credentials. You'll need to create the following files in your secrets directory:
    * `postgres_db`: Containing the name of the database.
    * `postgres_user`: Containing the database username.
    * `postgres_password`: Containing the database password.

### 2. Building the Application

The project uses Docker to build the Go application in a containerised environment, ensuring consistency across different development machines. The `docker-compose.yml` file defines the `deployment` service, which is responsible for building the application.

To build the application, you can run the following command:

```bash
make build
```
This will create a Docker image named oullin/infra-builder containing the compiled Go binary.

### 3. Running the Deployment
   The `main.go` file is the entry point for the deployment process. It reads the environment variables, creates a new deployment, and then runs it.

To run the deployment, you can execute the following command:

```bash
make run
```

This will trigger the deployment process, which in turn executes a make command with the necessary secrets and configurations to build and deploy the API application (for now).

## Description
The Oullin Infrastructure project is composed of several key components that work together to automate the deployment process.

- **Go Application:** The core of the project is a Go application that orchestrates the deployment. This application is responsible for reading configurations, parsing secrets, and executing the deployment commands.
- **Docker Integration:** Docker is used to create a consistent and isolated build environment. The docker-compose.yml file defines the build service and its dependencies, ensuring that the application is built with the correct Go version and libraries every time.
- **Makefile:** The project uses a Makefile to define the build and deployment commands. The deployer.go file executes a make command with the appropriate targets and arguments to build and deploy the application.
- **Secrets' Management:** The project has a robust system for managing secrets. It uses separate files for each secret, which are then read and parsed by the Go application at runtime. This ensures that sensitive information is not hardcoded in the source code.

## Makefile and Available Commands
A Makefile is a file that contains a set of directives used by the make build automation tool. In this project, the Makefile is used to automate the process of building and deploying the Go application.

Based on the `api/deployer.go` file, the following make command is being used:
- **build:** This is the primary command used for creating a production build of the application.

## How it Works
The Go application, specifically in the `api/deployer.go` file, dynamically constructs and executes a make `build:prod command`.
This command is passed a series of arguments that are used to configure the build process:

- **POSTGRES_USER_SECRET_PATH:** The path to the file containing the PostgreSQL username.
- **POSTGRES_PASSWORD_SECRET_PATH:** The path to the file containing the PostgreSQL password.
- **POSTGRES_DB_SECRET_PATH:** The path to the file containing the PostgreSQL database name.
- **ENV_DB_USER_NAME:** The PostgreSQL username.
- **ENV_DB_USER_PASSWORD:** The PostgreSQL password.
- **ENV_DB_DATABASE_NAME:** The PostgreSQL database name.

These variables are then used within the `Makefile` to configure the build environment, likely for embedding the database
credentials or other secrets into the final application binary or for use in the build process itself.

In summary, while the `Makefile` is not present, we can see that it's a crucial part of the deployment process, and the
main command available is `build:prod` for creating production-ready builds of your Go application.

## Contributing

Please feel free to fork this package and contribute by submitting a pull request to enhance its functionality.

## License

The MIT License (MIT). Please see [License File](https://github.com/oullin/infra/blob/main/LICENSE) for more information.

## How can I thank you?

There are many ways you would be able to support my open source work. There is not a right one to choose, so the choice is yours.

Nevertheless :grinning:, I would propose the following

- :arrow_up: Follow me on [Twitter](https://twitter.com/gocanto).
- :star: Star the repository.
- :handshake: Open a pull request to fix/improve the codebase.
- :writing_hand: Open a pull request to improve the documentation.
- :coffee: Buy me a [coffee](https://github.com/sponsors/gocanto)?

> Thank you for reading this far. :blush:
