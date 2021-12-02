package ec2

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"testing"
)

func TestValidSecurityGroupRuleDescription(t *testing.T) {
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

func TestValidateIPv4OrIPv6(t *testing.T) {
	validator := validateIPv4OrIPv6(
		validation.IsCIDRNetwork(16, 24),
		validation.IsCIDRNetwork(40, 64),
	)
	validCIDRs := []string{
		"10.0.0.0/16", // IPv4 CIDR /16 >= /16 and <= /24
		"10.0.0.0/23", // IPv4 CIDR /23 >= /16 and <= /24
		"10.0.0.0/24", // IPv4 CIDR /24 >= /16 and <= /24
		"2001::/40",   // IPv6 CIDR /40 >= /40 and <= /64
		"2001::/63",   // IPv6 CIDR /63 >= /40 and <= /64
		"2001::/64",   // IPv6 CIDR /64 >= /40 and <= /64
	}

	for _, v := range validCIDRs {
		_, errors := validator(v, "cidr_block")
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid CIDR block: %q", v, errors)
		}
	}

	invalidCIDRs := []string{
		"ASDQWE",      // not IPv4 nor IPv6 CIDR
		"0.0.0.0/0",   // IPv4 CIDR /0 < /16
		"10.0.0.0/8",  // IPv4 CIDR /8 < /16
		"10.0.0.1/24", // IPv4 CIDR with invalid network part
		"10.0.0.0/25", // IPv4 CIDR /25 > /24
		"10.0.0.0/32", // IPv4 CIDR /32 > /24
		"::/0",        // IPv6 CIDR /0 < /40
		"2001::/30",   // IPv6 CIDR /30 < /40
		"2001::1/64",  // IPv6 CIDR with invalid network part
		"2001::/65",   // IPv6 CIDR /65 > /64
		"2001::/128",  // IPv6 CIDR /128 > /64
	}

	for _, v := range invalidCIDRs {
		_, errors := validator(v, "cidr_block")
		if len(errors) == 0 {
			t.Fatalf("%q should be an invalid CIDR block", v)
		}
	}
}
