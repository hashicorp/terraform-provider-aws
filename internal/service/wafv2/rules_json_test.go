// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package wafv2

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/wafv2/types"
)

// apiRules decodes `Rules` as returned on the wire by GetWebACL / GetRuleGroup.
// encoding/json decodes the base64 SearchString into []byte, as the SDK does.
func apiRules(t *testing.T, s string) []awstypes.Rule {
	t.Helper()

	var r []awstypes.Rule
	if err := json.Unmarshal([]byte(s), &r); err != nil {
		t.Fatal(err)
	}
	return r
}

// GetWebACL output: Bot Control with scope-down statements.
const testRulesJSONAPIBotControl = `[{"Name":"BotControl-Targeted","Priority":0,"Statement":{"ManagedRuleGroupStatement":{"VendorName":"AWS","Name":"AWSManagedRulesBotControlRuleSet","Version":"Version_5.0","ScopeDownStatement":{"ByteMatchStatement":{"SearchString":"ZXhhbXBsZS5jb20=","FieldToMatch":{"SingleHeader":{"Name":"host"}},"TextTransformations":[{"Priority":0,"Type":"LOWERCASE"}],"PositionalConstraint":"EXACTLY"}},"ManagedRuleGroupConfigs":[{"AWSManagedRulesBotControlRuleSet":{"InspectionLevel":"TARGETED","EnableMachineLearning":true}}]}},"OverrideAction":{"None":{}},"VisibilityConfig":{"SampledRequestsEnabled":true,"CloudWatchMetricsEnabled":true,"MetricName":"BotControl-Targeted"}},{"Name":"BotControl-Common","Priority":1,"Statement":{"ManagedRuleGroupStatement":{"VendorName":"AWS","Name":"AWSManagedRulesBotControlRuleSet","Version":"Version_5.0","ScopeDownStatement":{"NotStatement":{"Statement":{"ByteMatchStatement":{"SearchString":"ZXhhbXBsZS5jb20=","FieldToMatch":{"SingleHeader":{"Name":"host"}},"TextTransformations":[{"Priority":0,"Type":"LOWERCASE"}],"PositionalConstraint":"EXACTLY"}}}},"ManagedRuleGroupConfigs":[{"AWSManagedRulesBotControlRuleSet":{"InspectionLevel":"COMMON"}}]}},"OverrideAction":{"None":{}},"VisibilityConfig":{"SampledRequestsEnabled":true,"CloudWatchMetricsEnabled":true,"MetricName":"BotControl-Common"}},{"Name":"block-canada","Priority":2,"Statement":{"NotStatement":{"Statement":{"GeoMatchStatement":{"CountryCodes":["CA"]}}}},"Action":{"Block":{}},"VisibilityConfig":{"SampledRequestsEnabled":true,"CloudWatchMetricsEnabled":true,"MetricName":"block-canada"}}]`

// The same rules as written with jsonencode(): out of Priority order, plain-text
// SearchString, header name in mixed case, EnableMachineLearning omitted, the
// "CloudwatchMetricsEnabled" key case used by the existing acceptance tests, and
// an empty RuleLabels.
const testRulesJSONConfigBotControl = `[
  {"Name":"block-canada","Priority":2,"Action":{"Block":{}},
   "Statement":{"NotStatement":{"Statement":{"GeoMatchStatement":{"CountryCodes":["CA"]}}}},
   "RuleLabels":[],
   "VisibilityConfig":{"CloudwatchMetricsEnabled":true,"MetricName":"block-canada","SampledRequestsEnabled":true}},
  {"Name":"BotControl-Targeted","Priority":0,"OverrideAction":{"None":{}},
   "Statement":{"ManagedRuleGroupStatement":{"Name":"AWSManagedRulesBotControlRuleSet","VendorName":"AWS","Version":"Version_5.0",
     "ManagedRuleGroupConfigs":[{"AWSManagedRulesBotControlRuleSet":{"InspectionLevel":"TARGETED"}}],
     "ScopeDownStatement":{"ByteMatchStatement":{"SearchString":"example.com","PositionalConstraint":"EXACTLY",
       "FieldToMatch":{"SingleHeader":{"Name":"Host"}},"TextTransformations":[{"Priority":0,"Type":"LOWERCASE"}]}}}},
   "VisibilityConfig":{"CloudWatchMetricsEnabled":true,"MetricName":"BotControl-Targeted","SampledRequestsEnabled":true}},
  {"Name":"BotControl-Common","Priority":1,"OverrideAction":{"None":{}},
   "Statement":{"ManagedRuleGroupStatement":{"Name":"AWSManagedRulesBotControlRuleSet","VendorName":"AWS","Version":"Version_5.0",
     "ManagedRuleGroupConfigs":[{"AWSManagedRulesBotControlRuleSet":{"InspectionLevel":"COMMON"}}],
     "ScopeDownStatement":{"NotStatement":{"Statement":{"ByteMatchStatement":{"SearchString":"example.com","PositionalConstraint":"EXACTLY",
       "FieldToMatch":{"SingleHeader":{"Name":"host"}},"TextTransformations":[{"Priority":0,"Type":"LOWERCASE"}]}}}}}},
   "VisibilityConfig":{"CloudWatchMetricsEnabled":true,"MetricName":"BotControl-Common","SampledRequestsEnabled":true}}
]`

