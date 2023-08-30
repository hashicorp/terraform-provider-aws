// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package namevaluesfilters_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/generate/namevaluesfilters"
)

func TestNameValuesFiltersEC2Tags(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		filters namevaluesfilters.NameValuesFilters
		want    map[string][]string
	}{
		{
			name:    "nil",
			filters: namevaluesfilters.EC2Tags(nil),
			want:    map[string][]string{},
		},
		{
			name:    "nil",
			filters: namevaluesfilters.EC2Tags(map[string]string{}),
			want:    map[string][]string{},
		},
		{
			name: "tags",
			filters: namevaluesfilters.EC2Tags(map[string]string{
				"Name":    acctest.ResourcePrefix,
				"Purpose": "testing",
			}),
			want: map[string][]string{
				"tag:Name":    {acctest.ResourcePrefix},
				"tag:Purpose": {"testing"},
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := testCase.filters.Map()

			testNameValuesFiltersVerifyMap(t, got, testCase.want)
		})
	}
}
