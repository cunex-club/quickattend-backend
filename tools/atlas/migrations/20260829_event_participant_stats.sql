BEGIN;

CREATE TABLE IF NOT EXISTS "event_participant_stats" (
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

COMMIT;