// GetWebACL output: custom response, and a rate-based rule with the defaults AWS fills in.
const testRulesJSONAPIRateBased = `[{"Name":"rule-1","Priority":0,"Statement":{"ByteMatchStatement":{"SearchString":"YmFy","FieldToMatch":{"SingleQueryArgument":{"Name":"foo"}},"TextTransformations":[{"Priority":0,"Type":"LOWERCASE"}],"PositionalConstraint":"EXACTLY"}},"Action":{"Block":{"CustomResponse":{"ResponseCode":418,"ResponseHeaders":[{"Name":"is-valid","Value":"very-valid"}]}}},"VisibilityConfig":{"SampledRequestsEnabled":true,"CloudWatchMetricsEnabled":true,"MetricName":"rule-1"}},{"Name":"rule-2","Priority":1,"Statement":{"RateBasedStatement":{"Limit":4000,"EvaluationWindowSec":300,"AggregateKeyType":"IP"}},"Action":{"Count":{}},"VisibilityConfig":{"SampledRequestsEnabled":true,"CloudWatchMetricsEnabled":true,"MetricName":"rule-2"}}]`

const testRulesJSONConfigRateBased = `[
  {"Name":"rule-1","Priority":0,
   "Statement":{"ByteMatchStatement":{"SearchString":"bar","FieldToMatch":{"SingleQueryArgument":{"Name":"FOO"}},"TextTransformations":[{"Priority":0,"Type":"LOWERCASE"}],"PositionalConstraint":"EXACTLY"}},
   "Action":{"Block":{"CustomResponse":{"ResponseCode":418,"ResponseHeaders":[{"Name":"is-valid","Value":"very-valid"}]}}},
   "VisibilityConfig":{"SampledRequestsEnabled":true,"CloudWatchMetricsEnabled":true,"MetricName":"rule-1"}},
  {"Name":"rule-2","Priority":1,
   "Statement":{"RateBasedStatement":{"Limit":4000}},
   "Action":{"Count":{}},
   "VisibilityConfig":{"SampledRequestsEnabled":true,"CloudWatchMetricsEnabled":true,"MetricName":"rule-2"}}
]`

// GetWebACL output: Monetize action.
const testRulesJSONAPIMonetize = `[{"Name":"bot-control","Priority":0,"Statement":{"ManagedRuleGroupStatement":{"VendorName":"AWS","Name":"AWSManagedRulesBotControlRuleSet","Version":"Version_6.0","ManagedRuleGroupConfigs":[{"AWSManagedRulesBotControlRuleSet":{"InspectionLevel":"COMMON"}}]}},"OverrideAction":{"Count":{}},"VisibilityConfig":{"SampledRequestsEnabled":true,"CloudWatchMetricsEnabled":true,"MetricName":"bot-control"}},{"Name":"monetize","Priority":1,"Statement":{"ByteMatchStatement":{"SearchString":"L3dlYXRoZXI=","FieldToMatch":{"UriPath":{}},"TextTransformations":[{"Priority":0,"Type":"NONE"}],"PositionalConstraint":"STARTS_WITH"}},"Action":{"Monetize":{"PriceMultiplier":"1"}},"VisibilityConfig":{"SampledRequestsEnabled":true,"CloudWatchMetricsEnabled":true,"MetricName":"monetize"}}]`

