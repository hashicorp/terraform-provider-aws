// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package wafv2_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/wafv2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
	"github.com/hashicorp/aws-sdk-go-base/v2/servicemocks"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	terraformsdk "github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/provider/sdkv2"
	tfwafv2 "github.com/hashicorp/terraform-provider-aws/internal/service/wafv2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// A rule nested deeper than the rule schema allows can be added to a web ACL out of band,
// with rule_json, or with aws_wafv2_web_acl_rule. Reading the web ACL must not fail on it.
// See https://github.com/hashicorp/terraform-provider-aws/issues/40688.
func TestWebACLRead_ruleNestedDeeperThanSchema(t *testing.T) {
	t.Parallel()
	ctx := t.Context()

	server, _ := testWebACLDeepRuleServer(t)
	meta := testWebACLDeepRuleMeta(ctx, t, server.URL)

	r := tfwafv2.ResourceWebACL()
	d := schema.TestResourceDataRaw(t, r.SchemaMap(), testWebACLDeepRuleConfig(map[string]any{
		names.AttrRule: []any{testWebACLDeepRuleShallowRule("shallow", 1)},
	}))
	d.SetId("web-acl-id")

	state, diags := r.RefreshWithoutUpgrade(ctx, d.State(), meta)
	if diags.HasError() {
		t.Fatalf("read: %v", diags)
	}

	if got, want := testWebACLDeepRuleWarnings(diags), 1; got != want {
		t.Errorf("read warnings = %d, want %d: %v", got, want, diags)
	}

	if got, want := state.Attributes[acctest.CtRulePound], "1"; got != want {
		t.Errorf("rule.# = %s, want %s", got, want)
	}
	for k, v := range state.Attributes {
		if strings.HasSuffix(k, ".name") && strings.HasPrefix(k, "rule.") && v != "shallow" {
			t.Errorf("%s = %s, want only the rule that fits the schema", k, v)
		}
	}
}

// A rule that is not in state because it is nested deeper than the rule schema allows
// must be sent back to UpdateWebACL, so that an unrelated update does not delete it.
// rule_json is authoritative: it replaces every rule.
func TestWebACLUpdate_preservesRuleNestedDeeperThanSchema(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		oldRules     map[string]any
		newRules     map[string]any
		wantDeepRule bool
	}{
		names.AttrRule: {
			oldRules: map[string]any{
				names.AttrRule: []any{testWebACLDeepRuleShallowRule("shallow", 1)},
			},
			newRules: map[string]any{
				names.AttrRule: []any{testWebACLDeepRuleShallowRule("shallow", 2)},
			},
			wantDeepRule: true,
		},
		"rule same name": {
			oldRules: map[string]any{
				names.AttrRule: []any{testWebACLDeepRuleShallowRule("shallow", 1)},
			},
			newRules: map[string]any{
				names.AttrRule: []any{testWebACLDeepRuleShallowRule("shallow", 1), testWebACLDeepRuleShallowRule("deep", 2)},
			},
		},
		"rule_json": {
			oldRules: map[string]any{
				"rule_json": testWebACLDeepRuleShallowRuleJSON("shallow", 1),
			},
			newRules: map[string]any{
				"rule_json": testWebACLDeepRuleShallowRuleJSON("shallow", 2),
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()

			server, updates := testWebACLDeepRuleServer(t)
			meta := testWebACLDeepRuleMeta(ctx, t, server.URL)

			r := tfwafv2.ResourceWebACL()
			old := schema.TestResourceDataRaw(t, r.SchemaMap(), testWebACLDeepRuleConfig(testCase.oldRules))
			old.SetId("web-acl-id")
			old.Set("lock_token", "lock-token-1")

			diff, err := r.Diff(ctx, old.State(), terraformsdk.NewResourceConfigRaw(testWebACLDeepRuleConfig(testCase.newRules)), meta)
			if err != nil {
				t.Fatalf("diff: %s", err)
			}

			if _, diags := r.Apply(ctx, old.State(), diff, meta); diags.HasError() {
				t.Fatalf("apply: %v", diags)
			}

			got := updates()
			if len(got) != 1 {
				t.Fatalf("UpdateWebACL calls = %d, want 1", len(got))
			}
			var deep []string
			for _, rule := range got[0] {
				if rule.Name == "deep" {
					deep = append(deep, rule.Statement)
				}
			}
			switch {
			case testCase.wantDeepRule && (len(deep) != 1 || !strings.Contains(deep[0], "NotStatement")):
				t.Errorf("UpdateWebACL rule deep = %v, want the remote rule sent back unchanged", deep)
			case !testCase.wantDeepRule && slices.ContainsFunc(deep, func(s string) bool { return strings.Contains(s, "NotStatement") }):
				t.Errorf("UpdateWebACL rule deep = %v, want the remote rule not to be sent back", deep)
			}
		})
	}
}

