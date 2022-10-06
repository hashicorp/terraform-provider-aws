package comprehend

import (
	"testing"
)

func TestValidateKMSKeyARN(t *testing.T) {
	testcases := map[string]struct {
		in    any
		valid bool
	}{
		"kms key id": {
			in:    "arn:aws:kms:us-west-2:123456789012:key/57ff7a43-341d-46b6-aee3-a450c9de6dc8", // lintignore:AWSAT003,AWSAT005
			valid: true,
		},
		"kms non-key id": {
			in:    "arn:aws:kms:us-west-2:123456789012:something/else", // lintignore:AWSAT003,AWSAT005
			valid: false,
		},
		"non-kms arn": {
			in:    "arn:aws:iam::123456789012:user/David", // lintignore:AWSAT005
			valid: false,
		},
		"not an arn": {
			in:    "not an arn",
			valid: false,
		},
		"not a string": {
			in:    123,
			valid: false,
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			aWs, aEs := validateKMSKeyARN(testcase.in, "field")
			if len(aWs) != 0 {
				t.Errorf("expected no warnings, got %v", aWs)
			}
			if testcase.valid {
				if len(aEs) != 0 {
					t.Errorf("expected no errors, got %v", aEs)
				}
			} else {
				if len(aEs) == 0 {
					t.Error("expected errors, got none")
				}
			}
		})
	}
}
