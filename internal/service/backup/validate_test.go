package backup

import (
	"strings"
	"testing"
)

func TestValidReportPlanName(t *testing.T) {
	validNames := []string{
		strings.Repeat("W", 256), // <= 256
	}
	for _, v := range validNames {
		_, errors := validReportPlanName(v, "name")
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid Backup Report Plan name: %q", v, errors)
		}
	}

	invalidNames := []string{
		"@error",
		strings.Repeat("W", 257), // >= 257
	}
	for _, v := range invalidNames {
		_, errors := validReportPlanName(v, "name")
		if len(errors) == 0 {
			t.Fatalf("%q should be a invalid Backup Report Plan name: %q", v, errors)
		}
	}
}
