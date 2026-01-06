BEGIN;

-- ====================================================
-- 1. SEED USERS
-- ====================================================
-- We hardcode UUIDs here so we can reference them in later inserts easily.
-- RefID must be unique uint64.

INSERT INTO users (id, ref_id, firstname_th, surname_th, title_th, firstname_en, surname_en, title_en)
VALUES 
    -- User 1: The Event Organizer
    ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a01', 1001, 'สมชาย', 'ใจดี', 'นาย', 'Somchai', 'Jaidee', 'Mr.'),
    
    -- User 2: The Staff Member (Scanner)
    ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a02', 1002, 'สมหญิง', 'จริงใจ', 'นางสาว', 'Somying', 'Jingjai', 'Ms.'),
    
    -- User 3: The Attendee (Student)
    ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a03', 6601001, 'นักเรียน', 'ตั้งใจ', 'นาย', 'Student', 'Tangjai', 'Mr.')
ON CONFLICT (id) DO NOTHING;


-- ====================================================
-- 2. SEED EVENTS
-- ====================================================

INSERT INTO events (
    id, 
    name, 
    organizer, 
    description, 
    date, 
    start_time, 
    end_time, 
    location, 
    attendence_type, 
    allow_all_to_scan, 
    evaluation_form, 
    revealed_fields
)
VALUES (
    'b0eebc99-9c0b-4ef8-bb6d-6bb9bd380b01', -- Event ID
    'CPE Tech Talk 2024',
    'Computer Engineering Dept.',
    'A deep dive into Go and Microservices.',
    NOW(), -- Timestamp for Date
    '09:00:00', -- Start Time
    '16:00:00', -- End Time
    'Auditorium A',
    'WHITELIST', -- Enum: attendence_type
    FALSE,
    'https://forms.google.com/example',
    '{NAME,ORGANIZATION}' -- Array of Enum: participant_data
)
ON CONFLICT (id) DO NOTHING;


-- ====================================================
-- 3. SEED EVENT ROLES (EventUser)
-- ====================================================

INSERT INTO event_users (role, user_id, event_id)
VALUES 
    -- User 1 is the OWNER
    ('OWNER', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a01', 'b0eebc99-9c0b-4ef8-bb6d-6bb9bd380b01'),
    
    -- User 2 is STAFF
    ('STAFF', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a02', 'b0eebc99-9c0b-4ef8-bb6d-6bb9bd380b01');


-- ====================================================
-- 4. SEED WHITELIST & FACULTIES
-- ====================================================

-- Whitelist User 3 (Attendee) for the event
INSERT INTO event_whitelists (event_id, attendee_ref_id)
VALUES 
    ('b0eebc99-9c0b-4ef8-bb6d-6bb9bd380b01', 6601001);

-- Allow Faculty Number 21 (Engineering)
INSERT INTO event_allowed_faculties (event_id, faculty_no)
VALUES 
    ('b0eebc99-9c0b-4ef8-bb6d-6bb9bd380b01', 21);


-- ====================================================
-- 5. SEED AGENDA
-- ====================================================

INSERT INTO event_agendas (event_id, activity_name, start_time, end_time)
VALUES 
    ('b0eebc99-9c0b-4ef8-bb6d-6bb9bd380b01', 'Registration', '09:00:00', '09:30:00'),
    ('b0eebc99-9c0b-4ef8-bb6d-6bb9bd380b01', 'Opening Ceremony', '09:30:00', '10:00:00'),
    ('b0eebc99-9c0b-4ef8-bb6d-6bb9bd380b01', 'Keynote: Golang', '10:00:00', '12:00:00');


-- ====================================================
-- 6. SEED PARTICIPANTS (Check-in records)
-- ====================================================

-- Simulating: User 3 walked in, scanned by User 2
INSERT INTO event_participants (
    event_id, 
    checkin_timestamp, 
    participant_ref_id, 
    first_name, 
    sur_name, 
    organization, 
    scanned_location, 
    scanner_id
)
VALUES (
    'b0eebc99-9c0b-4ef8-bb6d-6bb9bd380b01', -- Event ID
    NOW(), -- Checked in just now
    6601001, -- User 3 RefID
    'Student', -- Snapshotted name
    'Tangjai', 
    'Student Union',
    '(13.736717, 100.523186)', -- Point (Lat, Long) for Bangkok
    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a02' -- Scanned by User 2
);

COMMIT;
