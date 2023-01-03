package kms

import (
	"strings"
	"testing"
)

func TestValidGrantName(t *testing.T) {
	validValues := []string{
		"123",
		"Abc",
		"grant_1",
		"grant:/-",
	}

	for _, s := range validValues {
		_, errors := validGrantName(s, "name")
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
		_, errors := validGrantName(s, "name")
		if len(errors) == 0 {
			t.Fatalf("%q should not be a valid AWS KMS Grant Name", s)
		}
	}
}

func TestValidNameForDataSource(t *testing.T) {
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
		_, errors := validNameForDataSource(tc.Value, "name")
		if len(errors) != tc.ErrCount {
			t.Fatalf("AWS KMS Alias Name validation failed: %v", errors)
		}
	}
}

func TestValidNameForResource(t *testing.T) {
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
		_, errors := validNameForResource(tc.Value, "name")
		if len(errors) != tc.ErrCount {
			t.Fatalf("AWS KMS Alias Name validation failed: %v", errors)
		}
	}
}

func TestValidKey(t *testing.T) {
	cases := []struct {
		Value    string
		ErrCount int
	}{
		{
			Value:    "arbitrary-uuid-1234",
			ErrCount: 0,
		},
		{
			Value:    "arn:aws:kms:us-west-2:111122223333:key/arbitrary-uuid-1234", //lintignore:AWSAT003,AWSAT005
			ErrCount: 0,
		},
		{
			Value:    "alias/arbitrary-key",
			ErrCount: 0,
		},
		{
			Value:    "alias/arbitrary/key",
			ErrCount: 0,
		},
		{
			Value:    "arn:aws:kms:us-west-2:111122223333:alias/arbitrary-key", //lintignore:AWSAT003,AWSAT005
			ErrCount: 0,
		},
		{
			Value:    "arn:aws:kms:us-west-2:111122223333:alias/arbitrary/key", //lintignore:AWSAT003,AWSAT005
			ErrCount: 0,
		},
		{
			Value:    "$%wrongkey",
			ErrCount: 1,
		},
		{
			Value:    "arn:aws:lamda:foo:bar:key/xyz", //lintignore:AWSAT003,AWSAT005
			ErrCount: 1,
		},
	}

	for _, tc := range cases {
		_, errors := validKey(tc.Value, "key_id")
		if len(errors) != tc.ErrCount {
			t.Fatalf("%q validation failed: %v", tc.Value, errors)
		}
	}
}
