package service

import (
	"bytes"
	"testing"
	"time"

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
