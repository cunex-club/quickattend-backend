-- Modify "events" table
ALTER TABLE "events" ADD COLUMN "location_point" point NOT NULL;
-- Create index "idx_events_end_time" to table: "events"
CREATE INDEX "idx_events_end_time" ON "events" ("end_time");
-- Create "event_whitelist_pendings" table
CREATE TABLE "event_whitelist_pendings" ("id" uuid NOT NULL DEFAULT gen_random_uuid(), "event_id" uuid NOT NULL, "attendee_ref_id" bigint NOT NULL, PRIMARY KEY ("id"), CONSTRAINT "whitelist_pendings_unique_event_and_ref_id" UNIQUE ("event_id", "attendee_ref_id"), CONSTRAINT "fk_event_whitelist_pendings_event" FOREIGN KEY ("event_id") REFERENCES "events" ("id") ON UPDATE CASCADE ON DELETE CASCADE);
