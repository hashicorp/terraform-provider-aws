// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package waf

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func ExpandAction(l []interface{}) *waf.WafAction {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	return &waf.WafAction{
		Type: aws.String(m[names.AttrType].(string)),
	}
}

func expandOverrideAction(l []interface{}) *waf.WafOverrideAction {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	return &waf.WafOverrideAction{
		Type: aws.String(m[names.AttrType].(string)),
	}
}

func ExpandWebACLUpdate(updateAction string, aclRule map[string]interface{}) *waf.WebACLUpdate {
	var rule *waf.ActivatedRule

	switch aclRule[names.AttrType].(string) {
	case waf.WafRuleTypeGroup:
		rule = &waf.ActivatedRule{
			OverrideAction: expandOverrideAction(aclRule["override_action"].([]interface{})),
			Priority:       aws.Int64(int64(aclRule["priority"].(int))),
			RuleId:         aws.String(aclRule["rule_id"].(string)),
			Type:           aws.String(aclRule[names.AttrType].(string)),
		}
	default:
		rule = &waf.ActivatedRule{
			Action:   ExpandAction(aclRule["action"].([]interface{})),
			Priority: aws.Int64(int64(aclRule["priority"].(int))),
			RuleId:   aws.String(aclRule["rule_id"].(string)),
			Type:     aws.String(aclRule[names.AttrType].(string)),
		}
	}

	update := &waf.WebACLUpdate{
		Action:        aws.String(updateAction),
		ActivatedRule: rule,
	}

	return update
}

func FlattenAction(n *waf.WafAction) []map[string]interface{} {
	if n == nil {
		return nil
	}

	result := map[string]interface{}{
		names.AttrType: aws.StringValue(n.Type),
	}

	return []map[string]interface{}{result}
}

func FlattenWebACLRules(ts []*waf.ActivatedRule) []map[string]interface{} {
	out := make([]map[string]interface{}, len(ts))
	for i, r := range ts {
		m := make(map[string]interface{})

		switch aws.StringValue(r.Type) {
		case waf.WafRuleTypeGroup:
			actionMap := map[string]interface{}{
				names.AttrType: aws.StringValue(r.OverrideAction.Type),
			}
			m["override_action"] = []map[string]interface{}{actionMap}
		default:
			actionMap := map[string]interface{}{
				names.AttrType: aws.StringValue(r.Action.Type),
			}
			m["action"] = []map[string]interface{}{actionMap}
		}

		m["priority"] = int(aws.Int64Value(r.Priority))
		m["rule_id"] = aws.StringValue(r.RuleId)
		m[names.AttrType] = aws.StringValue(r.Type)
		out[i] = m
	}
	return out
}

func ExpandFieldToMatch(d map[string]interface{}) *waf.FieldToMatch {
	ftm := &waf.FieldToMatch{
		Type: aws.String(d[names.AttrType].(string)),
	}
	if data, ok := d["data"].(string); ok && data != "" {
		ftm.Data = aws.String(data)
	}
	return ftm
}

func FlattenFieldToMatch(fm *waf.FieldToMatch) []interface{} {
	m := make(map[string]interface{})
	if fm.Data != nil {
		m["data"] = aws.StringValue(fm.Data)
	}
	if fm.Type != nil {
		m[names.AttrType] = aws.StringValue(fm.Type)
	}
	return []interface{}{m}
}
