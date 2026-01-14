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
  checkin_timestamp timestamptz NOT NULL,
  participant_ref_id bigint NOT NULL,
  first_name text NOT NULL,
  sur_name text NOT NULL,
  organization text NOT NULL,
  scanned_location point NOT NULL,
  scanner_id uuid NULL,
  CONSTRAINT fk_event_participants_event
    FOREIGN KEY (event_id) REFERENCES events (id) ON UPDATE CASCADE ON DELETE CASCADE,
  CONSTRAINT fk_event_participants_user_ref
    FOREIGN KEY (participant_ref_id) REFERENCES users (ref_id) ON UPDATE CASCADE ON DELETE CASCADE,
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