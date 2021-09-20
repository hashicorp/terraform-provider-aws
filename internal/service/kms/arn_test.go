package kms

import (
	"regexp"
	"testing"

	tfkms "github.com/hashicorp/terraform-provider-aws/aws/internal/service/kms"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func TestAliasARNToKeyARN(t *testing.T) {
	testCases := []struct {
		TestName      string
		InputARN      string
		ExpectedError *regexp.Regexp
		ExpectedARN   string
	}{
		{
			TestName:      "empty ARN",
			InputARN:      "",
			ExpectedError: regexp.MustCompile(`error parsing ARN`),
		},
		{
			TestName:      "unparsable ARN",
			InputARN:      "test",
			ExpectedError: regexp.MustCompile(`error parsing ARN`),
		},
		{
			TestName:      "invalid ARN service",
			InputARN:      "arn:aws:ec2:us-west-2:123456789012:alias/test-alias",
			ExpectedError: regexp.MustCompile(`expected service kms`),
		},
		{
			TestName:    "valid ARN",
			InputARN:    "arn:aws:kms:us-west-2:123456789012:alias/test-alias",
			ExpectedARN: "arn:aws:kms:us-west-2:123456789012:key/test-key",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.TestName, func(t *testing.T) {
			got, err := tfkms.AliasARNToKeyARN(testCase.InputARN, "test-key")

			if err == nil && testCase.ExpectedError != nil {
				t.Fatalf("expected error %s, got no error", testCase.ExpectedError.String())
			}

			if err != nil && testCase.ExpectedError == nil {
				t.Fatalf("got unexpected error: %s", err)
			}

			if err != nil && !testCase.ExpectedError.MatchString(err.Error()) {
				t.Fatalf("expected error %s, got: %s", testCase.ExpectedError.String(), err)
			}

			if got != testCase.ExpectedARN {
				t.Errorf("got %s, expected %s", got, testCase.ExpectedARN)
			}
		})
	}
}

func TestKeyARNOrIDEqual(t *testing.T) {
	testCases := []struct {
		name   string
		first  string
		second string
		want   bool
	}{
		{
			name:   "empty",
			first:  "",
			second: "",
			want:   true,
		},
		{
			name:   "equal IDs",
			first:  "1234abcd-12ab-34cd-56ef-1234567890ab",
			second: "1234abcd-12ab-34cd-56ef-1234567890ab",
			want:   true,
		},
		{
			name:   "not equal IDs",
			first:  "1234abcd-12ab-34cd-56ef-1234567890ab",
			second: "1234abcd-12ab-34cd-56ef-1234567890ac",
		},
		{
			name:   "equal ARNs",
			first:  "arn:aws:kms:us-east-2:111122223333:key/1234abcd-12ab-34cd-56ef-1234567890ab",
			second: "arn:aws:kms:us-east-2:111122223333:key/1234abcd-12ab-34cd-56ef-1234567890ab",
			want:   true,
		},
		{
			name:   "not equal ARNs",
			first:  "arn:aws:kms:us-east-2:111122223333:key/1234abcd-12ab-34cd-56ef-1234567890ab",
			second: "arn:aws:kms:us-east-2:111122224444:key/1234abcd-12ab-34cd-56ef-1234567890ab",
		},
		{
			name:   "equal first ID, second ARN",
			first:  "1234abcd-12ab-34cd-56ef-1234567890ab",
			second: "arn:aws:kms:us-east-2:111122223333:key/1234abcd-12ab-34cd-56ef-1234567890ab",
			want:   true,
		},
		{
			name:   "equal first ARN, second ID",
			first:  "arn:aws:kms:us-east-2:111122223333:key/1234abcd-12ab-34cd-56ef-1234567890ab",
			second: "1234abcd-12ab-34cd-56ef-1234567890ab",
			want:   true,
		},
		{
			name:   "not equal first ID, second ARN",
			first:  "1234abcd-12ab-34cd-56ef-1234567890ab",
			second: "arn:aws:kms:us-east-2:111122223333:key/1234abcd-12ab-34cd-56ef-1234567890ac",
		},
		{
			name:   "not equal first ARN, second ID",
			first:  "arn:aws:kms:us-east-2:111122223333:key/1234abcd-12ab-34cd-56ef-1234567890ab",
			second: "1234abcd-12ab-34cd-56ef-1234567890ac",
		},
		{
			name:   "not equal first ID, second incorrect ARN",
			first:  "1234abcd-12ab-34cd-56ef-1234567890ab",
			second: "arn:aws:kms:us-east-2:111122223333:alias/1234abcd-12ab-34cd-56ef-1234567890ab",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			got := tfkms.KeyARNOrIDEqual(testCase.first, testCase.second)

			if got != testCase.want {
				t.Errorf("unexpected Equal: %t", got)
			}
		})
	}
}
