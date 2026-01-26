CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TYPE attendence_type AS ENUM ('WHITELIST', 'FACULTIES', 'ALL');
CREATE TYPE participant_data AS ENUM ('NAME', 'ORGANIZATION', 'REFID', 'PHOTO');
CREATE TYPE role AS ENUM ('OWNER', 'STAFF', 'MANAGER');

CREATE TABLE users (
  id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
  ref_id bigint NOT NULL UNIQUE,
  firstname_th text NOT NULL,
  surname_th text NOT NULL,
  title_th text NOT NULL,
  firstname_en text NOT NULL,
  surname_en text NOT NULL,
  title_en text NOT NULL
);

CREATE TABLE events (
  id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
  name text NOT NULL,
  organizer text NOT NULL,
  description text,
  start_time timestamptz NOT NULL,
  end_time timestamptz NOT NULL,
  location text NOT NULL,
  attendence_type attendence_type NOT NULL,
  allow_all_to_scan boolean NOT NULL,
  evaluation_form text,
  revealed_fields participant_data[] NOT NULL
);

CREATE TABLE event_whitelists (
  id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
  event_id uuid NOT NULL,
  attendee_ref_id bigint NOT NULL,
  CONSTRAINT fk_event_whitelist_event
    FOREIGN KEY (event_id) REFERENCES events (id) ON UPDATE CASCADE ON DELETE CASCADE,
  CONSTRAINT fk_event_whitelist_user
    FOREIGN KEY (attendee_ref_id) REFERENCES users (ref_id) ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE TABLE event_allowed_faculties (
  id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
  event_id uuid NOT NULL,
  faculty_no int8 NOT NULL,
  CONSTRAINT fk_event_allowed_faculties_event
    FOREIGN KEY (event_id) REFERENCES events (id) ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE TABLE event_agendas (
  id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
  event_id uuid NOT NULL,
  activity_name text NOT NULL,
  start_time timestamptz NOT NULL,
  end_time timestamptz NOT NULL,
  CONSTRAINT fk_event_agenda_event
    FOREIGN KEY (event_id) REFERENCES events (id) ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE TABLE event_participants (
  id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
  event_id uuid NOT NULL,
  checkin_timestamp timestamptz,
  scanned_timestamp timestamptz NOT NULL,
  comment text,
  participant_id uuid NOT NULL UNIQUE,
  organization text NOT NULL,
  scanned_location point NOT NULL,
  scanner_id uuid NULL,
  CONSTRAINT fk_event_participants_event
    FOREIGN KEY (event_id) REFERENCES events (id) ON UPDATE CASCADE ON DELETE CASCADE,
  CONSTRAINT fk_event_participants_participant
    FOREIGN KEY (participant_id) REFERENCES users (id) ON UPDATE CASCADE ON DELETE CASCADE,
  CONSTRAINT fk_event_participants_scanner
    FOREIGN KEY (scanner_id) REFERENCES users (id) ON UPDATE CASCADE ON DELETE SET NULL
);

CREATE TABLE event_users (
  id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
  role role NOT NULL,
  user_id uuid NOT NULL,
  event_id uuid NOT NULL,
  CONSTRAINT fk_event_users_event
    FOREIGN KEY (event_id) REFERENCES events (id) ON UPDATE CASCADE ON DELETE CASCADE,
  CONSTRAINT fk_event_users_user
    FOREIGN KEY (user_id) REFERENCES users (id) ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE INDEX idx_events_name_trgm ON events USING GIN (name gin_trgm_ops);
CREATE INDEX idx_events_organizer_trgm ON events USING GIN (organizer gin_trgm_ops);
CREATE INDEX idx_events_description_trgm ON events USING GIN (description gin_trgm_ops);
CREATE INDEX idx_events_location_trgm ON events USING GIN (location gin_trgm_ops);
CREATE INDEX idx_events_evaluation_form_trgm ON events USING GIN (evaluation_form gin_trgm_ops);

CREATE INDEX idx_event_whitelists_event_id ON events_whitelists (event_id);
CREATE INDEX idx_event_users_event_id ON event_users (event_id);
CREATE INDEX idx_event_allowed_faculties_event_id ON event_allowed_faculties (event_id);
CREATE INDEX idx_event_agendas_event_id ON event_agendas (event_id);
CREATE INDEX idx_event_participants_event_id ON event_participants (event_id);