package aws

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func Test_expandNetworkACLEntry(t *testing.T) {
	input := []interface{}{
		map[string]interface{}{
			"protocol":   "tcp",
			"from_port":  22,
			"to_port":    22,
			"cidr_block": "0.0.0.0/0",
			"action":     "deny",
			"rule_no":    1,
		},
		map[string]interface{}{
			"protocol":   "tcp",
			"from_port":  443,
			"to_port":    443,
			"cidr_block": "0.0.0.0/0",
			"action":     "deny",
			"rule_no":    2,
		},
		map[string]interface{}{
			"protocol":   "-1",
			"from_port":  443,
			"to_port":    443,
			"cidr_block": "0.0.0.0/0",
			"action":     "deny",
			"rule_no":    2,
		},
	}
	expanded, _ := expandNetworkAclEntries(input, "egress")

	expected := []*ec2.NetworkAclEntry{
		{
			Protocol: aws.String("6"),
			PortRange: &ec2.PortRange{
				From: aws.Int64(22),
				To:   aws.Int64(22),
			},
			RuleAction: aws.String("deny"),
			RuleNumber: aws.Int64(1),
			CidrBlock:  aws.String("0.0.0.0/0"),
			Egress:     aws.Bool(true),
		},
		{
			Protocol: aws.String("6"),
			PortRange: &ec2.PortRange{
				From: aws.Int64(443),
				To:   aws.Int64(443),
			},
			RuleAction: aws.String("deny"),
			RuleNumber: aws.Int64(2),
			CidrBlock:  aws.String("0.0.0.0/0"),
			Egress:     aws.Bool(true),
		},
		{
			Protocol: aws.String("-1"),
			PortRange: &ec2.PortRange{
				From: aws.Int64(443),
				To:   aws.Int64(443),
			},
			RuleAction: aws.String("deny"),
			RuleNumber: aws.Int64(2),
			CidrBlock:  aws.String("0.0.0.0/0"),
			Egress:     aws.Bool(true),
		},
	}

	if !reflect.DeepEqual(expanded, expected) {
		t.Fatalf(
			"Got:\n\n%#v\n\nExpected:\n\n%#v\n",
			expanded,
			expected)
	}

}

func Test_validatePorts(t *testing.T) {
	for _, ts := range []struct {
		to       int64
		from     int64
		expected *expectedPortPair
		wanted   bool
	}{
		{0, 0, &expectedPortPair{0, 0}, true},
		{0, 1, &expectedPortPair{0, 0}, false},
	} {
		got := validatePorts(ts.to, ts.from, *ts.expected)
		if got != ts.wanted {
			t.Fatalf("Got: %t; Expected: %t\n", got, ts.wanted)
		}
	}
}

func Test_cidrBlocksEqual(t *testing.T) {
	for _, ts := range []struct {
		cidr1 string
		cidr2 string
		equal bool
	}{
		{"10.2.2.0/24", "10.2.2.0/24", true},
		{"10.2.2.0/1234", "10.2.2.0/24", false},
		{"10.2.2.0/24", "10.2.2.0/1234", false},
		{"2001::/15", "2001::/15", true},
		{"::/0", "2001::/15", false},
		{"::/0", "::0/0", true},
		{"", "", false},
	} {
		equal := cidrBlocksEqual(ts.cidr1, ts.cidr2)
		if ts.equal != equal {
			t.Fatalf("cidrBlocksEqual(%q, %q) should be: %t", ts.cidr1, ts.cidr2, ts.equal)
		}
	}
}

func Test_validateCIDRBlock(t *testing.T) {
	for _, ts := range []struct {
		cidr  string
		valid bool
	}{
		{"10.2.2.0/24", true},
		{"10.2.2.0/1234", false},
		{"10.2.2.2/24", false},
		{"::/0", true},
		{"::0/0", true},
		{"2000::/15", true},
		{"2001::/15", false},
		{"", false},
	} {
		err := validateCIDRBlock(ts.cidr)
		if !ts.valid && err == nil {
			t.Fatalf("Input '%s' should error but didn't!", ts.cidr)
		}
		if ts.valid && err != nil {
			t.Fatalf("Got unexpected error for '%s' input: %s", ts.cidr, err)
		}
	}
}

func Test_validateIpv4CIDRBlock(t *testing.T) {
	for _, ts := range []struct {
		cidr  string
		valid bool
	}{
		{"10.2.2.0/24", true},
		{"10.2.2.0/1234", false},
		{"10/24", false},
		{"10.2.2.2/24", false},
		{"::/0", false},
		{"2000::/15", false},
		{"", false},
	} {
		err := validateIpv4CIDRBlock(ts.cidr)
		if !ts.valid && err == nil {
			t.Fatalf("Input '%s' should error but didn't!", ts.cidr)
		}
		if ts.valid && err != nil {
			t.Fatalf("Got unexpected error for '%s' input: %s", ts.cidr, err)
		}
	}
}

func Test_validateIpv6CIDRBlock(t *testing.T) {
	for _, ts := range []struct {
		cidr  string
		valid bool
	}{
		{"10.2.2.0/24", false},
		{"10.2.2.0/1234", false},
		{"::/0", true},
		{"::0/0", true},
		{"2000::/15", true},
		{"2001::/15", false},
		{"2001:db8::/122", true},
		{"", false},
	} {
		err := validateIpv6CIDRBlock(ts.cidr)
		if !ts.valid && err == nil {
			t.Fatalf("Input '%s' should error but didn't!", ts.cidr)
		}
		if ts.valid && err != nil {
			t.Fatalf("Got unexpected error for '%s' input: %s", ts.cidr, err)
		}
	}
}
