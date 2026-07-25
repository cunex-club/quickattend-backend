package service

import (
	"bytes"
	"testing"
	"time"

	dtoReq "github.com/cunex-club/quickattend-backend/internal/dto/request"
	"github.com/cunex-club/quickattend-backend/internal/entity"
	"github.com/xuri/excelize/v2"
)

func TestFirstNonBlank(t *testing.T) {
	t.Parallel()

	blank := " "
	faculty := "Engineering"

	if got := firstNonBlank(nil, &blank, &faculty); got == nil || *got != faculty {
		t.Fatalf("firstNonBlank() = %v, want %q", got, faculty)
	}
	if got := firstNonBlank(nil, &blank); got != nil {
		t.Fatalf("firstNonBlank() = %q, want nil", *got)
	}
}

func TestParseFacultyOrgCode(t *testing.T) {
	t.Parallel()

	numeric := "26"
	code, rejected := parseFacultyOrgCode(&numeric)
	if code == nil || *code != 26 || rejected != "" {
		t.Fatalf("parseFacultyOrgCode(%q) = (%v, %q), want (26, \"\")", numeric, code, rejected)
	}

	// CU NEX can return letter codes for newly-created cross-faculty programs
	// (e.g. "BG"). This must not be treated as an error — it just matches no
	// faculty.
	letters := "BG"
	code, rejected = parseFacultyOrgCode(&letters)
	if code != nil || rejected != "BG" {
		t.Fatalf("parseFacultyOrgCode(%q) = (%v, %q), want (nil, \"BG\")", letters, code, rejected)
	}

	// Staff members get a nil facultyCode from CU NEX.
	code, rejected = parseFacultyOrgCode(nil)
	if code != nil || rejected != "" {
		t.Fatalf("parseFacultyOrgCode(nil) = (%v, %q), want (nil, \"\")", code, rejected)
	}

	empty := "  "
	code, rejected = parseFacultyOrgCode(&empty)
	if code != nil || rejected != "" {
		t.Fatalf("parseFacultyOrgCode(%q) = (%v, %q), want (nil, \"\")", empty, code, rejected)
	}
}

func TestStripOwnerEntriesRejectsClientSuppliedOwner(t *testing.T) {
	t.Parallel()

	// A MANAGER (who is allowed to call UpdateEvent) must never be able to
	// grant OWNER to themselves or anyone else via the request body — owner
	// is always re-derived from the DB by the caller after this strip.
	in := []dtoReq.ManagerStaffReq{
		{RefID: 111, Role: string(entity.MANAGER)},
		{RefID: 222, Role: string(entity.OWNER)}, // attacker-controlled
		{RefID: 333, Role: string(entity.STAFF)},
	}

	got := stripOwnerEntries(in)

	if len(got) != 2 {
		t.Fatalf("stripOwnerEntries() returned %d entries, want 2: %#v", len(got), got)
	}
	for _, person := range got {
		if person.Role == string(entity.OWNER) {
			t.Fatalf("stripOwnerEntries() kept an OWNER entry: %#v", got)
		}
	}

	// Must not mutate the caller's backing array (append(entries[:0:0], ...)
	// only reuses the len/cap trick to allocate fresh, not to alias `in`).
	if len(in) != 3 || in[1].Role != string(entity.OWNER) {
		t.Fatalf("stripOwnerEntries() mutated its input: %#v", in)
	}
}

func TestBuildParticipantWorkbookHandlesNullStaffFields(t *testing.T) {
	t.Parallel()

	firstName := "Som"
	content, err := buildParticipantWorkbook(&entity.EventParticipantExportData{
		EventName: "Staff event",
		StartTime: time.Date(2026, 7, 25, 0, 0, 0, 0, time.UTC),
		Rows: []entity.EventParticipantExportRow{{
			ScannedTimestamp: time.Date(2026, 7, 25, 3, 4, 5, 0, time.UTC),
			UserType:         entity.STAFFS,
			RefID:            12345,
			FirstnameEN:      &firstName,
		}},
	})
	if err != nil {
		t.Fatal(err)
	}

	book, err := excelize.OpenReader(bytes.NewReader(content))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = book.Close() }()

	rows, err := book.GetRows("Participants")
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 {
		t.Fatalf("row count = %d, want 2", len(rows))
	}
	if rows[1][1] != "staff" || rows[1][2] != "00012345" || rows[1][4] != "Som" {
		t.Fatalf("unexpected export row: %#v", rows[1])
	}
}
