// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package waf

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/waf/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func ExpandAction(l []interface{}) *awstypes.WafAction {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	return &awstypes.WafAction{
		Type: awstypes.WafActionType(m[names.AttrType].(string)),
	}
}

func expandOverrideAction(l []interface{}) *awstypes.WafOverrideAction {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	return &awstypes.WafOverrideAction{
		Type: awstypes.WafOverrideActionType(m[names.AttrType].(string)),
	}
}

func ExpandWebACLUpdate(updateAction string, aclRule map[string]interface{}) awstypes.WebACLUpdate {
	var rule *awstypes.ActivatedRule

	switch aclRule[names.AttrType].(string) {
	case string(awstypes.WafRuleTypeGroup):
		rule = &awstypes.ActivatedRule{
			OverrideAction: expandOverrideAction(aclRule["override_action"].([]interface{})),
			Priority:       aws.Int32(int32(aclRule["priority"].(int))),
			RuleId:         aws.String(aclRule["rule_id"].(string)),
			Type:           awstypes.WafRuleType(aclRule[names.AttrType].(string)),
		}
	default:
		rule = &awstypes.ActivatedRule{
			Action:   ExpandAction(aclRule["action"].([]interface{})),
			Priority: aws.Int32(int32(aclRule["priority"].(int))),
			RuleId:   aws.String(aclRule["rule_id"].(string)),
			Type:     awstypes.WafRuleType(aclRule[names.AttrType].(string)),
		}
	}

	update := awstypes.WebACLUpdate{
		Action:        awstypes.ChangeAction(updateAction),
		ActivatedRule: rule,
	}

	return update
}

func FlattenAction(n *awstypes.WafAction) []map[string]interface{} {
	if n == nil {
		return nil
	}

	result := map[string]interface{}{
		names.AttrType: string(n.Type),
	}

	return []map[string]interface{}{result}
}

func FlattenWebACLRules(ts []awstypes.ActivatedRule) []map[string]interface{} {
	out := make([]map[string]interface{}, len(ts))
	for i, r := range ts {
		m := make(map[string]interface{})

		switch r.Type {
		case awstypes.WafRuleTypeGroup:
			actionMap := map[string]interface{}{
				names.AttrType: r.OverrideAction.Type,
			}
			m["override_action"] = []map[string]interface{}{actionMap}
		default:
			actionMap := map[string]interface{}{
				names.AttrType: r.Action.Type,
			}
			m["action"] = []map[string]interface{}{actionMap}
		}

		m["priority"] = r.Priority
		m["rule_id"] = aws.ToString(r.RuleId)
		m[names.AttrType] = string(r.Type)
		out[i] = m
	}
	return out
}

func ExpandFieldToMatch(d map[string]interface{}) *awstypes.FieldToMatch {
	ftm := &awstypes.FieldToMatch{
		Type: awstypes.MatchFieldType(d[names.AttrType].(string)),
	}
	if data, ok := d["data"].(string); ok && data != "" {
		ftm.Data = aws.String(data)
	}
	return ftm
}

func FlattenFieldToMatch(fm *awstypes.FieldToMatch) []interface{} {
	m := make(map[string]interface{})
	if fm.Data != nil {
		m["data"] = aws.ToString(fm.Data)
	}

	m[names.AttrType] = string(fm.Type)

	return []interface{}{m}
}
