DROP INDEX IF EXISTS manga_search_tsv_idx;
DROP TRIGGER IF EXISTS tsvectorupdate ON "manga";
DROP FUNCTION IF EXISTS manga_search_trigger();
ALTER TABLE "manga" DROP COLUMN IF EXISTS "search_tsv";
