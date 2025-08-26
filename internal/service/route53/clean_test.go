// Copyright (c) HashiCorp, Inc.
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
		{"\\052.example.com", "\\052.example.com"},
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
		output := normalizeDomainName(tc.input)

		if got, want := output, tc.output; got != want {
			t.Errorf("normalizeDomainName(%q) = %v, want %v", tc.input, got, want)
		}
	}
}
