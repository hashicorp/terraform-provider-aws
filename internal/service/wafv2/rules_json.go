// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package wafv2

import (
	"bytes"
	"cmp"
	"encoding/json"
	"reflect"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/wafv2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// Drift detection for aws_wafv2_web_acl.rule_json and aws_wafv2_rule_group.rules_json.
//
// Both the configured JSON and the rules returned by GetWebACL / GetRuleGroup are
// decoded into []awstypes.Rule, normalized (defaults AWS fills in, case it folds,
// list order it imposes) and rendered to one canonical JSON string. Two values are
// equivalent when their canonical forms are equal.

type rulesJSONExpander func(string) ([]awstypes.Rule, error)

// rulesJSONStateFunc stores the canonical form of the configured rules.
func rulesJSONStateFunc(expand rulesJSONExpander) schema.SchemaStateFunc {
	return func(v any) string {
		s, err := canonicalRulesJSON(expand, v.(string))
		if err != nil {
			// e.g. the value is unknown. Keep the previous behavior.
			json, _ := structure.NormalizeJsonString(v)
			return json
		}
		return s
	}
}

func rulesJSONDiffSuppressFunc(expand rulesJSONExpander) schema.SchemaDiffSuppressFunc {
	return func(k, old, new string, d *schema.ResourceData) bool {
		co, err := canonicalRulesJSON(expand, old)
		if err != nil {
			return verify.SuppressEquivalentJSONDiffs(k, old, new, d)
		}
		cn, err := canonicalRulesJSON(expand, new)
		if err != nil {
			return verify.SuppressEquivalentJSONDiffs(k, old, new, d)
		}
		return co == cn
	}
}

func canonicalRulesJSON(expand rulesJSONExpander, v string) (string, error) {
	rules, err := expand(v)
	if err != nil {
		return "", err
	}
	return flattenRulesJSON(rules)
}

// rulesJSONToSet returns the value to store for rules_json / rule_json in Read.
// It keeps the prior value when the remote rules are equivalent, so state does not
// churn, and otherwise returns the canonical remote rules so that the plan shows drift.
func rulesJSONToSet(expand rulesJSONExpander, prior string, remote []awstypes.Rule) (string, error) {
	canonicalRemote, err := flattenRulesJSON(remote)
	if err != nil {
		return "", err
	}

	if canonicalPrior, err := canonicalRulesJSON(expand, prior); err == nil && canonicalPrior == canonicalRemote {
		return prior, nil
	}

	return canonicalRemote, nil
}

func webACLRulesJSONToSet(prior string, remote []awstypes.Rule) (string, error) {
	// Remove the AWS-added Shield Advanced auto mitigation rule, as is done for `rule`.
	configRules, _ := expandWebACLRulesJSON(prior)
	remote = filterWebACLRules(remote, configRules)

	return rulesJSONToSet(expandWebACLRulesJSON, prior, remote)
}

func ruleGroupRulesJSONToSet(prior string, remote []awstypes.Rule) (string, error) {
	return rulesJSONToSet(expandRuleGroupRulesJSON, prior, remote)
}

// flattenRulesJSON renders rules as canonical JSON: normalized, sorted keys,
// no null, "" or [] members, and ByteMatchStatement.SearchString in the plain-text
// form that rule_json / rules_json accept on input.
func flattenRulesJSON(rules []awstypes.Rule) (string, error) {
	// Deep copy through JSON so that the caller's rules are not modified.
	b, err := json.Marshal(rules)
	if err != nil {
		return "", err
	}
	var cp []awstypes.Rule
	if err := json.Unmarshal(b, &cp); err != nil {
		return "", err
	}

	normalizeRules(cp)

	if b, err = json.Marshal(cp); err != nil {
		return "", err
	}
	var tree any
	if err := json.Unmarshal(b, &tree); err != nil {
		return "", err
	}
	tree = pruneRulesJSON(tree, "")

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(tree); err != nil {
		return "", err
	}

	return strings.TrimSuffix(buf.String(), "\n"), nil
}

// normalizeRules rewrites rules in place into canonical form. It never changes
// the meaning of a rule as AWS evaluates it.
func normalizeRules(rules []awstypes.Rule) {
	// Rules are evaluated by Priority, which is unique within a web ACL or rule group.
	slices.SortStableFunc(rules, func(a, b awstypes.Rule) int {
		return cmp.Compare(a.Priority, b.Priority)
	})

	for i := range rules {
		normalizeRulesValue(reflect.ValueOf(&rules[i]))
	}
}

var textTransformationsType = reflect.TypeFor[[]awstypes.TextTransformation]()

// normalizeRulesValue walks any SDK value by reflection, so nesting depth is
// unbounded and new statement types are traversed without code changes.
func normalizeRulesValue(v reflect.Value) {
	switch v.Kind() {
	case reflect.Pointer:
		if v.IsNil() {
			return
		}
		applyRulesJSONDefaults(v.Interface())
		normalizeRulesValue(v.Elem())
	case reflect.Struct:
		for _, f := range v.Fields() {
			if !f.CanSet() {
				continue // Unexported.
			}
			if f.Type() == textTransformationsType {
				// Text transformations are applied by Priority.
				slices.SortStableFunc(f.Interface().([]awstypes.TextTransformation), func(a, b awstypes.TextTransformation) int {
					return cmp.Compare(a.Priority, b.Priority)
				})
			}
			if f.Kind() == reflect.Struct {
				normalizeRulesValue(f.Addr())
				continue
			}
			normalizeRulesValue(f)
		}
	case reflect.Slice:
		for i := range v.Len() {
			if e := v.Index(i); e.Kind() == reflect.Struct {
				normalizeRulesValue(e.Addr())
			} else {
				normalizeRulesValue(e)
			}
		}
	}
}

// applyRulesJSONDefaults fills in values that AWS adds when they are omitted, and
// folds values that AWS returns in a different case.
func applyRulesJSONDefaults(p any) {
	switch x := p.(type) {
	case *awstypes.SingleHeader:
		// The name is returned in lower case.
		if x.Name != nil {
			x.Name = aws.String(strings.ToLower(aws.ToString(x.Name)))
		}
	case *awstypes.SingleQueryArgument:
		if x.Name != nil {
			x.Name = aws.String(strings.ToLower(aws.ToString(x.Name)))
		}
	case *awstypes.Body:
		if x.OversizeHandling == "" {
			x.OversizeHandling = awstypes.OversizeHandlingContinue
		}
	case *awstypes.JsonBody:
		if x.OversizeHandling == "" {
			x.OversizeHandling = awstypes.OversizeHandlingContinue
		}
	case *awstypes.UriFragment:
		if x.FallbackBehavior == "" {
			x.FallbackBehavior = awstypes.FallbackBehaviorMatch
		}
	case *awstypes.RateBasedStatement:
		if x.AggregateKeyType == "" {
			x.AggregateKeyType = awstypes.RateBasedStatementAggregateKeyTypeIp
		}
		if x.EvaluationWindowSec == 0 {
			x.EvaluationWindowSec = 300
		}
	case *awstypes.AWSManagedRulesBotControlRuleSet:
		// TARGETED returns EnableMachineLearning=true when it is not set. COMMON returns nothing.
		if x.EnableMachineLearning == nil && x.InspectionLevel == awstypes.InspectionLevelTargeted {
			x.EnableMachineLearning = aws.Bool(true)
		}
	case *awstypes.AWSManagedRulesAntiDDoSRuleSet:
		if x.SensitivityToBlock == "" {
			x.SensitivityToBlock = awstypes.SensitivityToActLow
		}
	case *awstypes.ClientSideAction:
		if x.Sensitivity == "" {
			x.Sensitivity = awstypes.SensitivityToActHigh
		}
	}
}

func pruneRulesJSON(v any, parent string) any {
	switch x := v.(type) {
	case map[string]any:
		for k, e := range x {
			if parent == "ByteMatchStatement" && k == "SearchString" {
				if s, ok := e.(string); ok {
					if raw, err := decodeJSONBase64(s); err == nil && utf8.Valid(raw) {
						x[k] = string(raw)
						continue
					}
				}
			}
			e = pruneRulesJSON(e, k)
			if isEmptyRulesJSONValue(e) {
				delete(x, k)
				continue
			}
			x[k] = e
		}
		// An empty object is meaningful ({"Block":{}}, {"UriPath":{}}) and is kept.
		return x
	case []any:
		for i := range x {
			x[i] = pruneRulesJSON(x[i], parent)
		}
		return x
	}

	return v
}

func isEmptyRulesJSONValue(v any) bool {
	switch x := v.(type) {
	case nil:
		return true
	case string:
		return x == ""
	case []any:
		return len(x) == 0
	}

	return false
}

// decodeJSONBase64 decodes a []byte value as encoding/json marshals it.
func decodeJSONBase64(s string) ([]byte, error) {
	var b []byte
	err := json.Unmarshal([]byte(`"`+s+`"`), &b)
	return b, err
}
