// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package observabilityadmin

import (
	"slices"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/observabilityadmin/types"
)

func TestNormalizeTelemetryRuleRegionSelection(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		rule        *awstypes.TelemetryRule
		wantRegions []string
	}{
		"nil rule": {},
		"all regions": {
			rule: &awstypes.TelemetryRule{
				AllRegions: aws.Bool(true),
				Regions:    []string{"region-1", "region-2"},
			},
		},
		"selected regions": {
			rule: &awstypes.TelemetryRule{
				Regions: []string{"region-1"},
			},
			wantRegions: []string{"region-1"},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			normalizeTelemetryRuleRegionSelection(testCase.rule)

			if testCase.rule == nil {
				return
			}
			if got, want := testCase.rule.Regions, testCase.wantRegions; !slices.Equal(got, want) {
				t.Errorf("Regions = %v, want %v", got, want)
			}
		})
	}
}
