-- Genres Table: Stores all possible genres
CREATE TABLE "genres" (
  "id" uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  "name" varchar(50) UNIQUE NOT NULL
);

-- Manga Table
CREATE TYPE manga_status AS ENUM ('ongoing', 'completed', 'hiatus', 'cancelled');

CREATE TABLE "manga" (
  "id" uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  "title" varchar(255) NOT NULL,
  "description" text NOT NULL,
  "author" varchar(100) NOT NULL,
  "status" manga_status NOT NULL DEFAULT 'ongoing',
  "cover_image_url" varchar(255),
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT (now())
);

-- Manga_Genres Junction Table: Many-to-many relationship
CREATE TABLE "manga_genres" (
  "manga_id" uuid NOT NULL REFERENCES "manga" ("id") ON DELETE CASCADE,
  "genre_id" uuid NOT NULL REFERENCES "genres" ("id") ON DELETE CASCADE,
  PRIMARY KEY ("manga_id", "genre_id")
);

-- Add a new permission for managing manga
INSERT INTO permissions (code) VALUES ('manga:manage');

-- Assign the new permission to the Admin role
-- Note: You need the Admin role's static UUID from the first migration
INSERT INTO roles_permissions (role_id, permission_id)
SELECT
  'a4198182-a398-4244-9635-5b58f3286d79', -- Admin Role ID
  id FROM permissions WHERE code = 'manga:manage';

-- Seed some initial genres for testing
INSERT INTO genres (name) VALUES
('Action'), ('Adventure'), ('Comedy'), ('Drama'), ('Fantasy'), ('Horror'),
('Isekai'), ('Mecha'), ('Mystery'), ('Romance'), ('Sci-Fi'), ('Slice of Life');