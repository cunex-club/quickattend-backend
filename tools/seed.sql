BEGIN;

-- =========================
-- USERS
-- =========================
INSERT INTO users (id, ref_id, firstname_th, surname_th, title_th, firstname_en, surname_en, title_en, faculty_name_th, faculty_name_en, profile_image_url)
VALUES
  ('11111111-1111-1111-1111-111111111111', 10001, 'สมชาย', 'ใจดี', 'นาย', 'Somchai', 'Jaidee', 'Mr.', 'คณะ ก', 'faculty A', 'https://'),
  ('22222222-2222-2222-2222-222222222222', 10002, 'สมหญิง', 'แสนดี', 'นางสาว', 'Somying', 'Saendee', 'Ms.', 'คณะ ข', 'faculty B', 'https://'),
  ('33333333-3333-3333-3333-333333333333', 10003, 'วิทยา', 'เก่งงาน', 'นาย', 'Withaya', 'Kengngan', 'Mr.', 'คณะ ค', 'faculty C', 'https://'),
  ('44444444-4444-4444-4444-444444444444', 10004, 'อรทัย', 'ตั้งใจ', 'นาง', 'Orathai', 'Tangjai', 'Mrs.', 'คณะ ง', 'faculty D', 'https://'),
  ('55555555-5555-5555-5555-555555555555', 6631321321, 'ธนกร', 'ไชยยุทธ', 'นาย', 'Thanagorn', 'Chaiyut', 'Mr.', 'คณะวิศวกรรมศาสตร์', 'faculty of engineering', '');
  -- ('66666666-6666-6666-6666-666666666666', 6638067221, 'ดวิษ', 'สุฤทธิ์ภิญโญ', 'นาย', 'Dawit', 'Surithpinyo', 'Mr.', '', '', '');
-- =========================
-- EVENTS
-- =========================
INSERT INTO events (
  id, name, organizer, description,
  start_time, end_time, location, location_point,
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
  POINT(100.5018, 13.7563),
  'ALL',
  true,
  'https://forms.example.com/open-tech',
  ARRAY['NAME', 'ORGANIZATION']::participant_data[]
  ),

  (
  'dddddddd-dddd-dddd-dddd-dddddddddddd',
  'Closed Tech Talk',
  'Engineering Club',
  'Public tech sharing session',
  NOW() - INTERVAL '1 hour',
  NOW() + INTERVAL '1 hour',
  'Main Hall',
  POINT(100.5018, 13.7563),
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
  POINT(100.5030, 13.7570),
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
  POINT(100.5040, 13.7580),
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
  ('OWNER',   '55555555-5555-5555-5555-555555555555', 'dddddddd-dddd-dddd-dddd-dddddddddddd'),
  ('STAFF',   '11111111-1111-1111-1111-111111111111', 'cccccccc-cccc-cccc-cccc-cccccccccccc');

-- =========================
-- EVENT WHITELISTS
-- =========================
INSERT INTO event_whitelists (event_id, attendee_ref_id)
VALUES
  ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 10001),
  ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 10003),
  ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 6631321321);

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
  comment_timestamp,
  comment,
  participant_id,
  organization,
  scanned_location,
  scanner_id
  )
