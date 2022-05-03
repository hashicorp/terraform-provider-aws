package namevaluesfilters_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/generate/namevaluesfilters"
)

func TestNameValuesFiltersEc2Tags(t *testing.T) {
	testCases := []struct {
		name    string
		filters namevaluesfilters.NameValuesFilters
		want    map[string][]string
	}{
		{
			name:    "nil",
			filters: namevaluesfilters.Ec2Tags(nil),
			want:    map[string][]string{},
		},
		{
			name:    "nil",
			filters: namevaluesfilters.Ec2Tags(map[string]string{}),
			want:    map[string][]string{},
		},
		{
			name: "tags",
			filters: namevaluesfilters.Ec2Tags(map[string]string{
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
		t.Run(testCase.name, func(t *testing.T) {
			got := testCase.filters.Map()

			testNameValuesFiltersVerifyMap(t, got, testCase.want)
		})
	}
}
