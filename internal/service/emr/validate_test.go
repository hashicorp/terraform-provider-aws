package emr

import (
	"testing"
)

func TestValidCustomAMIID(t *testing.T) {
	cases := []struct {
		Value    string
		ErrCount int
	}{
		{
			Value:    "ami-dbcf88b1", //lintignore:AWSAT002
			ErrCount: 0,
		},
		{
			Value:    "vol-as7d65ash",
			ErrCount: 1,
		},
	}

	for _, tc := range cases {
		_, errors := validCustomAMIID(tc.Value, "custom_ami_id")

		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected %d errors, got %d: %s", tc.ErrCount, len(errors), errors)
		}
	}
}
