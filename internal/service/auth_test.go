package service

import "testing"

func TestIsAllowedRefID(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name    string
		refID   string
		allowed string
		want    bool
	}{
		{name: "student", refID: "6612345678", allowed: "12345678, 6612345678", want: true},
		{name: "staff leading zero", refID: "00123456", allowed: "00123456", want: true},
		{name: "not allowed", refID: "99999999", allowed: "12345678,6612345678", want: false},
		{name: "empty allowlist", refID: "12345678", allowed: "", want: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := isAllowedRefID(tc.refID, tc.allowed); got != tc.want {
				t.Fatalf("isAllowedRefID(%q, %q) = %v, want %v", tc.refID, tc.allowed, got, tc.want)
			}
		})
	}
}
