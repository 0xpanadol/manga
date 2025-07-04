# Project Blueprint: Go Manga API

This document serves as the comprehensive architectural and implementation guide for the Go Manga API project. It details the project structure, data models, API endpoints, and core component signatures.

## 1. Project Architecture

The project follows a **Clean Architecture** / **Hexagonal** structure, emphasizing separation of concerns and dependency inversion.

### 1.1. Directory Structure

```
manga-api/
├── .github/workflows/      # CI/CD pipelines (GitHub Actions)
├── cmd/api/                # Main application entry point
├── docs/                   # Auto-generated Swagger/OpenAPI files
├── internal/
│   ├── config/             # Configuration loading (Viper)
│   ├── domain/             # Core business models (structs)
│   ├── repository/         # Data access layer (interfaces & mocks)
│   │   ├── mocks/
│   │   └── postgres/       # PostgreSQL implementation of repositories
│   ├── service/            # Business logic orchestration
│   └── transport/http/     # HTTP-specific components
│       ├── handler/        # Gin handlers
│       ├── middleware/     # Gin middleware
│       └── router/         # Route definitions
├── migrations/             # SQL migration files
├── pkg/                    # Shared, reusable packages
│   ├── jwtauth/            # JWT generation and validation
│   ├── password/           # Bcrypt password hashing
│   └── uploader/           # MinIO/S3 file uploader
├── .env.example            # Example environment variables
├── .golangci.yml           # Linter configuration
├── .mockery.yml            # Mock generation configuration
├── Dockerfile              # Application Docker image definition
├── docker-compose.yml      # Local development environment
├── go.mod                  # Go module definition
└── Makefile                # Developer command shortcuts
```

### 1.2. Data Flow
`HTTP Request` -> `transport/http/router` -> `transport/http/middleware` -> `transport/http/handler` -> `service` -> `repository` (interface) -> `repository/postgres` (implementation) -> `Database`

## 2. Database Schema
**Database**: PostgreSQL
### Tables:
- `users`: Stores user credentials and `role_id`.
- `roles`: Defines roles (e.g., 'Admin', 'User').
- `permissions`: Defines granular permissions (e.g., 'manga:manage').
- `roles_permissions`: Links roles to permissions (many-to-many).
- `manga`: Core manga catalog information.
- `genres`: Stores all possible genre names.
- `manga_genres`: Links manga to genres (many-to-many).
- `chapters`: Stores chapter details, linked to a manga.
- `comments`: Polymorphic table for comments, linked to a user and EITHER a manga OR a chapter.
- `user_favorites`: Links users to their favorited manga (many-to-many).
- `user_reading_progress`: Links users to chapters they have read (many-to-many).

## 3. Core Domain Models (`internal/domain/`)

- **`User`**: `{ ID, Username, Email, PasswordHash, RoleID, CreatedAt, UpdatedAt }`
- **`Role`**: `{ ID, Name, Permissions[] }`
- **`Permission`**: `{ ID, Code }`
- **`Manga`**: `{ ID, Title, Description, Author, Status, CoverImageURL, Genres[], CreatedAt, UpdatedAt }`
- **`Chapter`**: `{ ID, MangaID, ChapterNumber, Title, Pages[], CreatedAt, UpdatedAt }`
- **`Comment`**: `{ ID, UserID, MangaID*, ChapterID*, Content, CreatedAt, UpdatedAt }` (*nullable)
- **`CommentWithUser`**: `Comment` struct + `Username`

## 4. Component Signatures

### 4.1. Services (`internal/service/`)

- `NewAuthService(repo, jwtSecrets...)` -> `*AuthService`
  - `Register(ctx, username, email, password)` -> `(*User, error)`
  - `Login(ctx, email, password)` -> `(*jwtauth.TokenDetails, error)`
- `NewUserService(repo)` -> `*UserService`
  - `GetProfile(ctx, userID)` -> `(*User, error)`
- `NewMangaService(repo)` -> `*MangaService`
  - `Create(ctx, manga)` -> `error`
  - `GetByID(ctx, id)` -> `(*Manga, error)`
  - `List(ctx, params)` -> `([]*Manga, error)`
  - `Update(ctx, manga)` -> `error`
  - `Delete(ctx, id)` -> `error`
- `NewChapterService(repo, uploader)` -> `*ChapterService`
  - `Create(ctx, chapter)` -> `error`
  - `GetByID(ctx, id)` -> `(*Chapter, error)`
  - `ListByMangaID(ctx, params)` -> `([]*Chapter, error)`
  - `Update(ctx, chapter)` -> `error`
  - `Delete(ctx, id)` -> `error`
  - `UploadPages(ctx, chapterID, files)` -> `error`
- `NewSocialService(repo)` -> `*SocialService`
  - `ToggleFavorite(ctx, userID, mangaID)` -> `(*ToggleFavoriteResult, error)`
  - `ListFavorites(ctx, userID, params)` -> `([]*Manga, error)`
  - `MarkChapterAsRead(ctx, userID, chapterID)` -> `error`
  - `ListReadChapters(ctx, userID)` -> `([]*Chapter, error)`
  - `CreateComment(ctx, comment)` -> `error`
  - `ListComments(ctx, params)` -> `([]*CommentWithUser, error)`

### 4.2. Repositories (`internal/repository/`)

