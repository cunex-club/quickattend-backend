CREATE TYPE attendence_type AS ENUM ('WHITELIST', 'FACULTIES', 'ALL');
CREATE TYPE role AS ENUM ('OWNER', 'STAFF', 'MANAGER');
CREATE TYPE participant_data AS ENUM ('NAME', 'ORGANIZATION', 'REFID', 'PHOTO');

CREATE TABLE "users" (
    "id" uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    "ref_id" bigint NOT NULL,
    "firstname_th" text NOT NULL,
    "surname_th" text NOT NULL,
    "title_th" text NOT NULL,
    "firstname_en" text NOT NULL,
    "surname_en" text NOT NULL,
    "title_en" text NOT NULL,

    -- ref_id must be UNIQUE
    CONSTRAINT "uni_users_ref_id" UNIQUE ("ref_id")
);

CREATE TABLE "events" (
    "id" uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    "name" text NOT NULL,
    "organizer" text NOT NULL,
    "description" text,
    -- "date" timestamp NOT NULL,
    "start_time" timestamp NOT NULL,
    "end_time" timestamp NOT NULL,
    "location" text NOT NULL,
    "attendence_type" attendence_type NOT NULL,
    "allow_all_to_scan" boolean NOT NULL,
    "evaluation_form" text,
    "revealed_fields" participant_data[] NOT NULL
);

CREATE TABLE "event_users" (
    "id" uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    "role" role NOT NULL,
    "user_id" uuid NOT NULL,
    "event_id" uuid NOT NULL,

    CONSTRAINT "fk_event_users_user" FOREIGN KEY ("user_id") 
        REFERENCES "users"("id") ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT "fk_event_users_event" FOREIGN KEY ("event_id") 
        REFERENCES "events"("id") ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE TABLE "event_whitelists" (
    "id" uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    "event_id" uuid NOT NULL,
    "attendee_ref_id" bigint NOT NULL,

    CONSTRAINT "fk_event_whitelists_event" FOREIGN KEY ("event_id") 
        REFERENCES "events"("id") ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT "fk_event_whitelists_user" FOREIGN KEY ("attendee_ref_id") 
        REFERENCES "users"("ref_id") ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE TABLE "event_allowed_faculties" (
    "id" uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    "event_id" uuid NOT NULL,
    "faculty_no" bigint NOT NULL, -- not sure but for now, matched to gorm int8

    CONSTRAINT "fk_event_allowed_faculties_event" FOREIGN KEY ("event_id") 
        REFERENCES "events"("id") ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE TABLE "event_agendas" (
    "id" uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    "event_id" uuid NOT NULL,
    "activity_name" text NOT NULL,
    "start_time" timestamptz NOT NULL,
    "end_time" timestamptz NOT NULL,

    CONSTRAINT "fk_event_agendas_event" FOREIGN KEY ("event_id") 
        REFERENCES "events"("id") ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE TABLE "event_participants" (
    "id" uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    "event_id" uuid NOT NULL,
    "checkin_timestamp" timestamptz,
    "comment" text,
    "scanned_timestamp" timestamptz NOT NULL,
    "participant_ref_id" bigint NOT NULL,
    "first_name" text NOT NULL,
    "sur_name" text NOT NULL,
    "organization" text NOT NULL,
    "scanned_location" point NOT NULL,
    "scanner_id" uuid,

    CONSTRAINT "fk_event_participants_event" FOREIGN KEY ("event_id") 
        REFERENCES "events"("id") ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT "fk_event_participants_scanner" FOREIGN KEY ("scanner_id") 
        REFERENCES "users"("id") ON UPDATE CASCADE ON DELETE SET NULL
);

-- CREATE INDEX "idx_event_participants_event_id" ON "event_participants"("event_id");
-- CREATE INDEX "idx_event_participants_participant_ref_id" ON "event_participants"("participant_ref_id");
-- CREATE INDEX "idx_event_whitelists_event_id" ON "event_whitelists"("event_id");
