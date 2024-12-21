// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
)

func TestNormalizeNameIntoRoute53Representation(t *testing.T) {
	t.Parallel()

	cases := []struct {
		UserInput, ExpectedR53Output string
	}{
		// Preserve escape code
		{"a\\000c.example.com", "a\\000c.example.com"},
		{"a\\056c.example.com", "a\\056c.example.com"}, // with escaped "."

		// Preserve "-" / "_" as-is
		{"a-b.example.com", "a-b.example.com"},
		{"_abc.example.com", "_abc.example.com"},

		// no conversion
		{"www.example.com", "www.example.com"},

		// converted into lower-case
		{"AbC.example.com", "abc.example.com"},

		// convert into escape code
		{"*.example.com", "\\052.example.com"},
		{"!.example.com", "\\041.example.com"},
		{"a/b.example.com", "a\\057b.example.com"},
		{"/.example.com", "\\057.example.com"},
		{"~.example.com", "\\176.example.com"},
	}

	for _, tc := range cases {
		actual := normalizeNameIntoRoute53APIRepresentation(tc.UserInput)

		if actual != tc.ExpectedR53Output {
			t.Errorf(
				"user input: %+q\nexpected: %+q\nr53 output: %+q",
				tc.UserInput, tc.ExpectedR53Output, actual,
			)
		}
	}
}

func TestNormalizeAliasName(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Input, Output string
	}{
		{"www.example.com", "www.example.com"},
		{"www.example.com.", "www.example.com"},
		{"dualstack.name-123456789.region.elb.amazonaws.com", "dualstack.name-123456789.region.elb.amazonaws.com"},
		{"dualstacktest.test", "dualstacktest.test"},
		{"ipv6.name-123456789.region.elb.amazonaws.com", "ipv6.name-123456789.region.elb.amazonaws.com"},
		{"NAME-123456789.region.elb.amazonaws.com", "name-123456789.region.elb.amazonaws.com"},
		{"name-123456789.region.elb.amazonaws.com", "name-123456789.region.elb.amazonaws.com"},
		{"\\052.example.com", "\\052.example.com"},
	}

	for _, tc := range cases {
		actual := normalizeAliasName(tc.Input)
		if actual != tc.Output {
			t.Fatalf("input: %s\noutput: %s", tc.Input, actual)
		}
	}
}

func TestCleanZoneID(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Input, Output string
	}{
		{"/hostedzone/foo", "foo"},
		{"/change/foo", "/change/foo"},
		{"/bar", "/bar"},
	}

	for _, tc := range cases {
		actual := cleanZoneID(tc.Input)
		if actual != tc.Output {
			t.Fatalf("input: %s\noutput: %s", tc.Input, actual)
		}
	}
}

func TestNormalizeZoneName(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Input  interface{}
		Output string
	}{
		{"example.com", "example.com"},
		{"example.com.", "example.com"},
		{"www.example.com.", "www.example.com"},
		{"www.ExAmPlE.COM.", "www.example.com"},
		{"", ""},
		{".", "."},
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
		actual := normalizeZoneName(tc.Input)
		if actual != tc.Output {
			t.Fatalf("input: %s\noutput: %s", tc.Input, actual)
		}
	}
}
