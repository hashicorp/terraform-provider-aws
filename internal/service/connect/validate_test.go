package connect

import (
	"testing"
)

func TestValidDeskPhoneNumber(t *testing.T) {
	validNumbers := []string{
		"+12345678912",
		"+6598765432",
	}
	for _, v := range validNumbers {
		_, errors := validDeskPhoneNumber(v, "desk_phone_number")
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid desk phone number: %q", v, errors)
		}
	}

	invalidNumbers := []string{
		"1232",
		"invalid",
	}
	for _, v := range invalidNumbers {
		_, errors := validDeskPhoneNumber(v, "desk_phone_number")
		if len(errors) == 0 {
			t.Fatalf("%q should be a invalid desk phone number: %q", v, errors)
		}
	}
}
