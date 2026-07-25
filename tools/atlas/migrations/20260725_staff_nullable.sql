-- Brings a database created from the v1.0.1 schema up to the current one.
--
-- This repo has no migration runner, so this file is applied by hand and may
-- be run against a database in an unknown state. Every statement is therefore
-- idempotent, and the whole thing runs in one transaction so a failure can't
-- leave the schema half-applied.
--
-- Derived by diffing tools/atlas/schema.sql at tag v1.0.1 against HEAD, then
-- verified on PostgreSQL 17 by applying it to a v1.0.1 database (with rows
-- present) and diffing the result — columns, constraints and indexes — against
-- a database built directly from the current schema.sql.

BEGIN;

-- 1) user_type enum
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'user_type') THEN
        CREATE TYPE user_type AS ENUM ('student', 'staff');
    END IF;
END
$$;

-- 2) columns added after v1.0.1
ALTER TABLE users ADD COLUMN IF NOT EXISTS user_type user_type;
ALTER TABLE users ADD COLUMN IF NOT EXISTS faculty_name_th text;
ALTER TABLE users ADD COLUMN IF NOT EXISTS faculty_name_en text;

-- 3) Backfill user_type on pre-existing rows before enforcing NOT NULL.
--    Inferred from ref_id length: students are 10 digits, everyone else is staff.
UPDATE users
SET user_type = CASE
    WHEN length(ref_id::text) = 10 THEN 'student'::user_type
    ELSE 'staff'::user_type
END
WHERE user_type IS NULL;

ALTER TABLE users ALTER COLUMN user_type SET NOT NULL;

-- 4) CU NEX doesn't always return Thai names or an affiliation for staff,
--    so these can no longer be NOT NULL.
ALTER TABLE users
    ALTER COLUMN firstname_th   DROP NOT NULL,
    ALTER COLUMN surname_th     DROP NOT NULL,
    ALTER COLUMN title_th       DROP NOT NULL,
    ALTER COLUMN faculty_name_th DROP NOT NULL,
    ALTER COLUMN firstname_en   DROP NOT NULL,
    ALTER COLUMN surname_en     DROP NOT NULL,
    ALTER COLUMN title_en       DROP NOT NULL,
    ALTER COLUMN faculty_name_en DROP NOT NULL;

ALTER TABLE event_participants ALTER COLUMN organization DROP NOT NULL;

-- 5) CU NEX returns an expiring SAS URL. It must never be persisted.
ALTER TABLE users DROP COLUMN IF EXISTS profile_image_url;

-- 6) Table added after v1.0.1: holds roles for people who haven't logged in
--    or been scanned yet.
CREATE TABLE IF NOT EXISTS event_user_pendings (
    "id"          uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    "role"        role   NOT NULL,
    "user_ref_id" bigint NOT NULL,
    "event_id"    uuid   NOT NULL,

    CONSTRAINT "unique_user_and_event_pendings" UNIQUE ("user_ref_id", "event_id"),
    CONSTRAINT "fk_event_user_pendings_event" FOREIGN KEY ("event_id")
        REFERENCES "events"("id") ON UPDATE CASCADE ON DELETE CASCADE
);

COMMIT;