func TestWebACLRulesJSONToSet_equivalent(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		config, api string
	}{
		"bot control scope-down": {testRulesJSONConfigBotControl, testRulesJSONAPIBotControl},
		"rate-based defaults":    {testRulesJSONConfigRateBased, testRulesJSONAPIRateBased},
		"monetize round trip":    {"", testRulesJSONAPIMonetize},
		"uri fragment fallback": {
			`[{"Name":"rule","Priority":0,"Action":{"Block":{}},"Statement":{"ByteMatchStatement":{"SearchString":"a","PositionalConstraint":"CONTAINS","FieldToMatch":{"UriFragment":{}},"TextTransformations":[{"Priority":0,"Type":"NONE"}]}},"VisibilityConfig":{"CloudWatchMetricsEnabled":false,"MetricName":"rule","SampledRequestsEnabled":false}}]`,
			`[{"Name":"rule","Priority":0,"Statement":{"ByteMatchStatement":{"SearchString":"YQ==","FieldToMatch":{"UriFragment":{"FallbackBehavior":"MATCH"}},"TextTransformations":[{"Priority":0,"Type":"NONE"}],"PositionalConstraint":"CONTAINS"}},"Action":{"Block":{}},"VisibilityConfig":{"SampledRequestsEnabled":false,"CloudWatchMetricsEnabled":false,"MetricName":"rule"}}]`,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			remote := apiRules(t, tc.api)
			config := tc.config
			if config == "" {
				// The canonical form of the API rules must be accepted as input.
				var err error
				if config, err = flattenRulesJSON(remote); err != nil {
					t.Fatal(err)
				}
			}
			if wire, _ := json.Marshal(remote); string(wire) == config {
				t.Fatal("test case does not exercise normalization")
			}

			got, err := webACLRulesJSONToSet(config, remote)
			if err != nil {
				t.Fatal(err)
			}
			if got != config {
				c, _ := canonicalRulesJSON(expandWebACLRulesJSON, config)
				r, _ := flattenRulesJSON(remote)
				t.Fatalf("not equivalent:\nconfig: %s\nremote: %s", c, r)
			}
		})
	}
}

func TestRuleGroupRulesJSONToSet_equivalent(t *testing.T) {
	t.Parallel()

	remote := apiRules(t, testRulesJSONAPIRateBased)
	got, err := ruleGroupRulesJSONToSet(testRulesJSONConfigRateBased, remote)
	if err != nil {
		t.Fatal(err)
	}
	if got != testRulesJSONConfigRateBased {
		t.Fatalf("not equivalent: %s", got)
	}
}

func TestWebACLRulesJSONToSet_drift(t *testing.T) {
	t.Parallel()

	remote := apiRules(t, testRulesJSONAPIRateBased)
	remote[1].Statement.RateBasedStatement.Limit = aws.Int64(2000) // Changed outside Terraform.

	got, err := webACLRulesJSONToSet(testRulesJSONConfigRateBased, remote)
	if err != nil {
		t.Fatal(err)
	}
	if got == testRulesJSONConfigRateBased {
		t.Fatal("drift not detected")
	}
	want, err := canonicalRulesJSON(expandWebACLRulesJSON, testRulesJSONConfigRateBased)
	if err != nil {
		t.Fatal(err)
	}
	// The canonical forms differ only in the changed value.
	if !strings.Contains(got, `"Limit":2000`) || strings.Replace(got, `"Limit":2000`, `"Limit":4000`, 1) != want {
		t.Fatalf("unexpected value:\ngot:  %s\nwant: %s", got, want)
	}
}

func TestRuleGroupRulesJSONToSet_drift(t *testing.T) {
	t.Parallel()

	remote := apiRules(t, testRulesJSONAPIRateBased)
	remote = remote[:1] // Rule deleted outside Terraform.

	got, err := ruleGroupRulesJSONToSet(testRulesJSONConfigRateBased, remote)
	if err != nil {
		t.Fatal(err)
	}
	if got == testRulesJSONConfigRateBased || strings.Contains(got, "rule-2") {
		t.Fatalf("drift not detected: %s", got)
	}
}

func TestWebACLRulesJSONToSet_shieldRule(t *testing.T) {
	t.Parallel()

	remote := apiRules(t, testRulesJSONAPIRateBased)
	remote = append(remote, awstypes.Rule{
		Name:     aws.String("ShieldMitigationRuleGroup_123456789012_5e665b1c-1793-4bc0-8d3f-5fbd5d4ee8b7_uniqueid"),
		Priority: 10000000,
		Statement: &awstypes.Statement{
			RuleGroupReferenceStatement: &awstypes.RuleGroupReferenceStatement{},
		},
		OverrideAction: &awstypes.OverrideAction{
			None: &awstypes.NoneAction{},
		},
		VisibilityConfig: &awstypes.VisibilityConfig{
			MetricName: aws.String("shield"),
		},
	})

	got, err := webACLRulesJSONToSet(testRulesJSONConfigRateBased, remote)
	if err != nil {
		t.Fatal(err)
	}
	if got != testRulesJSONConfigRateBased {
		t.Fatalf("Shield Advanced rule not filtered: %s", got)
	}
}

