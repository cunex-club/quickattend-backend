-- Modify "event_user_pendings" table
ALTER TABLE "event_user_pendings" DROP CONSTRAINT "unique_user_and_event_pendings", DROP COLUMN "user_id", ADD COLUMN "user_ref_id" bigint NOT NULL, ADD CONSTRAINT "unique_user_and_event_pendings" UNIQUE ("user_ref_id", "event_id");
