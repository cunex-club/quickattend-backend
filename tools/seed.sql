BEGIN;

INSERT INTO users (id, ref_id, firstname_th, surname_th, title_th, firstname_en, surname_en, title_en)
VALUES
    ('cb11ed5c-6d11-43f1-b004-a4625c7551df', 6631234521, 'ก', 'ข', 'นาย', 'A', 'B', 'MR'),
    ('a99e1c32-ac26-4bb3-9add-340f57749d87', 6874440950, 'ค', 'ง', 'นางสาว', 'C', 'D', 'MS'),
    ('cb147cf1-3d70-403c-9590-89371001623a', 12345678, 'จจจ', 'กขค', 'นาย', 'jjj', 'abc', 'MR'),
    ('0b76a343-1e72-4882-bfe8-124c36cf5ebb', 04498344, 'รรรร', 'ขค', 'นาง', 'rrrr', 'kk', 'MRS'),
    ('28e55bae-89e1-4c56-b5d2-7470284b2e2c', 987654321, 'AB', 'CD', 'EEEE', 'FG', 'HI', 'JJJJ');

INSERT INTO events (
    id,
    name,
    organizer,
    description,
    start_time,
    end_time,
    location,
    attendence_type,
    allow_all_to_scan,
    evaluation_form,
    revealed_fields
)
VALUES
    (
        '7525aa0f-0cb9-4e08-9503-f005d38c5151',
        'freshmen night',
        'sgcu',
        NULL,
        '2025-08-01 12:00:00+00',
        '2025-08-01 14:00:00+00',
        'ศาลาพระเกี้ยว',
        'ALL',
        true,
        'xyz',
        '{NAME, REFID}'
    ),
    (
        'bfc840ff-0051-42a7-88e7-252cd0458839',
        'aa',
        'คณะวิทย์',
        'lmnopqrstuv',
        '2026-12-12 03:00:00+00',
        '2026-12-12 09:00:00+00',
        'มหามกุฏ',
        'WHITELIST',
        false,
        NULL,
        '{NAME, ORGANIZATION, PHOTO}'
    ),
    (
        '51fb7771-562d-450e-b6b2-bf7af8060e7a',
        'DROP-TABLE',
        '123xxx',
        'idk',
        '2026-05-28 23:00:00+00',
        '2026-05-29 11:00:00+00',
        'random place',
        'FACULTIES',
        false,
        'https://abc',
        '{NAME}'
    ),
    (
        '3cf8ca23-b68c-48a8-9f97-ffeb7c42f176',
        'meeting789',
        'pmcu',
        'รายละเอียด',
        '2027-01-13 13:00:00+00',
        '2027-01-13 16:00:00+00',
        'pmcu building room whatever',
        'WHITELIST',
        true,
        NULL,
        '{NAME, ORGANIZATION, PHOTO}'
    ),
    (
        '8d24e04d-4856-49af-adc5-bdafb1ace72c',
        'ประชุมประจำปี',
        'คณะ?',
        NULL,
        '2028-03-07 02:00:00+00',
        '2028-03-07 06:00:00+00',
        '00000',
        'FACULTIES',
        true,
        ';?:',
        '{}'
    );


INSERT INTO event_whitelists (event_id, attendee_ref_id)
VALUES
    ('bfc840ff-0051-42a7-88e7-252cd0458839', 6631234521),
    ('bfc840ff-0051-42a7-88e7-252cd0458839', 12345678),
    ('3cf8ca23-b68c-48a8-9f97-ffeb7c42f176', 4498344),
    ('3cf8ca23-b68c-48a8-9f97-ffeb7c42f176', 6874440950),
    ('3cf8ca23-b68c-48a8-9f97-ffeb7c42f176', 6631234521);

INSERT INTO event_allowed_faculties (event_id, faculty_no)
VALUES
    ('51fb7771-562d-450e-b6b2-bf7af8060e7a', 21),
    ('51fb7771-562d-450e-b6b2-bf7af8060e7a', 32),
    ('8d24e04d-4856-49af-adc5-bdafb1ace72c', 50);

