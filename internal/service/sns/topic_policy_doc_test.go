package sns

import (
	"testing"
)

func TestIsValidPrincipal(t *testing.T) {
	t.Parallel()

	testcases := map[string]struct {
		value string
		valid bool
	}{
		"role_arn": {
			value: "arn:aws:iam::123456789012:role/role-name", // lintignore:AWSAT005
			valid: true,
		},
		"root_arn": {
			value: "arn:aws:iam::123456789012:root", // lintignore:AWSAT005
			valid: true,
		},
		"account_id": {
			value: "123456789012",
			valid: true,
		},
		"unique_id": {
			value: "AROAS5MHDZS6NEXAMPLE",
			valid: false,
		},
	}

	for name, testcase := range testcases {
		testcase := testcase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			a := isValidPrincipal(testcase.value)

			if e := testcase.valid; a != e {
				t.Fatalf("expected %t, got %t", e, a)
			}
		})
	}
}
