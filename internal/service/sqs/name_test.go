// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sqs_test

import (
	"testing"

	tfsqs "github.com/hashicorp/terraform-provider-aws/internal/service/sqs"
)

func TestQueueNameFromURL(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		Name              string
		URL               string
		ExpectedQueueName string
		ExpectError       bool
	}{
		{
			Name:        "empty URL",
			ExpectError: true,
		},
		{
			Name:        "invalid URL",
			URL:         "---",
			ExpectError: true,
		},
		{
			Name:        "too few path parts",
			URL:         "http://sqs.us-west-2.amazonaws.com", //lintignore:AWSAT003
			ExpectError: true,
		},
		{
			Name:        "too many path parts",
			URL:         "http://sqs.us-west-2.amazonaws.com/123456789012/queueName/extra", //lintignore:AWSAT003
			ExpectError: true,
		},
		{
			Name:              "valid URL",
			URL:               "http://sqs.us-west-2.amazonaws.com/123456789012/queueName", //lintignore:AWSAT003
			ExpectedQueueName: "queueName",
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()

			got, err := tfsqs.QueueNameFromURL(testCase.URL)

			if err != nil && !testCase.ExpectError {
				t.Errorf("got unexpected error: %s", err)
			}

			if err == nil && testCase.ExpectError {
				t.Errorf("expected error, but received none")
			}

			if got != testCase.ExpectedQueueName {
				t.Errorf("got %s, expected %s", got, testCase.ExpectedQueueName)
			}
		})
	}
}
