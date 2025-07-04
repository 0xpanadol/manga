-- Comments Table with Polymorphic Association
CREATE TABLE "comments" (
  "id" uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  "user_id" uuid NOT NULL REFERENCES "users" ("id") ON DELETE CASCADE,
  "manga_id" uuid REFERENCES "manga" ("id") ON DELETE CASCADE,
  "chapter_id" uuid REFERENCES "chapters" ("id") ON DELETE CASCADE,
  "content" text NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT (now())

  -- Enforce that a comment belongs to EITHER a manga OR a chapter, but not both or neither.
  CONSTRAINT chk_comment_parent
  CHECK (
    (manga_id IS NOT NULL AND chapter_id IS NULL) OR
    (manga_id IS NULL AND chapter_id IS NOT NULL)
  )
);

-- Add indexes for faster lookups
CREATE INDEX ON "comments" ("user_id");
CREATE INDEX ON "comments" ("manga_id") WHERE manga_id IS NOT NULL;
CREATE INDEX ON "comments" ("chapter_id") WHERE chapter_id IS NOT NULL;