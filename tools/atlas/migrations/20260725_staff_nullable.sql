DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'user_type') THEN
        CREATE TYPE user_type AS ENUM ('student', 'staff');
    END IF;
END
$$;

ALTER TABLE users
    ADD COLUMN IF NOT EXISTS user_type user_type;

-- Legacy rows predate CU NEX userType storage. New writes always use the API value.
UPDATE users
SET user_type = CASE
    WHEN length(ref_id::text) = 10 THEN 'student'::user_type
    ELSE 'staff'::user_type
END
WHERE user_type IS NULL;

ALTER TABLE users
    ALTER COLUMN user_type SET NOT NULL,
    ALTER COLUMN firstname_th DROP NOT NULL,
    ALTER COLUMN surname_th DROP NOT NULL,
    ALTER COLUMN title_th DROP NOT NULL,
    ALTER COLUMN faculty_name_th DROP NOT NULL,
    ALTER COLUMN firstname_en DROP NOT NULL,
    ALTER COLUMN surname_en DROP NOT NULL,
    ALTER COLUMN title_en DROP NOT NULL,
    ALTER COLUMN faculty_name_en DROP NOT NULL;

ALTER TABLE event_participants
    ALTER COLUMN organization DROP NOT NULL;

-- CU NEX returns an expiring SAS URL. It must never be persisted.
ALTER TABLE users
    DROP COLUMN IF EXISTS profile_image_url;