VALUES
  -- Checked in, no comment
  (
  'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
  NOW() - INTERVAL '10 minutes',
  NULL,
  NULL,
  '22222222-2222-2222-2222-222222222222',
  'Student Council',
  POINT(100.5018, 13.7563),
  '22222222-2222-2222-2222-222222222222'
  ),

  -- Checked in + comment
  (
  'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
  NOW() - INTERVAL '30 minutes',
  NOW() - INTERVAL '25 minutes',
  'Arrived early',
  '33333333-3333-3333-3333-333333333333',
  'Engineering Faculty',
  POINT(100.5020, 13.7565),
  '11111111-1111-1111-1111-111111111111'
  ),

  -- VIP whitelist attendee (checked in + comment)
  (
  'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb',
  NOW() - INTERVAL '5 minutes',
  NOW() - INTERVAL '3 minutes',
  'VIP guest',
  '11111111-1111-1111-1111-111111111111',
  'Board Office',
  POINT(100.5030, 13.7570),
  '33333333-3333-3333-3333-333333333333'
  ),

  -- Faculty-based attendee (checked in + comment)
  (
  'cccccccc-cccc-cccc-cccc-cccccccccccc',
  NOW() + INTERVAL '2 days 10 minutes',
  NOW() + INTERVAL '2 days 12 minutes',
  'Late arrival',
  '55555555-5555-5555-5555-555555555555',
  'Science Faculty',
  POINT(100.5040, 13.7580),
  '55555555-5555-5555-5555-555555555555'
  );



-- =========================
-- ADDITIONAL PAST EVENTS FOR USER 111
-- =========================

