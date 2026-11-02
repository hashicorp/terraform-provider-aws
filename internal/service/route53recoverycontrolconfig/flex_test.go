// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package route53recoverycontrolconfig

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/route53recoverycontrolconfig/types"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestFlattenClusterEndpoints(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    []awstypes.ClusterEndpoint
		expected []any
	}{
		"nil": {
			input:    nil,
			expected: nil,
		},
		"sorted by region": {
			input: []awstypes.ClusterEndpoint{
				{Endpoint: aws.String("https://c"), Region: aws.String(endpoints.UsWest2RegionID)},
				{Endpoint: aws.String("https://a"), Region: aws.String(endpoints.EuWest1RegionID)},
				{Endpoint: aws.String("https://e"), Region: aws.String(endpoints.UsEast1RegionID)},
				{Endpoint: aws.String("https://b"), Region: aws.String(endpoints.ApSoutheast2RegionID)},
				{Endpoint: aws.String("https://d"), Region: aws.String(endpoints.ApNortheast1RegionID)},
			},
			expected: []any{
				map[string]any{names.AttrEndpoint: "https://d", names.AttrRegion: endpoints.ApNortheast1RegionID},
				map[string]any{names.AttrEndpoint: "https://b", names.AttrRegion: endpoints.ApSoutheast2RegionID},
				map[string]any{names.AttrEndpoint: "https://a", names.AttrRegion: endpoints.EuWest1RegionID},
				map[string]any{names.AttrEndpoint: "https://e", names.AttrRegion: endpoints.UsEast1RegionID},
				map[string]any{names.AttrEndpoint: "https://c", names.AttrRegion: endpoints.UsWest2RegionID},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := flattenClusterEndpoints(tc.input)

			if diff := cmp.Diff(tc.expected, got); diff != "" {
				t.Errorf("unexpected diff (-want +got):\n%s", diff)
			}
		})
	}
}
