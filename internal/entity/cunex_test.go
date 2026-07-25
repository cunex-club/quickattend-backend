package entity

import "testing"

func TestParseUserType(t *testing.T) {
	t.Parallel()

	for input, want := range map[string]UserTypes{
		"student": STUDENTS,
		" Staff ": STAFFS,
	} {
		got, ok := ParseUserType(input)
		if !ok || got != want {
			t.Fatalf("ParseUserType(%q) = %q, %v; want %q, true", input, got, ok, want)
		}
	}

	if _, ok := ParseUserType("unknown"); ok {
		t.Fatal("ParseUserType(unknown) accepted an unsupported user type")
	}
}
