package service

import "testing"

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
