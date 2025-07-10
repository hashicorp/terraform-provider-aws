// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kafka_test

import (
	"testing"

	tfkafka "github.com/hashicorp/terraform-provider-aws/internal/service/kafka"
)

func TestSortEndpointsString(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName string
		Input    string
		Expected string
	}{
		{
			TestName: "empty",
			Input:    "",
			Expected: "",
		},
		{
			TestName: "one endpoint",
			Input:    "this:123",
			Expected: "this:123",
		},
		{
			TestName: "three endpoints",
			Input:    "this:123,is:147,just.a.test:443",
			Expected: "is:147,just.a.test:443,this:123",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.TestName, func(t *testing.T) {
			t.Parallel()

			got := tfkafka.SortEndpointsString(testCase.Input)

			if got != testCase.Expected {
				t.Errorf("got %s, expected %s", got, testCase.Expected)
			}
		})
	}
}
