-- User_Favorites Table: A many-to-many relationship between users and manga
CREATE TABLE "user_favorites" (
  "user_id" uuid NOT NULL REFERENCES "users" ("id") ON DELETE CASCADE,
  "manga_id" uuid NOT NULL REFERENCES "manga" ("id") ON DELETE CASCADE,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  PRIMARY KEY ("user_id", "manga_id")
);

-- User_Reading_Progress Table: A many-to-many relationship between users and chapters
CREATE TABLE "user_reading_progress" (
  "user_id" uuid NOT NULL REFERENCES "users" ("id") ON DELETE CASCADE,
  "chapter_id" uuid NOT NULL REFERENCES "chapters" ("id") ON DELETE CASCADE,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  PRIMARY KEY ("user_id", "chapter_id")
);

-- No new permissions are needed as these are user-specific actions,
-- not administrative ones. Authentication is sufficient.