package kms_test

import (
	"regexp"
	"testing"

	tfkms "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/kms"
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
