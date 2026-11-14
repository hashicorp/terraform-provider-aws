// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package wafv2_test

import (
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
	"github.com/hashicorp/aws-sdk-go-base/v2/servicemocks"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	terraformsdk "github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/provider/sdkv2"
	tfwafv2 "github.com/hashicorp/terraform-provider-aws/internal/service/wafv2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// The Shield Advanced automatic application layer DDoS mitigation rule is added to a web ACL by AWS.
// An update must send it back to UpdateWebACL, whether the rules are configured with rule or rule_json.
// See https://github.com/hashicorp/terraform-provider-aws/issues/22869.
func TestWebACLUpdate_preservesShieldRule(t *testing.T) {
	t.Parallel()

	const shieldRuleName = "ShieldMitigationRuleGroup_123456789012_d2b7c2a5-6f3e-4b8e-9c1d-0a1b2c3d4e5f_example"

	testCases := map[string]struct {
		oldRules map[string]any
		newRules map[string]any
	}{
		names.AttrRule: {
			oldRules: map[string]any{
				names.AttrRule: []any{testWebACLShieldRuleRule("rule-1", 1)},
			},
			newRules: map[string]any{
				names.AttrRule: []any{testWebACLShieldRuleRule("rule-1", 2)},
			},
		},
		"rule_json": {
			oldRules: map[string]any{
				"rule_json": testWebACLShieldRuleRuleJSON("rule-1", 1),
			},
			newRules: map[string]any{
				"rule_json": testWebACLShieldRuleRuleJSON("rule-1", 2),
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()

			var (
				mu          sync.Mutex
				updateRules []string
			)
			webACLARN := fmt.Sprintf("arn:%s:wafv2:%s:123456789012:regional/webacl/test/web-acl-id", endpoints.AwsPartitionID, endpoints.UsEast1RegionID)
			ruleGroupARN := fmt.Sprintf("arn:%s:wafv2:%s:123456789012:regional/rulegroup/ShieldMitigationRuleGroup/rule-group-id", endpoints.AwsPartitionID, endpoints.UsEast1RegionID)
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
    "Rules": [{
      "Name": "` + shieldRuleName + `",
      "OverrideAction": {"None": {}},
      "Priority": 10000000,
      "Statement": {"RuleGroupReferenceStatement": {"ARN": "` + ruleGroupARN + `"}},
      "VisibilityConfig": {"CloudWatchMetricsEnabled": true, "MetricName": "shield", "SampledRequestsEnabled": true}
    }],
    "VisibilityConfig": {"CloudWatchMetricsEnabled": false, "MetricName": "test", "SampledRequestsEnabled": false}
  }
}`)
				case "UpdateWebACL":
					var input struct {
						Rules []struct {
							Name string
						}
					}
					if err := json.Unmarshal(body, &input); err != nil {
						t.Errorf("unmarshalling UpdateWebACL request: %s", err)
					}
					mu.Lock()
					for _, rule := range input.Rules {
						updateRules = append(updateRules, rule.Name)
					}
					mu.Unlock()
					write(`{"NextLockToken": "lock-token-3"}`)
				default:
					write(`{}`)
				}
			}))
			t.Cleanup(server.Close)

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
					"wafv2": server.URL,
				}},
			}))
			if diags.HasError() {
				t.Fatalf("configuring provider: %v", diags)
			}
			meta := p.Meta()

			r := tfwafv2.ResourceWebACL()
			old := schema.TestResourceDataRaw(t, r.SchemaMap(), testWebACLShieldRuleConfig(testCase.oldRules))
			old.SetId("web-acl-id")
			old.Set("lock_token", "lock-token-1")

			diff, err := r.Diff(ctx, old.State(), terraformsdk.NewResourceConfigRaw(testWebACLShieldRuleConfig(testCase.newRules)), meta)
			if err != nil {
				t.Fatalf("diff: %s", err)
			}

			if _, diags := r.Apply(ctx, old.State(), diff, meta); diags.HasError() {
				t.Fatalf("apply: %v", diags)
			}

			if !strings.Contains(strings.Join(updateRules, ","), shieldRuleName) {
				t.Errorf("UpdateWebACL rules = %v, want the Shield Advanced rule %q to be preserved", updateRules, shieldRuleName)
			}
		})
	}
}

func testWebACLShieldRuleConfig(rules map[string]any) map[string]any {
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

func testWebACLShieldRuleRule(name string, priority int) map[string]any {
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

func testWebACLShieldRuleRuleJSON(name string, priority int) string {
	return fmt.Sprintf(`[{
  "Name": %[1]q,
  "Priority": %[2]d,
  "Action": {"Count": {}},
  "Statement": {"GeoMatchStatement": {"CountryCodes": ["US"]}},
  "VisibilityConfig": {"CloudWatchMetricsEnabled": false, "MetricName": %[1]q, "SampledRequestsEnabled": false}
}]`, name, priority)
}
