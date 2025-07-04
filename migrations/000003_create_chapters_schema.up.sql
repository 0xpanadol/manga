-- Chapters Table
CREATE TABLE "chapters" (
  "id" uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  "manga_id" uuid NOT NULL REFERENCES "manga" ("id") ON DELETE CASCADE,
  "chapter_number" varchar(20) NOT NULL, -- Using varchar for flexibility (e.g., "10.5", "Extra")
  "title" varchar(255), -- Optional title for the chapter
  "pages" text[] NOT NULL DEFAULT '{}', -- Array of image URLs for the pages
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT (now()),
  UNIQUE("manga_id", "chapter_number") -- A manga can't have two chapters with the same number
);

-- Add a new permission for managing chapters
INSERT INTO permissions (code) VALUES ('chapters:manage');

-- Assign the new permission to the Admin role
-- Note: You need the Admin role's static UUID from the first migration
INSERT INTO roles_permissions (role_id, permission_id)
SELECT
  'a4198182-a398-4244-9635-5b58f3286d79', -- Admin Role ID
  id FROM permissions WHERE code = 'chapters:manage';