func TestPartitionWebACLRules(t *testing.T) {
	t.Parallel()

	geo := func() awstypes.Statement {
		return awstypes.Statement{GeoMatchStatement: &awstypes.GeoMatchStatement{CountryCodes: []awstypes.CountryCode{awstypes.CountryCodeUs}}}
	}
	// not wraps s in n levels of not_statement.
	not := func(n int, s awstypes.Statement) awstypes.Statement {
		for range n {
			inner := s
			s = awstypes.Statement{NotStatement: &awstypes.NotStatement{Statement: &inner}}
		}
		return s
	}
	rateBased := func(scopeDown awstypes.Statement) awstypes.Statement {
		return awstypes.Statement{RateBasedStatement: &awstypes.RateBasedStatement{
			AggregateKeyType:   awstypes.RateBasedStatementAggregateKeyTypeIp,
			Limit:              aws.Int64(100),
			ScopeDownStatement: &scopeDown,
		}}
	}

	testCases := map[string]struct {
		statement awstypes.Statement
		wantFit   bool
	}{
		"leaf":                      {statement: geo(), wantFit: true},
		"3 levels":                  {statement: not(3, geo()), wantFit: true},
		"4 levels":                  {statement: not(4, geo())},
		"rate-based scope-down 2":   {statement: rateBased(not(2, geo())), wantFit: true},
		"rate-based scope-down 3":   {statement: rateBased(not(3, geo()))},
		"and/or/not mixed 4 levels": {statement: awstypes.Statement{AndStatement: &awstypes.AndStatement{Statements: []awstypes.Statement{geo(), not(1, awstypes.Statement{OrStatement: &awstypes.OrStatement{Statements: []awstypes.Statement{geo(), not(1, geo())}}})}}}},
		"and/or/not mixed 3 levels": {statement: awstypes.Statement{AndStatement: &awstypes.AndStatement{Statements: []awstypes.Statement{geo(), not(1, awstypes.Statement{OrStatement: &awstypes.OrStatement{Statements: []awstypes.Statement{geo(), geo()}}})}}}, wantFit: true},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			fit, deep := tfwafv2.PartitionWebACLRules([]awstypes.Rule{{Name: aws.String(name), Statement: &testCase.statement}})
			if got := len(fit) == 1 && len(deep) == 0; got != testCase.wantFit {
				t.Errorf("fits = %t (fit %d, deep %d), want %t", got, len(fit), len(deep), testCase.wantFit)
			}
		})
	}
}

type testWebACLDeepRuleUpdateRule struct {
	Name      string
	Statement string
}

// testWebACLDeepRuleServer fakes the WAFv2 API. The web ACL holds a rule that fits the
// schema ("shallow") and one with 4 levels of logical statements ("deep").
func testWebACLDeepRuleServer(t *testing.T) (*httptest.Server, func() [][]testWebACLDeepRuleUpdateRule) {
	t.Helper()

	var (
		mu      sync.Mutex
		updates [][]testWebACLDeepRuleUpdateRule
	)
	webACLARN := fmt.Sprintf("arn:%s:wafv2:%s:123456789012:regional/webacl/test/web-acl-id", endpoints.AwsPartitionID, endpoints.UsEast1RegionID)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("reading request: %s", err)
		}
		w.Header().Set("Content-Type", "application/x-amz-json-1.1")
		write := func(s string) {
			if _, err := io.WriteString(w, s); err != nil {
				t.Errorf("writing response: %s", err)
			}
		}

		switch strings.TrimPrefix(r.Header.Get("X-Amz-Target"), "AWSWAF_20190729.") {
		case "GetWebACL":
			write(`{
  "LockToken": "lock-token-2",
  "WebACL": {
    "ARN": "` + webACLARN + `",
    "DefaultAction": {"Allow": {}},
    "Id": "web-acl-id",
    "Name": "test",
    "Rules": [` + testWebACLDeepRuleShallowRuleAPI("shallow", 1) + `, {
      "Name": "deep",
      "Action": {"Block": {}},
      "Priority": 5,
      "Statement": {"AndStatement": {"Statements": [
        {"NotStatement": {"Statement": {"OrStatement": {"Statements": [
          {"NotStatement": {"Statement": {"GeoMatchStatement": {"CountryCodes": ["US"]}}}},
          {"GeoMatchStatement": {"CountryCodes": ["CA"]}}
        ]}}}},
        {"GeoMatchStatement": {"CountryCodes": ["NL"]}}
      ]}},
      "VisibilityConfig": {"CloudWatchMetricsEnabled": false, "MetricName": "deep", "SampledRequestsEnabled": false}
    }],
    "VisibilityConfig": {"CloudWatchMetricsEnabled": false, "MetricName": "test", "SampledRequestsEnabled": false}
  }
}`)
		case "UpdateWebACL":
			var input struct {
				Rules []struct {
					Name      string
					Statement json.RawMessage
				}
			}
			if err := json.Unmarshal(body, &input); err != nil {
				t.Errorf("unmarshalling UpdateWebACL request: %s", err)
			}
			var rules []testWebACLDeepRuleUpdateRule
			for _, rule := range input.Rules {
				rules = append(rules, testWebACLDeepRuleUpdateRule{Name: rule.Name, Statement: string(rule.Statement)})
			}
			mu.Lock()
			updates = append(updates, rules)
			mu.Unlock()
			write(`{"NextLockToken": "lock-token-3"}`)
		default:
			write(`{}`)
		}
	}))
	t.Cleanup(server.Close)

	return server, func() [][]testWebACLDeepRuleUpdateRule {
		mu.Lock()
		defer mu.Unlock()
		return slices.Clone(updates)
	}
}

