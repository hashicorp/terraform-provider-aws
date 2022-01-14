package s3

import (
	"testing"
)

func TestValidBucketLifecycleTimestamp(t *testing.T) {
	validDates := []string{
		"2016-01-01",
		"2006-01-02",
	}

	for _, v := range validDates {
		_, errors := validBucketLifecycleTimestamp(v, "date")
		if len(errors) != 0 {
			t.Fatalf("%q should be valid date: %q", v, errors)
		}
	}

	invalidDates := []string{
		"Jan 01 2016",
		"20160101",
	}

	for _, v := range invalidDates {
		_, errors := validBucketLifecycleTimestamp(v, "date")
		if len(errors) == 0 {
			t.Fatalf("%q should be invalid date", v)
		}
	}
}