WITH user_111_events AS ( 
  SELECT
      gen_random_uuid() AS id,
      'Past event ' || i || ' of user 111' AS name,
      'team 123' AS organizer,
      'Auto gen past event' AS description,
      NOW() - INTERVAL '2 hours' AS start_time,
      NOW() - INTERVAL '1 hour' AS end_time,
      'sjdkncdk' AS location,
      POINT(100.5050, 13.10) AS location_point,
      'ALL'::attendence_type AS attendence_type,
      TRUE AS allow_all_to_scan,
      'https eval ' || i AS evaluation_form,
      ARRAY['NAME', 'REFID', 'PHOTO']::participant_data[] AS revealed_fields,
      i AS sort_order
  FROM generate_series(1, 7) AS i
  UNION ALL
  SELECT
      gen_random_uuid() AS id,
      'Discovery event ' || i || ' for user 111' AS name,
      'dept xyz' AS organizer,
      'Auto gen discovery event' AS description,
      NOW() - INTERVAL '1 hour' AS start_time,
      NOW() + INTERVAL '2 hours' AS end_time,
      'building xyz' AS location,
      POINT(101, 13.668) AS location_point,
      'ALL'::attendence_type AS attendence_type,
      FALSE AS allow_all_to_scan,
      NULL AS evaluation_form,
      ARRAY['NAME', 'PHOTO']::participant_data[] AS revealed_fields,
      i + 7 AS sort_order
  FROM generate_series(1, 7) AS i
),
-- Insert 14 events into events table
inserted_events AS (
  INSERT INTO events (
    id, name, organizer, description,
    start_time, end_time, location, location_point,
    attendence_type, allow_all_to_scan,
    evaluation_form, revealed_fields
  )
  SELECT id, name, organizer, description,
    start_time, end_time, location, location_point,
    attendence_type, allow_all_to_scan,
    evaluation_form, revealed_fields 
  FROM user_111_events
),
user_111_events_with_rn AS (
  SELECT 
    id, 
    ROW_NUMBER() OVER (ORDER BY sort_order) AS rn
  FROM user_111_events
),
-- Insert into event_users
inserted_event_users AS (
  INSERT INTO event_users (role, user_id, event_id)

  -- Past Events
  SELECT
    'MANAGER'::role, 
    '11111111-1111-1111-1111-111111111111'::uuid,
    id FROM user_111_events_with_rn WHERE rn IN (1, 2, 3, 4)
  UNION ALL
  SELECT
    'OWNER'::role, 
    '33333333-3333-3333-3333-333333333333'::uuid,
    id FROM user_111_events_with_rn WHERE rn IN (1, 2, 3, 4)
  UNION ALL
  SELECT
    'STAFF'::role, 
    '11111111-1111-1111-1111-111111111111'::uuid,
    id FROM user_111_events_with_rn WHERE rn IN (5, 6, 7)
  UNION ALL
  SELECT
    'OWNER'::role, 
    '22222222-2222-2222-2222-222222222222'::uuid,
    id FROM user_111_events_with_rn WHERE rn IN (5, 6, 7)

  UNION ALL
  
  -- Discovery
  SELECT
    'OWNER'::role,
    '33333333-3333-3333-3333-333333333333'::uuid,
    id FROM user_111_events_with_rn WHERE rn IN (8, 9, 10, 11)
  UNION ALL
  SELECT
    'OWNER'::role,
    '22222222-2222-2222-2222-222222222222'::uuid,
    id FROM user_111_events_with_rn WHERE rn IN (12, 13, 14)
  UNION ALL
  SELECT
    'MANAGER'::role,
    '44444444-4444-4444-4444-444444444444'::uuid,
    id FROM user_111_events_with_rn WHERE rn IN (12, 13, 14)
  UNION ALL
  SELECT
    'STAFF'::role,
    '55555555-5555-5555-5555-555555555555'::uuid,
    id FROM user_111_events_with_rn WHERE rn IN (12, 13, 14)
),
-- Insert into event_agendas
agenda_inserted AS (
  INSERT INTO event_agendas (event_id, activity_name, start_time, end_time)

  SELECT
    id,
    'slot 1',
    NOW() - INTERVAL '2 hours',
    NOW() - INTERVAL '1 hour 30 minutes'
  FROM user_111_events_with_rn WHERE rn IN (1, 2, 3, 4, 5)
  UNION ALL
  SELECT
    id,
    'slot 2',
    NOW() - INTERVAL '1 hour 30 minutes',
    NOW() - INTERVAL '1 hour'
  FROM user_111_events_with_rn WHERE rn IN (1, 2, 3, 4, 5)
  UNION ALL
  SELECT
    id,
    'six seven 1',
    NOW() - INTERVAL '2 hours',
    NOW() - INTERVAL '1 hour 45 minutes'
  FROM user_111_events_with_rn WHERE rn IN (6, 7)
  UNION ALL
  SELECT
    id,
    'six seven 2',
    NOW() - INTERVAL '1 hour 45 minutes',
    NOW() - INTERVAL '1 hour 30 minutes'
  FROM user_111_events_with_rn WHERE rn IN (6, 7)
  UNION ALL
  SELECT
    id,
    'six seven 3',
    NOW() - INTERVAL '1 hour 30 minutes',
    NOW() - INTERVAL '1 hour'
  FROM user_111_events_with_rn WHERE rn IN (6, 7)
  UNION ALL
  SELECT
    id,
    'opening',
    NOW() - INTERVAL '1 hour',
    NOW() - INTERVAL '30 minutes'
  FROM user_111_events_with_rn WHERE rn IN (8, 9, 10, 11, 12, 13, 14)
  UNION ALL
  SELECT
    id,
    'something',
    NOW() - INTERVAL '30 minutes',
    NOW() + INTERVAL '1 hour 30 minutes'
  FROM user_111_events_with_rn WHERE rn IN (8, 9, 10, 11, 12, 13, 14)
  UNION ALL
  SELECT
    id,
    'closing',
    NOW() + INTERVAL '1 hour 30 minutes',
    NOW() + INTERVAL '2 hours'
  FROM user_111_events_with_rn WHERE rn IN (8, 9, 10, 11, 12, 13, 14)
)

-- Finally, insert into event_participants
INSERT INTO event_participants
(
  event_id,
  scanned_timestamp,
  comment_timestamp,
  comment,
  participant_id,
  organization,
  scanned_location,
  scanner_id
)
SELECT
  id,
  NOW() - INTERVAL '1 hour 30 minutes',
  NULL,
  NULL,
  '22222222-2222-2222-2222-222222222222'::uuid,
  'faculty of engineering',
  POINT(1.22, 888.2322),
  '11111111-1111-1111-1111-111111111111'::uuid
