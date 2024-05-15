// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dynamodb_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfdynamodb "github.com/hashicorp/terraform-provider-aws/internal/service/dynamodb"
)

func TestARNForNewRegion(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName      string
		ARN           string
		NewRegion     string
		ExpectedARN   string
		ErrorExpected bool
	}{
		{
			TestName:      acctest.CtBasic,
			ARN:           "arn:aws:dynamodb:us-west-2:786648903940:table/tf-acc-test-7864711876941043153", //lintignore:AWSAT003,AWSAT005
			NewRegion:     "us-east-2",                                                                     //lintignore:AWSAT003
			ExpectedARN:   "arn:aws:dynamodb:us-east-2:786648903940:table/tf-acc-test-7864711876941043153", //lintignore:AWSAT003,AWSAT005
			ErrorExpected: false,
		},
		{
			TestName:      "basic2",
			ARN:           "arn:aws:dynamodb:us-west-2:786648903940:table/tf-acc-test-7864711876941043153", //lintignore:AWSAT003,AWSAT005
			NewRegion:     "us-east-1",                                                                     //lintignore:AWSAT003
			ExpectedARN:   "arn:aws:dynamodb:us-east-1:786648903940:table/tf-acc-test-7864711876941043153", //lintignore:AWSAT003,AWSAT005
			ErrorExpected: false,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.TestName, func(t *testing.T) {
			t.Parallel()

			got, err := tfdynamodb.ARNForNewRegion(testCase.ARN, testCase.NewRegion)

			if err != nil && !testCase.ErrorExpected {
				t.Errorf("did not expect an error but got one: %s", err)
			}

			if err == nil && testCase.ErrorExpected {
				t.Error("expected an error but got none")
			}

			if got != testCase.ExpectedARN {
				t.Errorf("got %s, expected %s", got, testCase.ExpectedARN)
			}
		})
	}
}
