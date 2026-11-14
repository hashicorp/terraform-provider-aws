// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package wafv2

import (
	"slices"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/wafv2/types"
)

const testShieldMitigationRuleName = "ShieldMitigationRuleGroup_123456789012_d2b7c2a5-6f3e-4b8e-9c1d-0a1b2c3d4e5f_example"

func testShieldRuleGroupReferenceRule(name string, priority int32) awstypes.Rule {
	return awstypes.Rule{
		Name:     aws.String(name),
		Priority: priority,
		Statement: &awstypes.Statement{
			RuleGroupReferenceStatement: &awstypes.RuleGroupReferenceStatement{
				ARN: aws.String("rule-group-arn"),
			},
		},
	}
}

func testShieldManagedRuleGroupRule(name string, priority int32, vendorName, ruleGroupName string) awstypes.Rule {
	return awstypes.Rule{
		Name:     aws.String(name),
		Priority: priority,
		Statement: &awstypes.Statement{
			ManagedRuleGroupStatement: &awstypes.ManagedRuleGroupStatement{
				Name:       aws.String(ruleGroupName),
				VendorName: aws.String(vendorName),
			},
		},
	}
}

func testRuleNames(rules []awstypes.Rule) []string {
	var names []string
	for _, r := range rules {
		names = append(names, aws.ToString(r.Name))
	}
	return names
}

func TestIsShieldRule(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		rule awstypes.Rule
		want bool
	}{
		"ShieldMitigationRuleGroup": {
			rule: testShieldRuleGroupReferenceRule(testShieldMitigationRuleName, 10000000),
			want: true,
		},
		"AntiDDoS rule added by Shield": {
			rule: testShieldManagedRuleGroupRule(shieldAntiDDoSRuleName, shieldAntiDDoSRulePriority, "AWS", shieldAntiDDoSRuleName),
			want: true,
		},
		"AntiDDoS rule at another priority": {
			rule: testShieldManagedRuleGroupRule(shieldAntiDDoSRuleName, 1, "AWS", shieldAntiDDoSRuleName),
		},
		"AntiDDoS rule with another name": {
			rule: testShieldManagedRuleGroupRule("anti-ddos", shieldAntiDDoSRulePriority, "AWS", shieldAntiDDoSRuleName),
		},
		"other managed rule group with the AntiDDoS rule name": {
			rule: testShieldManagedRuleGroupRule(shieldAntiDDoSRuleName, shieldAntiDDoSRulePriority, "AWS", "AWSManagedRulesCommonRuleSet"),
		},
		"other vendor": {
			rule: testShieldManagedRuleGroupRule(shieldAntiDDoSRuleName, shieldAntiDDoSRulePriority, "Example", shieldAntiDDoSRuleName),
		},
		"AntiDDoS rule name without a statement": {
			rule: awstypes.Rule{Name: aws.String(shieldAntiDDoSRuleName), Priority: shieldAntiDDoSRulePriority},
		},
		"customer rule": {
			rule: testShieldRuleGroupReferenceRule("rule-1", 1),
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if got := isShieldRule(testCase.rule); got != testCase.want {
				t.Errorf("isShieldRule() = %t, want %t", got, testCase.want)
			}
		})
	}
}

func TestFilterWebACLRules(t *testing.T) {
	t.Parallel()

	customer := testShieldRuleGroupReferenceRule("rule-1", 1)
	mitigation := testShieldRuleGroupReferenceRule(testShieldMitigationRuleName, 10000000)
	antiDDoS := testShieldManagedRuleGroupRule(shieldAntiDDoSRuleName, shieldAntiDDoSRulePriority, "AWS", shieldAntiDDoSRuleName)

	testCases := map[string]struct {
		rules       []awstypes.Rule
		configRules []awstypes.Rule
		want        []string
	}{
		"no Shield rules": {
			rules:       []awstypes.Rule{customer},
			configRules: []awstypes.Rule{customer},
			want:        []string{"rule-1"},
		},
		"both Shield rules removed": {
			rules:       []awstypes.Rule{customer, antiDDoS, mitigation},
			configRules: []awstypes.Rule{customer},
			want:        []string{"rule-1"},
		},
		"only the AntiDDoS rule": {
			rules:       []awstypes.Rule{customer, antiDDoS},
			configRules: []awstypes.Rule{customer},
			want:        []string{"rule-1"},
		},
		"AntiDDoS rule declared in the configuration is kept": {
			rules:       []awstypes.Rule{customer, antiDDoS, mitigation},
			configRules: []awstypes.Rule{customer, antiDDoS},
			want:        []string{"rule-1", shieldAntiDDoSRuleName},
		},
		"customer AntiDDoS rule at another priority is kept": {
			rules:       []awstypes.Rule{testShieldManagedRuleGroupRule(shieldAntiDDoSRuleName, 5, "AWS", shieldAntiDDoSRuleName)},
			configRules: nil,
			want:        []string{shieldAntiDDoSRuleName},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if got := testRuleNames(filterWebACLRules(testCase.rules, testCase.configRules)); !slices.Equal(got, testCase.want) {
				t.Errorf("filterWebACLRules() = %v, want %v", got, testCase.want)
			}
		})
	}
}

func TestAppendShieldRules(t *testing.T) {
	t.Parallel()

	customer := testShieldRuleGroupReferenceRule("rule-1", 1)
	mitigation := testShieldRuleGroupReferenceRule(testShieldMitigationRuleName, 10000000)
	antiDDoS := testShieldManagedRuleGroupRule(shieldAntiDDoSRuleName, shieldAntiDDoSRulePriority, "AWS", shieldAntiDDoSRuleName)

	testCases := map[string]struct {
		rules       []awstypes.Rule
		remoteRules []awstypes.Rule
		want        []string
	}{
		"no Shield rules": {
			rules:       []awstypes.Rule{customer},
			remoteRules: []awstypes.Rule{customer},
			want:        []string{"rule-1"},
		},
		"both Shield rules added": {
			rules:       []awstypes.Rule{customer},
			remoteRules: []awstypes.Rule{customer, mitigation, antiDDoS},
			want:        []string{"rule-1", testShieldMitigationRuleName, shieldAntiDDoSRuleName},
		},
		"Shield rule declared in the configuration is not added twice": {
			rules:       []awstypes.Rule{customer, mitigation},
			remoteRules: []awstypes.Rule{customer, mitigation, antiDDoS},
			want:        []string{"rule-1", testShieldMitigationRuleName, shieldAntiDDoSRuleName},
		},
		"customer rules removed from the configuration are not added": {
			rules:       nil,
			remoteRules: []awstypes.Rule{customer, antiDDoS},
			want:        []string{shieldAntiDDoSRuleName},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if got := testRuleNames(appendShieldRules(testCase.rules, testCase.remoteRules)); !slices.Equal(got, testCase.want) {
				t.Errorf("appendShieldRules() = %v, want %v", got, testCase.want)
			}
		})
	}
}
