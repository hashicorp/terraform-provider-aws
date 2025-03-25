package types

import "testing"

func TestIsAWSOrganizationOUID(t *testing.T) { // nosemgrep:ci.aws-in-func-name
	t.Parallel()

	for _, tc := range []struct {
		id    string
		valid bool
	}{
		{id: "", valid: false},
		{id: "r-abc", valid: false},
		{id: "ou-z7jt", valid: false},
		{id: "ou-z7jt-19mqs9sp", valid: true},
	} {
		ok := IsAWSOrganizationOUID(tc.id)
		if got, want := ok, tc.valid; got != want {
			t.Errorf("IsAWSOrganizationOUID(%q) = %v; want %v", tc.id, got, want)
		}
	}
}
