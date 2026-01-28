BEGIN;
-- =========================
-- USERS
-- =========================
INSERT INTO users (id, ref_id, firstname_th, surname_th, title_th, firstname_en, surname_en, title_en)
VALUES
  ('11111111-1111-1111-1111-111111111111', 10001, 'สมชาย', 'ใจดี', 'นาย', 'Somchai', 'Jaidee', 'Mr.'),
  ('22222222-2222-2222-2222-222222222222', 10002, 'สมหญิง', 'แสนดี', 'นางสาว', 'Somying', 'Saendee', 'Ms.'),
  ('33333333-3333-3333-3333-333333333333', 10003, 'วิทยา', 'เก่งงาน', 'นาย', 'Withaya', 'Kengngan', 'Mr.'),
  ('44444444-4444-4444-4444-444444444444', 10004, 'อรทัย', 'ตั้งใจ', 'นาง', 'Orathai', 'Tangjai', 'Mrs.'),
  ('55555555-5555-5555-5555-555555555555', 10005, 'ธนา', 'สุขใจ', 'นาย', 'Thana', 'Sukjai', 'Mr.');

-- =========================
-- EVENTS
-- =========================
INSERT INTO events (
  id, name, organizer, description,
  start_time, end_time, location,
  attendence_type, allow_all_to_scan,
  evaluation_form, revealed_fields
  )
VALUES
  -- Event open to everyone
  (
  'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
  'Open Tech Talk',
  'Engineering Club',
  'Public tech sharing session',
  NOW() - INTERVAL '1 day',
  NOW() + INTERVAL '1 day',
  'Main Hall',
  'ALL',
  true,
  'https://forms.example.com/open-tech',
  ARRAY['NAME', 'ORGANIZATION']::participant_data[]
  ),

  -- Event whitelist only
  (
  'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb',
  'VIP Strategy Meeting',
  'Board Office',
  'Invitation-only meeting',
  NOW(),
  NOW() + INTERVAL '3 hours',
  'Meeting Room A',
  'WHITELIST',
  false,
  NULL,
  ARRAY['NAME', 'REFID', 'PHOTO']::participant_data[]
  ),

  -- Event by faculties
  (
  'cccccccc-cccc-cccc-cccc-cccccccccccc',
  'Faculty Research Day',
  'Academic Affairs',
  'Research presentations by faculty',
  NOW() + INTERVAL '2 days',
  NOW() + INTERVAL '3 days',
  'Conference Center',
  'FACULTIES',
  true,
  'https://forms.example.com/research',
  ARRAY['NAME', 'ORGANIZATION', 'REFID']::participant_data[]
  );

-- =========================
-- EVENT USERS (ROLES)
-- =========================
INSERT INTO event_users (role, user_id, event_id)
VALUES
  -- Open Tech Talk
  ('OWNER',   '11111111-1111-1111-1111-111111111111', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'),
  ('STAFF',   '22222222-2222-2222-2222-222222222222', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'),

  -- VIP Strategy
  ('OWNER',   '33333333-3333-3333-3333-333333333333', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb'),
  ('MANAGER', '44444444-4444-4444-4444-444444444444', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb'),

  -- Faculty Day
  ('OWNER',   '55555555-5555-5555-5555-555555555555', 'cccccccc-cccc-cccc-cccc-cccccccccccc'),
  ('STAFF',   '11111111-1111-1111-1111-111111111111', 'cccccccc-cccc-cccc-cccc-cccccccccccc');

-- =========================
-- EVENT WHITELISTS
-- =========================
INSERT INTO event_whitelists (event_id, attendee_ref_id)
VALUES
  ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 10001),
  ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 10003),
  ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 10005);

-- =========================
-- EVENT ALLOWED FACULTIES
-- =========================
INSERT INTO event_allowed_faculties (event_id, faculty_no)
VALUES
  ('cccccccc-cccc-cccc-cccc-cccccccccccc', 10),
  ('cccccccc-cccc-cccc-cccc-cccccccccccc', 20),
  ('cccccccc-cccc-cccc-cccc-cccccccccccc', 30);

-- =========================
-- EVENT AGENDAS
-- =========================
INSERT INTO event_agendas (event_id, activity_name, start_time, end_time)
VALUES
  -- Open Tech Talk
  (
  'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
  'Opening Session',
  NOW() - INTERVAL '1 hour',
  NOW()
  ),
  (
  'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
  'Tech Sharing',
  NOW(),
  NOW() + INTERVAL '2 hours'
  ),

  -- VIP Strategy
  (
  'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb',
  'Confidential Discussion',
  NOW(),
  NOW() + INTERVAL '3 hours'
  ),

  -- Faculty Day
  (
  'cccccccc-cccc-cccc-cccc-cccccccccccc',
  'Poster Presentation',
  NOW() + INTERVAL '2 days',
  NOW() + INTERVAL '2 days 4 hours'
  );

-- =========================
-- EVENT PARTICIPANTS
-- =========================
INSERT INTO event_participants (
  event_id,
  scanned_timestamp,
  checkin_timestamp,
  comment,
  participant_ref_id,
  first_name,
  sur_name,
  organization,
  scanned_location,
  scanner_id
  )
VALUES
  -- Scanned only (no check-in yet)
  (
  'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
  NOW() - INTERVAL '10 minutes',
  NULL,
  NULL,
  10002,
  'Somying',
  'Saendee',
  'Student Council',
  POINT(100.5018, 13.7563),
  '22222222-2222-2222-2222-222222222222'
  ),

  -- Scanned + checked in
  (
  'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
  NOW() - INTERVAL '30 minutes',
  NOW() - INTERVAL '25 minutes',
  'Arrived early',
  10003,
  'Withaya',
  'Kengngan',
  'Engineering Faculty',
  POINT(100.5020, 13.7565),
  '11111111-1111-1111-1111-111111111111'
  ),

  -- VIP whitelist attendee
  (
  'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb',
  NOW() - INTERVAL '5 minutes',
  NOW() - INTERVAL '3 minutes',
  'VIP guest',
  10001,
  'Somchai',
  'Jaidee',
  'Board Office',
  POINT(100.5030, 13.7570),
  '33333333-3333-3333-3333-333333333333'
  ),

  -- Faculty-based attendee
  (
  'cccccccc-cccc-cccc-cccc-cccccccccccc',
  NOW() + INTERVAL '2 days 10 minutes',
  NULL,
  'Late arrival',
  10005,
  'Thana',
  'Sukjai',
  'Science Faculty',
  POINT(100.5040, 13.7580),
  '55555555-5555-5555-5555-555555555555'
  );

COMMIT;

