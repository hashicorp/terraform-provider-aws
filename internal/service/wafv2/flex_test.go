// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package wafv2

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/wafv2/types"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func Test_expandWebACLRulesJSON(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		rawRules string
		want     []awstypes.Rule
		wantErr  bool
	}{
		"empty string": {
			rawRules: "",
			wantErr:  true,
		},
		"empty array": {
			rawRules: "[]",
			want:     []awstypes.Rule{},
		},
		"single empty object": {
			rawRules: "[{}]",
			wantErr:  true,
		},
		"single null object": {
			rawRules: "[null]",
			wantErr:  true,
		},
		"valid object": {
			rawRules: `[{"Action":{"Count":{}},"Name":"rule-1","Priority":1,"Statement":{"RateBasedStatement":{"AggregateKeyType":"IP","EvaluationWindowSec":600,"Limit":10000,"ScopeDownStatement":{"GeoMatchStatement":{"CountryCodes":["US","NL"]}}}},"VisibilityConfig":{"CloudwatchMetricsEnabled":false,"MetricName":"friendly-rule-metric-name","SampledRequestsEnabled":false}}]`,
			want: []awstypes.Rule{
				{
					Name:     aws.String("rule-1"),
					Priority: 1,
					Action: &awstypes.RuleAction{
						Count: &awstypes.CountAction{},
					},
					Statement: &awstypes.Statement{
						RateBasedStatement: &awstypes.RateBasedStatement{
							Limit:               aws.Int64(10000),
							AggregateKeyType:    awstypes.RateBasedStatementAggregateKeyType("IP"),
							EvaluationWindowSec: 600,
							ScopeDownStatement: &awstypes.Statement{
								GeoMatchStatement: &awstypes.GeoMatchStatement{
									CountryCodes: []awstypes.CountryCode{"US", "NL"},
								},
							},
						},
					},
					VisibilityConfig: &awstypes.VisibilityConfig{
						CloudWatchMetricsEnabled: false,
						MetricName:               aws.String("friendly-rule-metric-name"),
						SampledRequestsEnabled:   false,
					},
				},
			},
		},
		"valid and empty object": {
			rawRules: `[{"Action":{"Count":{}},"Name":"rule-1","Priority":1,"Statement":{"RateBasedStatement":{"AggregateKeyType":"IP","EvaluationWindowSec":600,"Limit":10000,"ScopeDownStatement":{"GeoMatchStatement":{"CountryCodes":["US","NL"]}}}},"VisibilityConfig":{"CloudwatchMetricsEnabled":false,"MetricName":"friendly-rule-metric-name","SampledRequestsEnabled":false}},{}]`,
			wantErr:  true,
		},
	}

	ignoreExportedOpts := cmpopts.IgnoreUnexported(
		awstypes.Rule{},
		awstypes.RuleAction{},
		awstypes.CountAction{},
		awstypes.Statement{},
		awstypes.RateBasedStatement{},
		awstypes.GeoMatchStatement{},
		awstypes.VisibilityConfig{},
	)

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := expandWebACLRulesJSON(tc.rawRules)
			if (err != nil) != tc.wantErr {
				t.Errorf("expandWebACLRulesJSON() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if diff := cmp.Diff(got, tc.want, ignoreExportedOpts); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}
