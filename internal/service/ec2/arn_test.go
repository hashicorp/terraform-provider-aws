// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"regexp"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestInstanceProfileARNToName(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName      string
		InputARN      string
		ExpectedError *regexp.Regexp
		ExpectedName  string
	}{
		{
			TestName:      "empty ARN",
			InputARN:      "",
			ExpectedError: regexache.MustCompile(`parsing ARN`),
		},
		{
			TestName:      "unparsable ARN",
			InputARN:      "test",
			ExpectedError: regexache.MustCompile(`parsing ARN`),
		},
		{
			TestName:      "invalid ARN service",
			InputARN:      "arn:aws:ec2:us-east-1:123456789012:instance/i-12345678", //lintignore:AWSAT003,AWSAT005
			ExpectedError: regexache.MustCompile(`expected service iam`),
		},
		{
			TestName:      "invalid ARN resource parts",
			InputARN:      "arn:aws:iam:us-east-1:123456789012:name", //lintignore:AWSAT003,AWSAT005
			ExpectedError: regexache.MustCompile(`expected at least 2 resource parts`),
		},
		{
			TestName:      "invalid ARN resource prefix",
			InputARN:      "arn:aws:iam:us-east-1:123456789012:role/name", //lintignore:AWSAT003,AWSAT005
			ExpectedError: regexache.MustCompile(`expected resource prefix instance-profile`),
		},
		{
			TestName:     "valid ARN",
			InputARN:     "arn:aws:iam:us-east-1:123456789012:instance-profile/name", //lintignore:AWSAT003,AWSAT005
			ExpectedName: names.AttrName,
		},
		{
			TestName:     "valid ARN with multiple parts",
			InputARN:     "arn:aws:iam:us-east-1:123456789012:instance-profile/path/name", //lintignore:AWSAT003,AWSAT005
			ExpectedName: names.AttrName,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.TestName, func(t *testing.T) {
			t.Parallel()

			got, err := instanceProfileARNToName(testCase.InputARN)

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
