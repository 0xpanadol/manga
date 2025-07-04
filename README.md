# Go Manga API
A production-ready HTTP API in Go for a manga reading platform, similar to MangaDex. This backend service is built with a focus on clean architecture, performance, and maintainability.

## Features

- **User Authentication**: JWT-based (access/refresh tokens) authentication with secure password hashing (bcrypt).
- **Role-Based Access Control (RBAC)**: Differentiated permissions for Admins and regular Users.
- **Manga & Chapter Management**: Full CRUD API for managing the manga catalog and its chapters.
- **Media Uploads**: S3-compatible object storage integration (using MinIO) for chapter page uploads.
- **Social Features**:
  - Favorite/Follow manga.
  - Track reading progress.
  - Polymorphic commenting system for both manga and chapters.
- **API Documentation**: Auto-generated, interactive Swagger/OpenAPI documentation.
- **Containerized**: Fully containerized with Docker and Docker Compose for easy local development.
- **Testing**: Comprehensive unit and integration tests with a dedicated test database.
- **CI/CD**: Automated linting, testing, and Docker image publishing via GitHub Actions.

## Tech Stack

- **Language**: Go 1.21+
- **Framework**: Gin Web Framework
- **Database**: PostgreSQL
- **Object Storage**: MinIO (S3-Compatible)
- **Authentication**: JWT
- **Migrations**: `golang-migrate`
- **Testing**: `testify`, `mockery`

---

## Getting Started

### Prerequisites

- [Go](https://golang.org/doc/install) (version 1.21 or later)
- [Docker](https://www.docker.com/get-started) and [Docker Compose](https://docs.docker.com/compose/install/)
- [make](https://www.gnu.org/software/make/) (optional, for using the Makefile shortcuts)
- `migrate` CLI: `go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest`
- `swag` CLI: `go install github.com/swaggo/swag/cmd/swag@latest`

### 1. Clone the Repository

```bash
git clone https://github.com/your-username/manga-api.git
cd manga-api
```

### 2. Configure Environment

Copy the example environment file and customize if needed. The default values are configured to work with the provided `docker-compose.yml`.

```bash
cp .env.example .env
```

### 3. Run the Development Environment

This single command will build the Docker images and start the API, PostgreSQL, and MinIO containers.

```bash
make up
```
The services will be available at:
- **API Server**: `http://localhost:8080`
- **PostgreSQL Database**: `localhost:5432`
- **MinIO Console**: `http://localhost:9001` (Login with `minioadmin`/`minioadmin`)

### 4. Apply Database Migrations

With the containers running, apply the database schema.

```bash
make migrateup
```

---

## Usage

### API Documentation

The interactive Swagger/OpenAPI documentation is the best way to explore the API. Once the application is running, navigate to:

**[http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)**

### Key API Endpoints

- **Auth**: `/api/v1/auth/register`, `/api/v1/auth/login`
- **Manga**: `/api/v1/manga`, `/api/v1/manga/{id}`
- **Chapters**: `/api/v1/chapters/{id}`, `/api/v1/manga/{manga_id}/chapters`
- **Comments**: `/api/v1/manga/{id}/comments`, `/api/v1/chapters/{id}/comments`
- **User Profile**: `/api/v1/users/me` (Protected)

---

## Development Workflow

### Running Tests

This command will spin up a dedicated test database, run all unit and integration tests, and then tear down the test database.

```bash
make test
```

### Generating Mocks

If you modify a repository interface, regenerate the mocks using:

```bash
make mocks
```
This uses the `.mockery.yml` configuration file.

### Regenerating API Docs

After adding or modifying handler annotations, update the Swagger documentation:

```bash
make swag
```

### Database Migrations

Create a new migration file:
```bash
migrate create -ext sql -dir migrations -seq your_migration_name
```

Apply all pending migrations:
```bash
make migrateup
```

Roll back the last applied migration:
```bash
make migratedown
```

### Stopping the Environment

To stop and remove all containers, use:
```bash
make down
```