// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package route53

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
)

func TestNormalizeAliasDomainName(t *testing.T) {
	t.Parallel()

	cases := []struct {
		input  any
		output string
	}{
		{"www.example.com", "www.example.com"},
		{"www.example.com.", "www.example.com"},
		{"dualstack.name-123456789.region.elb.amazonaws.com", "dualstack.name-123456789.region.elb.amazonaws.com"},
		{aws.String("dualstacktest.test"), "dualstacktest.test"},
		{"ipv6.name-123456789.region.elb.amazonaws.com", "ipv6.name-123456789.region.elb.amazonaws.com"},
		{"NAME-123456789.region.elb.amazonaws.com", "name-123456789.region.elb.amazonaws.com"},
		{"name-123456789.region.elb.amazonaws.com", "name-123456789.region.elb.amazonaws.com"},
		{"*.example.com", "\\052.example.com"},     // "*" is normalized to octal "\052"
		{"@.example.com", "\\100.example.com"},     // "@" is normalized to octal "\100"
		{"\\052.example.com", "\\052.example.com"}, // octal "\052" is preserved
		{42, ""},
	}

	for _, tc := range cases {
		output := normalizeAliasDomainName(tc.input)

		if got, want := output, tc.output; got != want {
			t.Errorf("normalizeAliasDomainName(%q) = %v, want %v", tc.input, got, want)
		}
	}
}

func TestCleanZoneID(t *testing.T) {
	t.Parallel()

	cases := []struct {
		input, output string
	}{
		{"/hostedzone/foo", "foo"},
		{"/change/foo", "/change/foo"},
		{"/bar", "/bar"},
	}

	for _, tc := range cases {
		output := cleanZoneID(tc.input)

		if got, want := output, tc.output; got != want {
			t.Errorf("cleanZoneID(%q) = %v, want %v", tc.input, got, want)
		}
	}
}

func TestNormalizeDomainName(t *testing.T) {
	t.Parallel()

	cases := []struct {
		input  any
		output string
	}{
		{"example.com", "example.com"},             // as-is
		{"example.com.", "example.com"},            // trailing dot is removed
		{"www.example.com.", "www.example.com"},    // trailing dot is removed
		{"www.ExAmPlE.COM.", "www.example.com"},    // case is normalized to lower case
		{"*.example.com", "\\052.example.com"},     // "*" is normalized to octal "\052"
		{"@.example.com", "\\100.example.com"},     // "@" is normalized to octal "\100"
		{"\\052.example.com", "\\052.example.com"}, // octal "\052" is preserved
		{"", ""},   // as-is
		{".", "."}, // dot-only is preserved
		{aws.String("example.com"), "example.com"},
		{aws.String("example.com."), "example.com"},
		{aws.String("www.example.com."), "www.example.com"},
		{aws.String("www.ExAmPlE.COM."), "www.example.com"},
		{aws.String(""), ""},
		{aws.String("."), "."},
		{(*string)(nil), ""},
		{42, ""},
		{nil, ""},
	}

	for _, tc := range cases {
		output := normalizeDomainName(tc.input)

		if got, want := output, tc.output; got != want {
			t.Errorf("normalizeDomainName(%q) = %v, want %v", tc.input, got, want)
		}
	}
}

func TestDenormalizeDomainName(t *testing.T) {
	t.Parallel()

	cases := []struct {
		input  any
		output string
	}{
		{"example.com", "example.com"},             // as-is
		{"example.com.", "example.com"},            // trailing dot is removed
		{"www.example.com.", "www.example.com"},    // trailing dot is removed
		{"\\052.example.com", "*.example.com"},     // octal "\052" ("*") is denormalized
		{"\\100.example.com", "\\100.example.com"}, // octal "\100" ("@") is preserved
		{"", ""},
		{".", "."},
		{aws.String("example.com"), "example.com"},
		{aws.String("example.com."), "example.com"},
		{aws.String("www.example.com."), "www.example.com"},
		{aws.String(""), ""},
		{aws.String("."), "."},
		{(*string)(nil), ""},
		{42, ""},
		{nil, ""},
	}

	for _, tc := range cases {
		output := denormalizeDomainName(tc.input)

		if got, want := output, tc.output; got != want {
			t.Errorf("denormalizeDomainName(%q) = %v, want %v", tc.input, got, want)
		}
	}
}
