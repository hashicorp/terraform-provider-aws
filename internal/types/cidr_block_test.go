// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package types

import "testing"

func TestValidateCIDRBlock(t *testing.T) {
	t.Parallel()

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
		err := ValidateCIDRBlock(ts.cidr)
		if !ts.valid && err == nil {
			t.Fatalf("Input '%s' should error but didn't!", ts.cidr)
		}
		if ts.valid && err != nil {
			t.Fatalf("Got unexpected error for '%s' input: %s", ts.cidr, err)
		}
	}
}

func TestCIDRBlocksEqual(t *testing.T) {
	t.Parallel()

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
		equal := CIDRBlocksEqual(ts.cidr1, ts.cidr2)
		if ts.equal != equal {
			t.Fatalf("CIDRBlocksEqual(%q, %q) should be: %t", ts.cidr1, ts.cidr2, ts.equal)
		}
	}
}

func TestCanonicalCIDRBlock(t *testing.T) {
	t.Parallel()

	for _, ts := range []struct {
		cidr     string
		expected string
	}{
		{"10.2.2.0/24", "10.2.2.0/24"},
		{"::/0", "::/0"},
		{"::0/0", "::/0"},
		{"", ""},
	} {
		got := CanonicalCIDRBlock(ts.cidr)
		if ts.expected != got {
			t.Fatalf("CanonicalCIDRBlock(%q) should be: %q, got: %q", ts.cidr, ts.expected, got)
		}
	}
}

func TestCIDRBlocksOverlap(t *testing.T) {
	t.Parallel()

	for _, ts := range []struct {
		cidr1    string
		cidr2    string
		overlaps bool
	}{
		{"10.0.0.0/16", "10.0.1.0/24", true},
		{"172.16.0.0/16", "172.16.0.0/16", true},
		{"192.168.0.0/24", "172.16.0.0/24", false},
		{"10.0.1.0/24", "10.0.0.0/24", false},
		{"10.0.0.0/16", "10.0.255.0/24", true},
		{"10.0.1.0/24", "10.0.0.0/16", false},
		{"2001:db8::/32", "2001:db8:1234::/48", true},
		{"2001:db8::/64", "2001:db8::/64", true},
		{"2001:db8:0000:0000::/64", "2001:db8::/64", true},
		{"2001:db8::/48", "2001:db9::/48", false},
		{"2001:db8::/32", "2a00:1450::/32", false},
		{"2001:db8::/48", "2001:db8:0:ffff::/64", true},
		{"", "10.0.0.0/24", false},
		{"10.0.0.0/24", "", false},
		{"", "", false},
		{"not-a-cidr", "10.0.0.0/24", false},
		{"10.0.0.0/24", "not-a-cidr", false},
		{"10.0.0.0/99", "10.0.0.0/24", false},
		{"0.0.0.0/0", "10.0.0.0/24", true},
		{"::/0", "2001:db8::/32", true},
		{"127.0.0.0/8", "127.0.0.1/32", true},
		{"::1/128", "::1/128", true},
	} {
		overlaps := CIDRBlocksOverlap(ts.cidr1, ts.cidr2)
		if ts.overlaps != overlaps {
			t.Fatalf("CIDRBlocksOverlap(%q, %q) should be: %t", ts.cidr1, ts.cidr2, ts.overlaps)
		}
	}
}
