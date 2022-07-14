package ec2_test

import (
	"regexp"
	"testing"

	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
)

func TestInstanceProfileARNToName(t *testing.T) {
	testCases := []struct {
		TestName      string
		InputARN      string
		ExpectedError *regexp.Regexp
		ExpectedName  string
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
			InputARN:      "arn:aws:ec2:us-east-1:123456789012:instance/i-12345678", //lintignore:AWSAT003,AWSAT005
			ExpectedError: regexp.MustCompile(`expected service iam`),
		},
		{
			TestName:      "invalid ARN resource parts",
			InputARN:      "arn:aws:iam:us-east-1:123456789012:name", //lintignore:AWSAT003,AWSAT005
			ExpectedError: regexp.MustCompile(`expected at least 2 resource parts`),
		},
		{
			TestName:      "invalid ARN resource prefix",
			InputARN:      "arn:aws:iam:us-east-1:123456789012:role/name", //lintignore:AWSAT003,AWSAT005
			ExpectedError: regexp.MustCompile(`expected resource prefix instance-profile`),
		},
		{
			TestName:     "valid ARN",
			InputARN:     "arn:aws:iam:us-east-1:123456789012:instance-profile/name", //lintignore:AWSAT003,AWSAT005
			ExpectedName: "name",
		},
		{
			TestName:     "valid ARN with multiple parts",
			InputARN:     "arn:aws:iam:us-east-1:123456789012:instance-profile/path/name", //lintignore:AWSAT003,AWSAT005
			ExpectedName: "name",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.TestName, func(t *testing.T) {
			got, err := tfec2.InstanceProfileARNToName(testCase.InputARN)

			if err == nil && testCase.ExpectedError != nil {
				t.Fatalf("expected error %s, got no error", testCase.ExpectedError.String())
			}

			if err != nil && testCase.ExpectedError == nil {
				t.Fatalf("got unexpected error: %s", err)
			}

			if err != nil && !testCase.ExpectedError.MatchString(err.Error()) {
				t.Fatalf("expected error %s, got: %s", testCase.ExpectedError.String(), err)
			}

			if got != testCase.ExpectedName {
				t.Errorf("got %s, expected %s", got, testCase.ExpectedName)
			}
		})
	}
}
