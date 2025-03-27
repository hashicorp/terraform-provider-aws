package types

import "testing"

func TestIsAWSOrganizationRootID(t *testing.T) { // nosemgrep:ci.aws-in-func-name
	t.Parallel()

	for _, tc := range []struct {
		id    string
		valid bool
	}{
		{id: "", valid: false},
		{id: "r-abc", valid: false},
		{id: "-e3zd", valid: false},
		{id: "r-y7zf", valid: true},
	} {
		ok := IsAWSOrganizationRootID(tc.id)
		if got, want := ok, tc.valid; got != want {
			t.Errorf("IsAWSOrganizationRootID(%q) = %v; want %v", tc.id, got, want)
		}
	}
}