FROM user_111_events_with_rn WHERE rn IN (1, 2, 3, 4)
UNION ALL
SELECT
  id,
  NOW() - INTERVAL '1 hour 30 minutes',
  NOW() - INTERVAL '1 hour 29 minutes',
  'user 555 (ธนกร) joined',
  '55555555-5555-5555-5555-555555555555'::uuid,
  'faculty of engineering',
  POINT(44.222, 7.77),
  '11111111-1111-1111-1111-111111111111'::uuid
FROM user_111_events_with_rn WHERE rn IN (1, 2, 3, 4)
UNION ALL
SELECT
  id,
  NOW() - INTERVAL '1 hour 55 minutes',
  NULL,
  NULL,
  '44444444-4444-4444-4444-444444444444'::uuid,
  'faculty of science',
  POINT(0.01, 379.44),
  '22222222-2222-2222-2222-222222222222'::uuid
FROM user_111_events_with_rn WHERE rn IN (5, 6, 7)
UNION ALL
SELECT
  id,
  NOW() - INTERVAL '47 minutes',
  NULL,
  NULL,
  '22222222-2222-2222-2222-222222222222'::uuid,
  'faculty of fine and applied arts',
  POINT(11.11, 33.33),
  '33333333-3333-3333-3333-333333333333'::uuid
FROM user_111_events_with_rn WHERE rn IN (8, 9, 10, 11)
UNION ALL
SELECT
  id,
  NOW() - INTERVAL '47 minutes',
  NOW() - INTERVAL '45 minutes',
  'user 444 (อรทัย) joined',
  '44444444-4444-4444-4444-444444444444'::uuid,
  'faculty of political science',
  POINT(11.12, 33.33),
  '33333333-3333-3333-3333-333333333333'::uuid
FROM user_111_events_with_rn WHERE rn IN (8, 9, 10, 11)
UNION ALL
SELECT
  id,
  NOW() - INTERVAL '59 minutes',
  NOW() - INTERVAL '55 minutes',
  'user 555 (ธนกร) joined',
  '55555555-5555-5555-5555-555555555555'::uuid,
  'faculty of engineering',
  POINT(11.12, 33.33),
  '33333333-3333-3333-3333-333333333333'::uuid
FROM user_111_events_with_rn WHERE rn IN (8, 9, 10, 11)
UNION ALL
SELECT
  id,
  NOW() - INTERVAL '2 minutes',
  NULL,
  NULL,
  '33333333-3333-3333-3333-333333333333'::uuid,
  'faculty of psychology',
  POINT(52.0, 3.33),
  '55555555-5555-5555-5555-555555555555'::uuid
FROM user_111_events_with_rn WHERE rn IN (12, 13, 14);


/*
Summary of additional events for user 111

***** 7 more past events: name = "Past event i of user 111"
- Already ended

For i = 1 to 4
- users
  - OWNER = 333 (วิทยา)
  - MANAGER = 111 (สมชาย)
- 2 participants
  - 111 scanned 222, no comment
  - 111 scanned 555, has comment

For i = 5 to 7
- users
  - OWNER = 222 (สมหญิง)
  - STAFF = 111 (สมชาย)
- 1 participant
  - 222 scanned 444, no comment


***** 7 more discovery events: name = "Discovery event i of user 111" 
- Ongoing

For i = 1 to 4
- users
  - OWNER = 333 (วิทยา)
- 3 participants
  - 333 scanned 222, no comment
  - 333 scanned 444, has comment
  - 333 scanned 555, has comment

For i = 5 to 7
- users
  - OWNER = 222 (สมหญิง)
  - MANAGER = 444 (อรทัย)
  - STAFF = 555 (ธนกร)
- 1 participant
  - 555 scanned 333, no comment
*/



