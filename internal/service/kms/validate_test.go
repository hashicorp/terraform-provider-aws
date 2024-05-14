// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kms

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestValidGrantName(t *testing.T) {
	t.Parallel()

	validValues := []string{
		"123",
		"Abc",
		"grant_1",
		"grant:/-",
	}

	for _, s := range validValues {
		_, errors := validGrantName(s, names.AttrName)
		if len(errors) > 0 {
			t.Fatalf("%q AWS KMS Grant Name should have been valid: %v", s, errors)
		}
	}

	invalidValues := []string{
		strings.Repeat("w", 257),
		"grant.invalid",
		";",
		"white space",
	}

	for _, s := range invalidValues {
		_, errors := validGrantName(s, names.AttrName)
		if len(errors) == 0 {
			t.Fatalf("%q should not be a valid AWS KMS Grant Name", s)
		}
	}
}

func TestValidNameForDataSource(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Value    string
		ErrCount int
	}{
		{
			Value:    "alias/aws/s3",
			ErrCount: 0,
		},
		{
			Value:    "alias/aws-service-test",
			ErrCount: 0,
		},
		{
			Value:    "alias/hashicorp",
			ErrCount: 0,
		},
		{
			Value:    "alias/Service:Test",
			ErrCount: 1,
		},
		{
			Value:    "hashicorp",
			ErrCount: 1,
		},
		{
			Value:    "hashicorp/terraform",
			ErrCount: 1,
		},
	}

	for _, tc := range cases {
		_, errors := validNameForDataSource(tc.Value, names.AttrName)
		if len(errors) != tc.ErrCount {
			t.Fatalf("AWS KMS Alias Name validation failed: %v", errors)
		}
	}
}

func TestValidNameForResource(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Value    string
		ErrCount int
	}{
		{
			Value:    "alias/hashicorp",
			ErrCount: 0,
		},
		{
			Value:    "alias/aws-service-test",
			ErrCount: 0,
		},
		{
			Value:    "alias/aws/s3",
			ErrCount: 1,
		},
		{
			Value:    "alias/Service:Test",
			ErrCount: 1,
		},
		{
			Value:    "hashicorp",
			ErrCount: 1,
		},
		{
			Value:    "hashicorp/terraform",
			ErrCount: 1,
		},
	}

	for _, tc := range cases {
		_, errors := validNameForResource(tc.Value, names.AttrName)
		if len(errors) != tc.ErrCount {
			t.Fatalf("AWS KMS Alias Name validation failed: %v", errors)
		}
	}
}

func TestValidateKeyOrAlias(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Value    string
		ErrCount int
		valid    bool
	}{
		{
			Value:    "57ff7a43-341d-46b6-aee3-a450c9de6dc8",
			ErrCount: 0,
			valid:    true,
		},
		{
			Value:    "arn:aws:kms:us-west-2:111122223333:key/57ff7a43-341d-46b6-aee3-a450c9de6dc8", //lintignore:AWSAT003,AWSAT005
			ErrCount: 0,
			valid:    true,
		},
		{
			Value:    "alias/arbitrary-key",
			ErrCount: 0,
			valid:    true,
		},
		{
			Value:    "alias/arbitrary/key",
			ErrCount: 0,
			valid:    true,
		},
		{
			Value:    "mrk-f827515944fb43f9b902a09d2c8b554f",
			ErrCount: 0,
			valid:    true,
		},
		{
			Value:    "arn:aws:kms:us-west-2:111122223333:key/mrk-a835af0b39c94b86a21a8fc9535df681", //lintignore:AWSAT003,AWSAT005
			ErrCount: 0,
			valid:    true,
		},
		{
			Value:    "arn:aws:kms:us-west-2:111122223333:alias/arbitrary-key", //lintignore:AWSAT003,AWSAT005
			ErrCount: 0,
			valid:    true,
		},
		{
			Value:    "arn:aws:kms:us-west-2:111122223333:alias/arbitrary/key", //lintignore:AWSAT003,AWSAT005
			ErrCount: 0,
			valid:    true,
		},
		{
			Value:    "$%wrongkey",
			ErrCount: 1,
			valid:    false,
		},
		{
			Value:    "arn:aws:lamda:foo:bar:key/xyz", //lintignore:AWSAT003,AWSAT005
			ErrCount: 1,
			valid:    false,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.Value, func(t *testing.T) {
			t.Parallel()

			_, errors := ValidateKeyOrAlias(tc.Value, names.AttrKeyID)
			if (len(errors) == 0) != tc.valid {
				t.Errorf("%q ValidateKMSKeyOrAlias failed: %v", tc.Value, errors)
			}
		})
	}
}

func TestValidateKeyARN(t *testing.T) {
	t.Parallel()

	testcases := map[string]struct {
		in    any
		valid bool
	}{
		"kms key id": {
			in:    "arn:aws:kms:us-west-2:123456789012:key/57ff7a43-341d-46b6-aee3-a450c9de6dc8", // lintignore:AWSAT003,AWSAT005
			valid: true,
		},
		"kms mrk key id": {
			in:    "arn:aws:kms:us-west-2:111122223333:key/mrk-a835af0b39c94b86a21a8fc9535df681", // lintignore:AWSAT003,AWSAT005
			valid: true,
		},
		"kms non-key id": {
			in:    "arn:aws:kms:us-west-2:123456789012:something/else", // lintignore:AWSAT003,AWSAT005
			valid: false,
		},
		"non-kms arn": {
			in:    "arn:aws:iam::123456789012:user/David", // lintignore:AWSAT005
			valid: false,
		},
		"not an arn": {
			in:    "not an arn",
			valid: false,
		},
		"not a string": {
			in:    123,
			valid: false,
		},
	}

	for name, testcase := range testcases {
		testcase := testcase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			aWs, aEs := validateKeyARN(testcase.in, names.AttrField)
			if len(aWs) != 0 {
				t.Errorf("expected no warnings, got %v", aWs)
			}
			if testcase.valid {
				if len(aEs) != 0 {
					t.Errorf("expected no errors, got %v", aEs)
				}
			} else {
				if len(aEs) == 0 {
					t.Error("expected errors, got none")
				}
			}
		})
	}
}
