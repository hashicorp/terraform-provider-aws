// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
)

func TestCleanRecordName(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Input, Output string
	}{
		{"www.example.com", "www.example.com"},
		{"\\052.example.com", "*.example.com"},
		{"\\100.example.com", "@.example.com"},
		{"\\043.example.com", "#.example.com"},
		{"example.com", "example.com"},
	}

	for _, tc := range cases {
		actual := cleanRecordName(tc.Input)
		if actual != tc.Output {
			t.Fatalf("input: %s\noutput: %s", tc.Input, actual)
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
		{"\\052.example.com", "*.example.com"},
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