INSERT INTO event_agendas (event_id, activity_name, start_time, end_time)
VALUES 
    (
        '7525aa0f-0cb9-4e08-9503-f005d38c5151',
        'opening',
        '2025-08-01 12:00:00+00',
        '2025-08-01 12:30:00+00'
    ),
    (
        '7525aa0f-0cb9-4e08-9503-f005d38c5151',
        'slot2',
        '2025-08-01 12:30:00+00',
        '2025-08-01 13:00:00+00'
    ),
    (
        '7525aa0f-0cb9-4e08-9503-f005d38c5151',
        'slot3',
        '2025-08-01 13:00:00+00',
        '2025-08-01 14:00:00+00'
    ),
    (
        'bfc840ff-0051-42a7-88e7-252cd0458839',
        'bb',
        '2026-12-12 03:00:00+00',
        '2026-12-12 06:00:00+00'
    ),
    (
        'bfc840ff-0051-42a7-88e7-252cd0458839',
        'cc',
        '2026-12-12 06:00:00+00',
        '2026-12-12 09:00:00+00'
    );


INSERT INTO event_participants (
    event_id,
    checkin_timestamp,
    scanned_timestamp,
    comment,
    participant_ref_id,
    first_name,
    sur_name,
    organization,
    scanned_location,
    scanner_id
)
VALUES
    (
        'bfc840ff-0051-42a7-88e7-252cd0458839',
        NULL,
        '2026-12-12 02:56:00+00',
        NULL,
        6631234521,
        'A',
        'B',
        'faculty of engineering',
        '(3.14,5.55)',
        'cb11ed5c-6d11-43f1-b004-a4625c7551df'
    ),
    (
        'bfc840ff-0051-42a7-88e7-252cd0458839',
        '2026-12-12 01:23:36+00',
        '2026-12-12 01:23:00+00',
        'comment 123',
        12345678,
        'jjj',
        'abc',
        'lol',
        '(4.444,999.88)',
        'cb11ed5c-6d11-43f1-b004-a4625c7551df'
    ),
    (
        '8d24e04d-4856-49af-adc5-bdafb1ace72c',
        '2028-03-07 02:31:02+00',
        '2028-03-07 02:30:22+00',
        NULL,
        6874440950,
        'C',
        'D',
        'smth',
        '(22.90884,0.23589)',
        'a99e1c32-ac26-4bb3-9add-340f57749d87'
    ),
    (
        '7525aa0f-0cb9-4e08-9503-f005d38c5151',
        NULL,
        '2025-08-01 11:43:56+00',
        NULL,
        6631234521,
        'A',
        'B',
        'faculty of engineering',
        '(1.11,39.0)',
        'cb11ed5c-6d11-43f1-b004-a4625c7551df'
    ),
    (
        '3cf8ca23-b68c-48a8-9f97-ffeb7c42f176',
        '2027-01-13 13:04:22+00',
        '2027-01-13 13:04:01+00',
        'rrrr checked in',
        4498344,
        'rrrr',
        'kk',
        'hhhhasxhbhjbh',
        '(12.09,45.333)',
        'cb11ed5c-6d11-43f1-b004-a4625c7551df'
    );


INSERT INTO event_users (role, user_id, event_id)
VALUES
    ('MANAGER', 'cb11ed5c-6d11-43f1-b004-a4625c7551df', '7525aa0f-0cb9-4e08-9503-f005d38c5151'),
    ('OWNER', 'a99e1c32-ac26-4bb3-9add-340f57749d87', '3cf8ca23-b68c-48a8-9f97-ffeb7c42f176'),
    ('STAFF', 'a99e1c32-ac26-4bb3-9add-340f57749d87', '7525aa0f-0cb9-4e08-9503-f005d38c5151');

COMMIT;