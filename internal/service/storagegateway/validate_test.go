package storagegateway

import (
	"testing"
)

func TestValid4ByteASN(t *testing.T) {
	validAsns := []string{
		"0",
		"1",
		"65534",
		"65535",
		"4294967294",
		"4294967295",
	}
	for _, v := range validAsns {
		_, errors := valid4ByteASN(v, "bgp_asn")
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid ASN: %q", v, errors)
		}
	}

	invalidAsns := []string{
		"-1",
		"ABCDEFG",
		"",
		"4294967296",
		"9999999999",
	}
	for _, v := range invalidAsns {
		_, errors := valid4ByteASN(v, "bgp_asn")
		if len(errors) == 0 {
			t.Fatalf("%q should be an invalid ASN", v)
		}
	}
}