- **`UserRepository`**: `Create`, `FindByEmail`, `FindByID`, `FindDefaultUserRoleID`, `GetRoleAndPermissions`
- **`MangaRepository`**: `Create`, `FindByID`, `List`, `Update`, `Delete`
- **`ChapterRepository`**: `Create`, `FindByID`, `ListByMangaID`, `Update`, `Delete`, `UpdatePages`
- **`SocialRepository`**: `ToggleFavorite`, `ListFavorites`, `MarkChapterAsRead`, `ListReadChapters`, `CreateComment`, `ListComments`

## 5. API Endpoints

**Base Path**: `/api/v1`

| Method | Endpoint                               | Handler Function         | Protection     | Description                                |
|--------|----------------------------------------|--------------------------|----------------|--------------------------------------------|
| **Auth** |                                        |                          |                |                                            |
| `POST` | `/auth/register`                       | `AuthHandler.Register`   | Public         | Register a new user.                       |
| `POST` | `/auth/login`                          | `AuthHandler.Login`      | Public         | Authenticate and receive JWTs.             |
| **Users** |                                        |                          |                |                                            |
| `GET`  | `/users/me`                            | `UserHandler.GetMe`      | Authenticated  | Get the current user's profile.            |
| **Manga** |                                        |                          |                |                                            |
| `POST` | `/manga`                               | `MangaHandler.CreateManga` | Admin          | Create a new manga.                        |
| `GET`  | `/manga`                               | `MangaHandler.ListManga` | Public         | List, filter, and paginate manga.          |
| `GET`  | `/manga/{id}`                          | `MangaHandler.GetManga`  | Public         | Get a single manga by ID.                  |
| `PUT`  | `/manga/{id}`                          | `MangaHandler.UpdateManga` | Admin          | Update a manga.                            |
| `DELETE`| `/manga/{id}`                          | `MangaHandler.DeleteManga` | Admin          | Delete a manga.                            |
| **Chapters** |                                        |                          |                |                                            |
| `POST` | `/manga/{manga_id}/chapters`           | `ChapterHandler.CreateChapter` | Admin      | Create a new chapter for a manga.          |
| `GET`  | `/manga/{manga_id}/chapters`           | `ChapterHandler.ListChapters`  | Public     | List chapters for a manga.                 |
| `GET`  | `/chapters/{id}`                       | `ChapterHandler.GetChapter`    | Public     | Get a single chapter by ID.                |
| `PUT`  | `/chapters/{id}`                       | `ChapterHandler.UpdateChapter` | Admin      | Update a chapter.                          |
| `DELETE`| `/chapters/{id}`                       | `ChapterHandler.DeleteChapter` | Admin      | Delete a chapter.                          |
| `POST` | `/chapters/{id}/pages`                 | `ChapterHandler.UploadPages`   | Admin      | Upload pages (images) for a chapter.       |
| **Social** |                                        |                          |                |                                            |
| `POST` | `/manga/{id}/favorite`                 | `SocialHandler.ToggleFavorite` | Authenticated | Toggle favorite status for a manga.        |
| `GET`  | `/users/me/favorites`                  | `SocialHandler.ListFavorites`  | Authenticated | List the current user's favorite manga.    |
| `POST` | `/chapters/{id}/progress`              | `SocialHandler.MarkChapterAsRead` | Authenticated | Mark a chapter as read.                    |
| `GET`  | `/users/me/progress`                   | `SocialHandler.ListReadChapters` | Authenticated | List chapters read by the current user.    |
| `POST` | `/manga/{id}/comments`                 | `SocialHandler.CreateMangaComment` | Authenticated | Post a comment on a manga.                 |
| `GET`  | `/manga/{id}/comments`                 | `SocialHandler.ListMangaComments`  | Public     | List comments for a manga.                 |
| `POST` | `/chapters/{id}/comments`              | `SocialHandler.CreateChapterComment` | Authenticated | Post a comment on a chapter.               |
| `GET`  | `/chapters/{id}/comments`              | `SocialHandler.ListChapterComments`  | Public     | List comments for a chapter.               |
| **System** |                                        |                          |                |                                            |
| `GET`  | `/healthz`                             | N/A                      | Public         | Health check endpoint.                     |
| `GET`  | `/swagger/*any`                        | N/A                      | Public         | Serves the Swagger UI.                     |

## 6. Configuration (`.env`)

The application is configured via environment variables, loaded from a `.env` file.

- `APP_PORT`: Port for the HTTP server.
- `DB_URL`: PostgreSQL connection string.
- `JWT_ACCESS_SECRET`: Secret key for access tokens.
- `JWT_REFRESH_SECRET`: Secret key for refresh tokens.
- `JWT_ACCESS_EXPIRES_IN`: Lifetime of access tokens (e.g., `15m`).
- `JWT_REFRESH_EXPIRES_IN`: Lifetime of refresh tokens (e.g., `168h`).
- `MINIO_ENDPOINT`: Endpoint for the S3-compatible storage.
- `MINIO_ACCESS_KEY`: Access key for MinIO.
- `MINIO_SECRET_KEY`: Secret key for MinIO.
- `MINIO_BUCKET_NAME`: Bucket name for storing pages.
- `MINIO_USE_SSL`: Whether to use SSL for MinIO connection.
- `TEST_DB_URL`: PostgreSQL connection string for the integration test database.