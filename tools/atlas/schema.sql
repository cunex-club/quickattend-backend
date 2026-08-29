CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TYPE attendence_type AS ENUM ('WHITELIST', 'FACULTIES', 'ALL');
CREATE TYPE role AS ENUM ('OWNER', 'STAFF', 'MANAGER');
CREATE TYPE participant_data AS ENUM ('NAME', 'ORGANIZATION', 'REFID', 'PHOTO');
CREATE TYPE user_type AS ENUM ('student', 'staff');

CREATE TABLE "users" (
    "id" uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    "ref_id" bigint NOT NULL UNIQUE,
    "user_type" user_type NOT NULL,
    "firstname_th" text,
    "surname_th" text,
    "title_th" text,
    "faculty_name_th" text,
    "firstname_en" text,
    "surname_en" text,
    "title_en" text,
    "faculty_name_en" text
);

CREATE TABLE "events" (
    "id" uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    "name" text NOT NULL,
    "organizer" text NOT NULL,
    "description" text,
    "start_time" timestamptz NOT NULL,
    "end_time" timestamptz NOT NULL,
    "location" text NOT NULL,
    "location_point" point NOT NULL,
    "attendence_type" attendence_type NOT NULL,
    "allow_all_to_scan" boolean NOT NULL,
    "evaluation_form" text,
    "revealed_fields" participant_data[] NOT NULL
);

CREATE TABLE "event_whitelists" (
    "id" uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    "event_id" uuid NOT NULL,
    "attendee_ref_id" bigint NOT NULL,

    CONSTRAINT "unique_event_and_ref_id" UNIQUE ("event_id", "attendee_ref_id"),
    CONSTRAINT "fk_event_whitelists_event" FOREIGN KEY ("event_id") 
        REFERENCES "events"("id") ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT "fk_event_whitelists_user" FOREIGN KEY ("attendee_ref_id") 
        REFERENCES "users"("ref_id") ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE TABLE "event_allowed_faculties" (
    "id" uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    "event_id" uuid NOT NULL,
    "faculty_no" bigint NOT NULL,

    CONSTRAINT "unique_event_and_faculty_no" UNIQUE ("event_id", "faculty_no"),
    CONSTRAINT "fk_event_allowed_faculties_event" FOREIGN KEY ("event_id") 
        REFERENCES "events"("id") ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE TABLE "event_agendas" (
    "id" uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    "event_id" uuid NOT NULL,
    "activity_name" text NOT NULL,
    "start_time" timestamptz NOT NULL,
    "end_time" timestamptz NOT NULL,

    CONSTRAINT "unique_event_start_end" UNIQUE ("event_id", "start_time", "end_time"),
    CONSTRAINT "fk_event_agendas_event" FOREIGN KEY ("event_id") 
        REFERENCES "events"("id") ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE TABLE "event_participants" (
    "id" uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    "event_id" uuid NOT NULL,
    "comment_timestamp" timestamptz,
    "comment" text,
    "scanned_timestamp" timestamptz NOT NULL,
    "participant_id" uuid NOT NULL,
    "organization" text,
    "scanned_location" point NOT NULL,
    "scanner_id" uuid,

    CONSTRAINT "unique_event_and_participant" UNIQUE ("event_id", "participant_id"),
    CONSTRAINT "fk_event_participants_event" FOREIGN KEY ("event_id") 
        REFERENCES "events"("id") ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT "fk_event_participants_participant" FOREIGN KEY ("participant_id") 
        REFERENCES "users" ("id") ON UPDATE CASCADE ON DELETE CASCADE,      
    CONSTRAINT "fk_event_participants_scanner" FOREIGN KEY ("scanner_id") 
        REFERENCES "users"("id") ON UPDATE CASCADE ON DELETE SET NULL
);

CREATE TABLE "event_users" (
    "id" uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    "role" role NOT NULL,
    "user_id" uuid NOT NULL,
    "event_id" uuid NOT NULL,

    CONSTRAINT "unique_user_and_event" UNIQUE ("user_id", "event_id"),
    CONSTRAINT "fk_event_users_user" FOREIGN KEY ("user_id") 
        REFERENCES "users"("id") ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT "fk_event_users_event" FOREIGN KEY ("event_id") 
        REFERENCES "events"("id") ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE TABLE "event_whitelist_pendings" (
    "id" uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    "event_id" uuid NOT NULL,
    "attendee_ref_id" bigint NOT NULL,

    CONSTRAINT "whitelist_pendings_unique_event_and_ref_id" UNIQUE ("event_id", "attendee_ref_id"),
    CONSTRAINT "fk_event_whitelist_pendings_event" FOREIGN KEY ("event_id") 
        REFERENCES "events"("id") ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE TABLE "event_user_pendings" (
    "id" uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    "role" role NOT NULL,
    "user_ref_id" bigint NOT NULL,
    "event_id" uuid NOT NULL,

    CONSTRAINT "unique_user_and_event_pendings" UNIQUE ("user_ref_id", "event_id"),
    CONSTRAINT "fk_event_user_pendings_event" FOREIGN KEY ("event_id") 
        REFERENCES "events"("id") ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE TABLE "event_participant_stats" (
    "event_id" uuid PRIMARY KEY,
    "total_participants" int NOT NULL,
    "total_student" int NOT NULL,
    "total_staff" int NOT NULL,
    "total_eligible" int,
    "organization_stats" jsonb NOT NULL,
    "time_series_stats" jsonb NOT NULL,
    "created_at" timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT "fk_event_participant_stats_event" FOREIGN KEY ("event_id")
        REFERENCES "events"("id") ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE INDEX idx_events_end_time ON events (end_time);
CREATE INDEX idx_events_name_trgm ON events USING GIN (name gin_trgm_ops);
CREATE INDEX idx_events_organizer_trgm ON events USING GIN (organizer gin_trgm_ops);
CREATE INDEX idx_events_description_trgm ON events USING GIN (description gin_trgm_ops);
CREATE INDEX idx_events_location_trgm ON events USING GIN (location gin_trgm_ops);
CREATE INDEX idx_events_evaluation_form_trgm ON events USING GIN (evaluation_form gin_trgm_ops);
