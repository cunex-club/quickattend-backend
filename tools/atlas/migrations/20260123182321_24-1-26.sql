-- Create enum type "attendence_type"
CREATE TYPE "attendence_type" AS ENUM ('WHITELIST', 'FACULTIES', 'ALL');
-- Create enum type "role"
CREATE TYPE "role" AS ENUM ('OWNER', 'STAFF', 'MANAGER');
-- Create enum type "participant_data"
CREATE TYPE "participant_data" AS ENUM ('NAME', 'ORGANIZATION', 'REFID', 'PHOTO');
-- Create "events" table
CREATE TABLE "events" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "name" text NOT NULL,
  "organizer" text NOT NULL,
  "description" text NULL,
  "start_time" timestamp NOT NULL,
  "end_time" timestamp NOT NULL,
  "location" text NOT NULL,
  "attendence_type" "attendence_type" NOT NULL,
  "allow_all_to_scan" boolean NOT NULL,
  "evaluation_form" text NULL,
  "revealed_fields" "participant_data"[] NOT NULL,
  PRIMARY KEY ("id")
);
-- Create "event_agendas" table
CREATE TABLE "event_agendas" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "event_id" uuid NOT NULL,
  "activity_name" text NOT NULL,
  "start_time" timestamptz NOT NULL,
  "end_time" timestamptz NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_event_agendas_event" FOREIGN KEY ("event_id") REFERENCES "events" ("id") ON UPDATE CASCADE ON DELETE CASCADE
);
-- Create "event_allowed_faculties" table
CREATE TABLE "event_allowed_faculties" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "event_id" uuid NOT NULL,
  "faculty_no" bigint NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_event_allowed_faculties_event" FOREIGN KEY ("event_id") REFERENCES "events" ("id") ON UPDATE CASCADE ON DELETE CASCADE
);
-- Create "users" table
CREATE TABLE "users" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "ref_id" bigint NOT NULL,
  "firstname_th" text NOT NULL,
  "surname_th" text NOT NULL,
  "title_th" text NOT NULL,
  "firstname_en" text NOT NULL,
  "surname_en" text NOT NULL,
  "title_en" text NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "uni_users_ref_id" UNIQUE ("ref_id")
);
-- Create "event_participants" table
CREATE TABLE "event_participants" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "event_id" uuid NOT NULL,
  "checkin_timestamp" timestamptz NULL,
  "comment" text NULL,
  "scanned_timestamp" timestamptz NOT NULL,
  "participant_ref_id" bigint NOT NULL,
  "first_name" text NOT NULL,
  "sur_name" text NOT NULL,
  "organization" text NOT NULL,
  "scanned_location" point NOT NULL,
  "scanner_id" uuid NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_event_participants_event" FOREIGN KEY ("event_id") REFERENCES "events" ("id") ON UPDATE CASCADE ON DELETE CASCADE,
  CONSTRAINT "fk_event_participants_scanner" FOREIGN KEY ("scanner_id") REFERENCES "users" ("id") ON UPDATE CASCADE ON DELETE SET NULL
);
-- Create "event_users" table
CREATE TABLE "event_users" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "role" "role" NOT NULL,
  "user_id" uuid NOT NULL,
  "event_id" uuid NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_event_users_event" FOREIGN KEY ("event_id") REFERENCES "events" ("id") ON UPDATE CASCADE ON DELETE CASCADE,
  CONSTRAINT "fk_event_users_user" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE CASCADE ON DELETE CASCADE
);
-- Create "event_whitelists" table
CREATE TABLE "event_whitelists" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "event_id" uuid NOT NULL,
  "attendee_ref_id" bigint NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_event_whitelists_event" FOREIGN KEY ("event_id") REFERENCES "events" ("id") ON UPDATE CASCADE ON DELETE CASCADE,
  CONSTRAINT "fk_event_whitelists_user" FOREIGN KEY ("attendee_ref_id") REFERENCES "users" ("ref_id") ON UPDATE CASCADE ON DELETE CASCADE
);
