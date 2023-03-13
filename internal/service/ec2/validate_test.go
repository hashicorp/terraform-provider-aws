package ec2

import (
	"testing"
)

func TestValidSecurityGroupRuleDescription(t *testing.T) {
	t.Parallel()

	validDescriptions := []string{
		"testrule",
		"testRule",
		"testRule 123",
		`testRule 123 ._-:/()#,@[]+=&;{}!$*`,
	}
	for _, v := range validDescriptions {
		_, errors := validSecurityGroupRuleDescription(v, "description")
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid security group rule description: %q", v, errors)
		}
	}

	invalidDescriptions := []string{
		"`",
		"%%",
		`\`,
	}
	for _, v := range invalidDescriptions {
		_, errors := validSecurityGroupRuleDescription(v, "description")
		if len(errors) == 0 {
			t.Fatalf("%q should be an invalid security group rule description", v)
		}
	}
}

func TestValidAmazonSideASN(t *testing.T) {
	t.Parallel()

	validAsns := []string{
		"7224",
		"9059",
		"10124",
		"17493",
		"64512",
		"64513",
		"65533",
		"65534",
		"4200000000",
		"4200000001",
		"4294967293",
		"4294967294",
	}
	for _, v := range validAsns {
		_, errors := validAmazonSideASN(v, "amazon_side_asn")
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid ASN: %q", v, errors)
		}
	}

	invalidAsns := []string{
		"1",
		"ABCDEFG",
		"",
		"7225",
		"9058",
		"10125",
		"17492",
		"64511",
		"65535",
		"4199999999",
		"4294967295",
		"9999999999",
	}
	for _, v := range invalidAsns {
		_, errors := validAmazonSideASN(v, "amazon_side_asn")
		if len(errors) == 0 {
			t.Fatalf("%q should be an invalid ASN", v)
		}
	}
}
