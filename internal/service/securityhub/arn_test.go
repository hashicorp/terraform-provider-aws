// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securityhub_test

import (
	"regexp"
	"testing"

	tfsecurityhub "github.com/hashicorp/terraform-provider-aws/internal/service/securityhub"
)

func TestStandardsControlARNToStandardsSubscriptionARN(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName      string
		InputARN      string
		ExpectedError *regexp.Regexp
		ExpectedARN   string
	}{
		{
			TestName:      "empty ARN",
			InputARN:      "",
			ExpectedError: regexp.MustCompile(`parsing ARN`),
		},
		{
			TestName:      "unparsable ARN",
			InputARN:      "test",
			ExpectedError: regexp.MustCompile(`parsing ARN`),
		},
		{
			TestName:      "invalid ARN service",
			InputARN:      "arn:aws:ec2:us-west-2:1234567890:control/cis-aws-foundations-benchmark/v/1.2.0/1.1", //lintignore:AWSAT003,AWSAT005
			ExpectedError: regexp.MustCompile(`expected service securityhub`),
		},
		{
			TestName:      "invalid ARN resource parts",
			InputARN:      "arn:aws:securityhub:us-west-2:1234567890:control/cis-aws-foundations-benchmark", //lintignore:AWSAT003,AWSAT005
			ExpectedError: regexp.MustCompile(`expected at least 3 resource parts`),
		},
		{
			TestName:    "valid ARN",
			InputARN:    "arn:aws:securityhub:us-west-2:1234567890:control/cis-aws-foundations-benchmark/v/1.2.0/1.1",  //lintignore:AWSAT003,AWSAT005
			ExpectedARN: "arn:aws:securityhub:us-west-2:1234567890:subscription/cis-aws-foundations-benchmark/v/1.2.0", //lintignore:AWSAT003,AWSAT005
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.TestName, func(t *testing.T) {
			t.Parallel()

			got, err := tfsecurityhub.StandardsControlARNToStandardsSubscriptionARN(testCase.InputARN)

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