func testWebACLDeepRuleMeta(ctx context.Context, t *testing.T, endpoint string) any {
	t.Helper()

	p, err := sdkv2.NewProvider(ctx)
	if err != nil {
		t.Fatal(err)
	}
	p.TerraformVersion = "1.0.0"
	diags := p.Configure(ctx, terraformsdk.NewResourceConfigRaw(map[string]any{
		names.AttrAccessKey:                 servicemocks.MockStaticAccessKey,
		names.AttrSecretKey:                 servicemocks.MockStaticSecretKey,
		names.AttrRegion:                    endpoints.UsEast1RegionID,
		names.AttrSkipCredentialsValidation: true,
		names.AttrSkipRequestingAccountID:   true,
		names.AttrEndpoints: []any{map[string]any{
			"wafv2": endpoint,
		}},
	}))
	if diags.HasError() {
		t.Fatalf("configuring provider: %v", diags)
	}

	return p.Meta()
}

func testWebACLDeepRuleWarnings(diags diag.Diagnostics) int {
	var n int
	for _, d := range diags {
		if d.Severity == diag.Warning {
			n++
		}
	}
	return n
}

func testWebACLDeepRuleConfig(rules map[string]any) map[string]any {
	config := map[string]any{
		names.AttrName:  "test",
		names.AttrScope: "REGIONAL",
		names.AttrDefaultAction: []any{map[string]any{
			"allow": []any{map[string]any{}},
		}},
		"visibility_config": []any{map[string]any{
			"cloudwatch_metrics_enabled": false,
			names.AttrMetricName:         "test",
			"sampled_requests_enabled":   false,
		}},
	}
	maps.Copy(config, rules)
	return config
}

func testWebACLDeepRuleShallowRule(name string, priority int) map[string]any {
	return map[string]any{
		names.AttrName:     name,
		names.AttrPriority: priority,
		names.AttrAction: []any{map[string]any{
			"count": []any{map[string]any{}},
		}},
		"statement": []any{map[string]any{
			"geo_match_statement": []any{map[string]any{
				"country_codes": []any{"US"},
			}},
		}},
		"visibility_config": []any{map[string]any{
			"cloudwatch_metrics_enabled": false,
			names.AttrMetricName:         name,
			"sampled_requests_enabled":   false,
		}},
	}
}

func testWebACLDeepRuleShallowRuleAPI(name string, priority int) string {
	return fmt.Sprintf(`{
  "Name": %[1]q,
  "Priority": %[2]d,
  "Action": {"Count": {}},
  "Statement": {"GeoMatchStatement": {"CountryCodes": ["US"]}},
  "VisibilityConfig": {"CloudWatchMetricsEnabled": false, "MetricName": %[1]q, "SampledRequestsEnabled": false}
}`, name, priority)
}

func testWebACLDeepRuleShallowRuleJSON(name string, priority int) string {
	return "[" + testWebACLDeepRuleShallowRuleAPI(name, priority) + "]"
}