// Nesting depth is not limited: 12 logical levels.
func TestWebACLRulesJSONToSet_deepNesting(t *testing.T) {
	t.Parallel()

	leaf := `{"ByteMatchStatement":{"SearchString":"x","FieldToMatch":{"SingleHeader":{"Name":"X-Deep"}},"TextTransformations":[{"Priority":1,"Type":"LOWERCASE"},{"Priority":0,"Type":"URL_DECODE"}],"PositionalConstraint":"CONTAINS"}}`
	geo := `{"GeoMatchStatement":{"CountryCodes":["US"]}}`
	stmt := leaf
	for i := range 12 {
		if i%2 == 0 {
			stmt = fmt.Sprintf(`{"NotStatement":{"Statement":%s}}`, stmt)
		} else {
			stmt = fmt.Sprintf(`{"AndStatement":{"Statements":[%s,%s]}}`, stmt, geo)
		}
	}
	config := fmt.Sprintf(`[{"Name":"deep","Priority":0,"Action":{"Block":{}},"Statement":%s,"VisibilityConfig":{"CloudWatchMetricsEnabled":false,"MetricName":"deep","SampledRequestsEnabled":false}}]`, stmt)

	rules, err := expandWebACLRulesJSON(config)
	if err != nil {
		t.Fatal(err)
	}
	// As the API returns it: header name in lower case, text transformations in Priority order.
	wire, err := json.Marshal(rules)
	if err != nil {
		t.Fatal(err)
	}
	api := strings.ReplaceAll(string(wire), `"X-Deep"`, `"x-deep"`)
	api = strings.Replace(api, `[{"Priority":1,"Type":"LOWERCASE"},{"Priority":0,"Type":"URL_DECODE"}]`, `[{"Priority":0,"Type":"URL_DECODE"},{"Priority":1,"Type":"LOWERCASE"}]`, 1)

	got, err := webACLRulesJSONToSet(config, apiRules(t, api))
	if err != nil {
		t.Fatal(err)
	}
	if got != config {
		t.Fatalf("not equivalent: %s", got)
	}
	canonical, err := canonicalRulesJSON(expandWebACLRulesJSON, config)
	if err != nil {
		t.Fatal(err)
	}
	if n := strings.Count(canonical, "NotStatement") + strings.Count(canonical, "AndStatement"); n != 12 {
		t.Fatalf("got %d logical levels, want 12", n)
	}
}

func TestRulesJSONStateFunc_idempotent(t *testing.T) {
	t.Parallel()

	for _, expand := range []rulesJSONExpander{expandWebACLRulesJSON, expandRuleGroupRulesJSON} {
		stateFunc, diffSuppressFunc := rulesJSONStateFunc(expand), rulesJSONDiffSuppressFunc(expand)

		for _, config := range []string{testRulesJSONConfigBotControl, testRulesJSONConfigRateBased} {
			a := stateFunc(config)
			if b := stateFunc(a); a != b {
				t.Fatalf("not idempotent:\n%s\n%s", a, b)
			}
			if !diffSuppressFunc("", config, a, nil) {
				t.Fatal("canonical form not equivalent to input")
			}
		}
	}
}

func TestRulesJSONStateFunc_invalid(t *testing.T) {
	t.Parallel()

	stateFunc := rulesJSONStateFunc(expandWebACLRulesJSON)

	// An unknown value is stored as is, as before.
	const unknown = "74D93920-ED26-11E3-AC10-0800200C9A66"
	if got := stateFunc(unknown); got != unknown {
		t.Fatalf("got %q", got)
	}
	// JSON that is not a list of rules is normalized, as before.
	if got := stateFunc(`{ "a" : 1 }`); got != `{"a":1}` {
		t.Fatalf("got %q", got)
	}
}

// A SearchString that is not valid UTF-8 (for example an MD5 digest matched after
// the MD5 text transformation) cannot be written as plain text, so it is kept in
// base64 and does not round-trip.
func TestFlattenRulesJSON_binarySearchString(t *testing.T) {
	t.Parallel()

	remote := apiRules(t, `[{"Name":"md5","Priority":0,"Action":{"Allow":{}},"Statement":{"ByteMatchStatement":{"SearchString":"/F4DjTilcDIIVEHn/nAQsA==","FieldToMatch":{"SingleHeader":{"Name":"authorization"}},"TextTransformations":[{"Priority":0,"Type":"MD5"}],"PositionalConstraint":"EXACTLY"}},"VisibilityConfig":{"CloudWatchMetricsEnabled":false,"MetricName":"md5","SampledRequestsEnabled":false}}]`)
	got, err := flattenRulesJSON(remote)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, `"/F4DjTilcDIIVEHn/nAQsA=="`) {
		t.Fatalf("unexpected value: %s", got)
	}
}
