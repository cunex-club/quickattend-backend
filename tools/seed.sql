-- Uncomment if running manually, but 'make db-reset' handles this via schema clean.
-- ============================================================
-- TRUNCATE TABLE "event_participants", "event_agendas", "event_allowed_faculties", 
-- "event_whitelists", "event_users", "events", "users" RESTART IDENTITY CASCADE;

-- ============================================================
-- 1. SEED: USERS
-- We use fixed UUIDs so we can reliably link them in subsequent inserts.
-- ============================================================
INSERT INTO "users" (
    "id", "ref_id", 
    "firstname_th", "surname_th", "title_th", 
    "firstname_en", "surname_en", "title_en"
) VALUES 
-- [ADMINS / STAFF]
-- User 1: The Event Owner (RefID: 10001)
(
    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 10001,
    'สมชาย', 'ใจดี', 'ดร.',
    'Somchai', 'Jaidee', 'Dr.'
),
-- User 2: The Staff Member / Scanner (RefID: 10002)
(
    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a22', 10002,
    'สมหญิง', 'งานดี', 'นางสาว',
    'Somying', 'Ngandee', 'Ms.'
),

-- [PARTICIPANTS]
-- User 3: Engineering Student (RefID: 6630001 -> Faculty 21 Logic)
(
    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a33', 6630001,
    'วิศวะ', 'เกียร์มัว', 'นาย',
    'Visava', 'Gearmua', 'Mr.'
),
-- User 4: Arts Student (RefID: 6640001 -> Faculty 22 Logic)
(
    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a44', 6640001,
    'ศิลป์', 'ภาษา', 'นางสาว',
    'Silp', 'Pasa', 'Ms.'
),
-- User 5: External / General Public (RefID: 9900001)
(
    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a55', 9900001,
    'สมศักดิ์', 'รักเรียน', 'นาย',
    'Somsak', 'Rakrian', 'Mr.'
);

-- ============================================================
-- 2. SEED: EVENTS
-- We create one event for each 'attendence_type' to test all logic paths.
-- ============================================================
INSERT INTO "events" (
    "id", 
    "name", 
    "organizer", 
    "description", 
    "date", 
    "start_time", 
    "end_time", 
    "location", 
    "attendence_type", 
    "allow_all_to_scan", 
    "evaluation_form", 
    "revealed_fields"
) VALUES 
-- Event 1: PUBLIC EVENT (Type: ALL)
-- Test: Anyone should be able to check in.
(
    'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380b11',
    'Chula Open House 2025',
    'Central Admin',
    'Open for everyone.',
    '2025-10-15',
    '08:00:00',
    '17:00:00',
    'Main Auditorium',
    'ALL',
    true, -- "Self Check-in" enabled
    'https://forms.gle/openhouse',
    '{NAME,ORGANIZATION}'::participant_data[] -- Cast array literal to enum array
),

-- Event 2: PRIVATE EVENT (Type: WHITELIST)
-- Test: Only people in 'event_whitelists' can check in.
(
    'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380b22',
    'VIP Gala Dinner',
    'Dean Office',
    'Exclusive dinner for selected guests.',
    '2025-11-01',
    '18:00:00',
    '22:00:00',
    'Luxury Hotel Ballroom',
    'WHITELIST',
    false, -- Only Staff can scan
    NULL,
    '{NAME,PHOTO,REFID}'::participant_data[]
),

-- Event 3: FACULTY EVENT (Type: FACULTIES)
-- Test: Only students from specific faculties can check in.
(
    'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380b33',
    'Engineering Robot Workshop',
    'Faculty of Engineering',
    'Learning to build robots.',
    '2025-12-05',
    '09:00:00',
    '16:00:00',
    'Eng Building 3',
    'FACULTIES',
    false,
    'https://forms.gle/robot',
    '{NAME,REFID}'::participant_data[]
);

-- ============================================================
-- 3. SEED: EVENT USERS (Staff Permissions)
-- ============================================================
INSERT INTO "event_users" ("id", "role", "user_id", "event_id") VALUES 
-- 'Somchai' is OWNER of all events
(gen_random_uuid(), 'OWNER', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380b11'),
(gen_random_uuid(), 'OWNER', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380b22'),
(gen_random_uuid(), 'OWNER', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380b33'),

-- 'Somying' is STAFF for Open House & VIP Dinner (can scan people)
(gen_random_uuid(), 'STAFF', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a22', 'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380b11'),
(gen_random_uuid(), 'STAFF', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a22', 'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380b22');

-- ============================================================
-- 4. SEED: ACCESS CONTROLS
-- ============================================================
-- A. Whitelist for VIP Dinner (Event 2)
-- Only 'Visava' (Eng Student) and 'Somsak' (External) are invited.
-- 'Silp' (Arts Student) is NOT invited -> Should fail check-in.
INSERT INTO "event_whitelists" ("id", "event_id", "attendee_ref_id") VALUES
(gen_random_uuid(), 'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380b22', 6630001), -- Visava
(gen_random_uuid(), 'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380b22', 9900001); -- Somsak

-- B. Allowed Faculties for Engineering Workshop (Event 3)
-- Only Faculty 21 (Engineering) allowed.
INSERT INTO "event_allowed_faculties" ("id", "event_id", "faculty_no") VALUES
(gen_random_uuid(), 'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380b33', 21);

-- ============================================================
-- 5. SEED: AGENDAS
-- ============================================================
INSERT INTO "event_agendas" ("id", "event_id", "activity_name", "start_time", "end_time") VALUES 
(gen_random_uuid(), 'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380b11', 'Registration', '08:00:00', '09:00:00'),
(gen_random_uuid(), 'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380b11', 'Opening Ceremony', '09:00:00', '10:00:00'),
(gen_random_uuid(), 'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380b11', 'Campus Tour', '10:00:00', '12:00:00');

-- ============================================================
-- 6. SEED: PARTICIPANTS (Simulated Check-ins)
-- ============================================================
INSERT INTO "event_participants" (
    "id", 
    "event_id", 
    "checkin_timestamp", 
    "participant_ref_id", 
    "first_name", 
    "sur_name", 
    "organization", 
    "scanned_location", 
    "scanner_id"
) VALUES 
-- 1. Open House: Visava checked in by Staff (Somying)
(
    gen_random_uuid(), 
    'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380b11', 
    '2025-10-15 08:30:00+07', 
    6630001, 
    'Visava', 'Gearmua', 'Engineering',
    '(13.736717, 100.533118)', -- Point
    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a22' -- Scanned by Somying
),
-- 2. Open House: Silp checked in by Owner (Somchai)
(
    gen_random_uuid(), 
    'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380b11', 
    '2025-10-15 08:45:00+07', 
    6640001, 
    'Silp', 'Pasa', 'Arts',
    '(13.736717, 100.533118)', 
    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11' -- Scanned by Somchai
),
-- 3. VIP Dinner: Visava (Whitelisted) checked in successfully
(
    gen_random_uuid(), 
    'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380b22', 
    '2025-11-01 18:15:00+07', 
    6630001, 
    'Visava', 'Gearmua', 'Engineering',
    '(13.736717, 100.533118)', 
    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a22'
);
