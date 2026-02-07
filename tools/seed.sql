-- AI GENERATED SEED

BEGIN;

TRUNCATE TABLE 
    "event_participants", 
    "event_agendas", 
    "event_users", 
    "event_allowed_faculties", 
    "event_whitelists", 
    "events", 
    "users" 
RESTART IDENTITY CASCADE;

-- 2. INSERT USERS (Creating 5 personas)
-- We use explicit UUIDs to make testing relationships easier.
-- Ref IDs follow a pattern (e.g., 660001) to simulate student/staff IDs.

INSERT INTO "users" ("id", "ref_id", "firstname_th", "surname_th", "title_th", "firstname_en", "surname_en", "title_en") VALUES
-- User 1: The System Admin / Super Organizer
('00000000-0000-0000-0000-000000000001', 1001, 'สมชาย', 'ใจดี', 'นาย', 'Somchai', 'Jaidee', 'Mr'),
-- User 2: A Staff Member (Scanner)
('00000000-0000-0000-0000-000000000002', 1002, 'สมหญิง', 'จริงใจ', 'นางสาว', 'Somying', 'Jingjai', 'Ms'),
-- User 3: Student from Faculty 21 (Engineering)
('00000000-0000-0000-0000-000000000003', 660021, 'กล้า', 'หาญ', 'นาย', 'Kla', 'Han', 'Mr'),
-- User 4: Student from Faculty 22 (Arts)
('00000000-0000-0000-0000-000000000004', 660022, 'มานี', 'มีตา', 'นางสาว', 'Manee', 'Meeta', 'Ms'),
-- User 5: Random Participant
('00000000-0000-0000-0000-000000000005', 999999, 'ปิติ', 'พอใจ', 'นาย', 'Piti', 'Porjai', 'Mr');


-- 3. INSERT EVENTS (Creating 3 distinct scenarios)

INSERT INTO "events" 
("id", "name", "organizer", "description", "start_time", "end_time", "location", "attendence_type", "allow_all_to_scan", "evaluation_form", "revealed_fields") 
VALUES
-- Event A: VIP Gala (Whitelist only, strict privacy, owned by User 1)
('10000000-0000-0000-0000-000000000001', 
 'VIP Charity Gala', 
 'Central Admin', 
 'An exclusive fundraising dinner.', 
 NOW() + INTERVAL '1 day', 
 NOW() + INTERVAL '1 day 4 hours', 
 'Grand Hall', 
 'WHITELIST', 
 FALSE, -- Only staff can scan people in
 'https://forms.gle/vip-eval', 
 '{NAME, PHOTO}' -- Only reveal name and photo
),

-- Event B: Engineering Fair (Faculty restricted, students can scan themselves)
('10000000-0000-0000-0000-000000000002', 
 'Engineering Open House', 
 'Faculty of Engineering', 
 'Showcase of senior projects.', 
 NOW() - INTERVAL '2 days', 
 NOW() - INTERVAL '2 days 6 hours', 
 'Engineering Building 3', 
 'FACULTIES', 
 TRUE, -- Students can scan QR codes themselves
 NULL, 
 '{NAME, REFID, ORGANIZATION}'
),

-- Event C: Music Festival (Open to ALL)
('10000000-0000-0000-0000-000000000003', 
 'Campus Music Fest', 
 'Student Union', 
 'Music, Food, and Fun.', 
 NOW() + INTERVAL '5 days', 
 NOW() + INTERVAL '5 days 8 hours', 
 'Football Field', 
 'ALL', 
 TRUE, 
 'https://forms.gle/music-eval', 
 '{NAME, PHOTO}'
);


-- 4. CONFIGURE RULES (Whitelists & Faculties)

-- Event A (VIP): Add User 3 (Kla) to the whitelist.
INSERT INTO "event_whitelists" ("event_id", "attendee_ref_id") VALUES
('10000000-0000-0000-0000-000000000001', 660021); -- Kla is invited

-- Event B (Eng Fair): Allow Faculty 21 (Engineering).
INSERT INTO "event_allowed_faculties" ("event_id", "faculty_no") VALUES
('10000000-0000-0000-0000-000000000002', 21);


-- 5. EVENT AGENDAS (Schedule)

INSERT INTO "event_agendas" ("event_id", "activity_name", "start_time", "end_time") VALUES
-- Agenda for VIP Gala
('10000000-0000-0000-0000-000000000001', 'Registration', NOW() + INTERVAL '1 day', NOW() + INTERVAL '1 day 30 minutes'),
('10000000-0000-0000-0000-000000000001', 'Dinner Service', NOW() + INTERVAL '1 day 1 hour', NOW() + INTERVAL '1 day 3 hours');


-- 6. EVENT STAFFING (Roles)

INSERT INTO "event_users" ("event_id", "user_id", "role") VALUES
-- User 1 is the OWNER of VIP Gala
('10000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000001', 'OWNER'),
-- User 2 is STAFF for VIP Gala (can scan people)
('10000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000002', 'STAFF'),
-- User 3 is MANAGER of Engineering Fair
('10000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000003', 'MANAGER');


-- 7. PARTICIPATION (Simulating Scans)

INSERT INTO "event_participants" 
("event_id", "participant_id", "scanner_id", "organization", "scanned_location", "scanned_timestamp", "comment", "comment_timestamp") 
VALUES
-- Scenario 1: User 3 (Kla) attends VIP Gala. Scanned by Staff (User 2).
('10000000-0000-0000-0000-000000000001', 
 '00000000-0000-0000-0000-000000000003', -- Participant
 '00000000-0000-0000-0000-000000000002', -- Scanner (Staff)
 'Engineering Dept', 
 point(13.736717, 100.523186), -- Bangkok coordinates
 NOW() + INTERVAL '1 day 10 minutes', 
 'VIP arrived with guest', 
 NOW() + INTERVAL '1 day 15 minutes'
),

-- Scenario 2: User 3 (Kla) attends Eng Fair. Scanned by Self (Scanner_id is NULL or Self).
-- Note: Logic for self-scan usually implies scanner_id matches participant_id or is NULL depending on app logic.
('10000000-0000-0000-0000-000000000002', 
 '00000000-0000-0000-0000-000000000003', -- Participant
 '00000000-0000-0000-0000-000000000003', -- Self-scan
 'Engineering Dept', 
 point(13.736717, 100.523186), 
 NOW() - INTERVAL '2 days 1 hour', 
 NULL, NULL
);

COMMIT;