-- More events for testing whitelist pendings and event user pendings
INSERT INTO events (
  id, name, organizer, description,
  start_time, end_time, location, location_point,
  attendence_type, allow_all_to_scan,
  evaluation_form, revealed_fields
)
VALUES
  (
    '5d7fbc22-ac19-4c70-95ab-6f7cf1056867'::uuid,
    'Event abc with whitelist system',
    'abc org',
    NULL,
    NOW() - INTERVAL '2 hours',
    NOW() + INTERVAL '5 hours',
    'location of abc',
    POINT(10, 50),
    'WHITELIST'::attendence_type,
    FALSE,
    NULL,
    ARRAY['PHOTO', 'NAME', 'REFID', 'ORGANIZATION']::participant_data[]
  ),
  (
    '40ae65ef-9b95-4031-ab00-257ab7cbdf70'::uuid,
    'Event jjjj with whitelist system',  
    'jjjj',
    'nothing',
    NOW() - INTERVAL '1 hour',
    NOW() + INTERVAL '4 hours',
    'location of jjjj',
    POINT(0, 3.33),
    'WHITELIST'::attendence_type,
    TRUE,
    NULL,
    ARRAY['NAME', 'REFID', 'ORGANIZATION']::participant_data[]
  );

-- Whitelist people with account in system
INSERT INTO event_whitelists (event_id, attendee_ref_id)
VALUES
  ('5d7fbc22-ac19-4c70-95ab-6f7cf1056867'::uuid, 10001),
  ('5d7fbc22-ac19-4c70-95ab-6f7cf1056867'::uuid, 6631321321),
  ('40ae65ef-9b95-4031-ab00-257ab7cbdf70'::uuid, 10002),
  ('40ae65ef-9b95-4031-ab00-257ab7cbdf70'::uuid, 10003);

-- Whitelist people without account
INSERT INTO event_whitelist_pendings (event_id, attendee_ref_id)
VALUES
  ('5d7fbc22-ac19-4c70-95ab-6f7cf1056867'::uuid, 99999),
  ('5d7fbc22-ac19-4c70-95ab-6f7cf1056867'::uuid, 6638067221),
  ('40ae65ef-9b95-4031-ab00-257ab7cbdf70'::uuid, 888888);


-- Event staff with account in system
INSERT INTO event_users (event_id, user_id, role)
VALUES
  ('5d7fbc22-ac19-4c70-95ab-6f7cf1056867'::uuid, '44444444-4444-4444-4444-444444444444'::uuid, 'OWNER'),
  ('40ae65ef-9b95-4031-ab00-257ab7cbdf70'::uuid, '11111111-1111-1111-1111-111111111111'::uuid, 'OWNER'),
  ('40ae65ef-9b95-4031-ab00-257ab7cbdf70'::uuid, '44444444-4444-4444-4444-444444444444'::uuid, 'OWNER');

-- Event staff without account
INSERT INTO event_user_pendings (event_id, user_ref_id, role)
VALUES
  ('40ae65ef-9b95-4031-ab00-257ab7cbdf70'::uuid, 6638067221, 'STAFF'),
  ('5d7fbc22-ac19-4c70-95ab-6f7cf1056867'::uuid, 222222, 'MANAGER');

INSERT INTO event_agendas (event_id, activity_name, start_time, end_time)
VALUES
  (
    '5d7fbc22-ac19-4c70-95ab-6f7cf1056867'::uuid, 
    'slot1',
    NOW() - INTERVAL '2 hours',
    NOW() + INTERVAL '1 hour'
  ),
  (
    '5d7fbc22-ac19-4c70-95ab-6f7cf1056867'::uuid, 
    'slot2',
    NOW() + INTERVAL '1 hour',
    NOW() + INTERVAL '5 hours'
  ),
  (
    '40ae65ef-9b95-4031-ab00-257ab7cbdf70'::uuid,
    'slot1',
    NOW() - INTERVAL '1 hour',
    NOW() + INTERVAL '2 hours'
  ),
  (
    '40ae65ef-9b95-4031-ab00-257ab7cbdf70'::uuid,
    'slot2',
    NOW() + INTERVAL '2 hours',
    NOW() + INTERVAL '4 hours'   
  );


COMMIT;