// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package route53recoverycontrolconfig

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/route53recoverycontrolconfig/types"
	"github.com/google/go-cmp/cmp"
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
				{Endpoint: aws.String("https://c.us-west-2"), Region: aws.String("us-west-2")},
				{Endpoint: aws.String("https://a.eu-west-1"), Region: aws.String("eu-west-1")},
				{Endpoint: aws.String("https://e.us-east-1"), Region: aws.String("us-east-1")},
				{Endpoint: aws.String("https://b.ap-southeast-2"), Region: aws.String("ap-southeast-2")},
				{Endpoint: aws.String("https://d.ap-northeast-1"), Region: aws.String("ap-northeast-1")},
			},
			expected: []any{
				map[string]any{names.AttrEndpoint: "https://d.ap-northeast-1", names.AttrRegion: "ap-northeast-1"},
				map[string]any{names.AttrEndpoint: "https://b.ap-southeast-2", names.AttrRegion: "ap-southeast-2"},
				map[string]any{names.AttrEndpoint: "https://a.eu-west-1", names.AttrRegion: "eu-west-1"},
				map[string]any{names.AttrEndpoint: "https://e.us-east-1", names.AttrRegion: "us-east-1"},
				map[string]any{names.AttrEndpoint: "https://c.us-west-2", names.AttrRegion: "us-west-2"},
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
