-- 1. Add a new column to the manga table to store the tsvector.
ALTER TABLE "manga" ADD COLUMN "search_tsv" tsvector;

-- 2. Create a function that will be used by the trigger.
-- This function concatenates the title and description and converts them to a tsvector.
-- 'coalesce' is used to handle potential NULL values gracefully.
CREATE OR REPLACE FUNCTION manga_search_trigger() RETURNS trigger AS $$
begin
  new.search_tsv :=
    setweight(to_tsvector('pg_catalog.english', coalesce(new.title,'')), 'A') ||
    setweight(to_tsvector('pg_catalog.english', coalesce(new.description,'')), 'B');
  return new;
end
$$ LANGUAGE plpgsql;

-- 3. Create a trigger that calls the function whenever a row is inserted or updated.
CREATE TRIGGER tsvectorupdate BEFORE INSERT OR UPDATE
ON "manga" FOR EACH ROW EXECUTE PROCEDURE manga_search_trigger();

-- 4. Populate the new column for all existing manga records.
-- This is important for data that already exists before the trigger is created.
UPDATE "manga" SET search_tsv =
    setweight(to_tsvector('pg_catalog.english', coalesce(title,'')), 'A') ||
    setweight(to_tsvector('pg_catalog.english', coalesce(description,'')), 'B');

-- 5. Create a GIN index on the new column for fast searching.
CREATE INDEX manga_search_tsv_idx ON "manga" USING GIN(search_tsv);