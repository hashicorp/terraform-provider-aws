// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssmcontacts

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssmcontacts/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func expandContactChannelAddress(deliveryAddress []any) *types.ContactChannelAddress {
	if len(deliveryAddress) == 0 || deliveryAddress[0] == nil {
		return nil
	}

	m := deliveryAddress[0].(map[string]any)

	contactChannelAddress := &types.ContactChannelAddress{}

	if v, ok := m["simple_address"].(string); ok {
		contactChannelAddress.SimpleAddress = aws.String(v)
	}

	return contactChannelAddress
}

func flattenContactChannelAddress(contactChannelAddress *types.ContactChannelAddress) []any {
	m := map[string]any{}

	if v := contactChannelAddress.SimpleAddress; v != nil {
		m["simple_address"] = aws.ToString(v)
	}

	return []any{m}
}

func expandStages(stages []any) []types.Stage {
	var stageList []types.Stage

	for _, stage := range stages {
		s := types.Stage{}

		stageData := stage.(map[string]any)

		if v, ok := stageData["duration_in_minutes"].(int); ok {
			s.DurationInMinutes = aws.Int32(int32(v))
		}

		if v, ok := stageData[names.AttrTarget].([]any); ok {
			s.Targets = expandTargets(v)
		}

		stageList = append(stageList, s)
	}

	return stageList
}

func flattenStages(stages []types.Stage) []any {
	var result []any

	for _, stage := range stages {
		s := map[string]any{}

		if v := stage.DurationInMinutes; v != nil {
			s["duration_in_minutes"] = aws.ToInt32(v)
		}

		if v := stage.Targets; v != nil {
			s[names.AttrTarget] = flattenTargets(v)
		}

		result = append(result, s)
	}

	return result
}

func expandTargets(targets []any) []types.Target {
	targetList := make([]types.Target, 0)

	for _, target := range targets {
		if target == nil {
			continue
		}

		t := types.Target{}

		targetData := target.(map[string]any)

		if v, ok := targetData["channel_target_info"].([]any); ok {
			t.ChannelTargetInfo = expandChannelTargetInfo(v)
		}

		if v, ok := targetData["contact_target_info"].([]any); ok {
			t.ContactTargetInfo = expandContactTargetInfo(v)
		}

		targetList = append(targetList, t)
	}

	return targetList
}

func flattenTargets(targets []types.Target) []any {
	result := make([]any, 0)

	for _, target := range targets {
		t := map[string]any{}

		if v := target.ChannelTargetInfo; v != nil {
			t["channel_target_info"] = flattenChannelTargetInfo(v)
		}

		if v := target.ContactTargetInfo; v != nil {
			t["contact_target_info"] = flattenContactTargetInfo(v)
		}

		result = append(result, t)
	}

	return result
}

func expandChannelTargetInfo(channelTargetInfo []any) *types.ChannelTargetInfo {
	if len(channelTargetInfo) == 0 {
		return nil
	}

	c := &types.ChannelTargetInfo{}

	channelTargetInfoData := channelTargetInfo[0].(map[string]any)

	if v, ok := channelTargetInfoData["contact_channel_id"].(string); ok && v != "" {
		c.ContactChannelId = aws.String(v)
	}

	if v, ok := channelTargetInfoData["retry_interval_in_minutes"].(int); ok {
		c.RetryIntervalInMinutes = aws.Int32(int32(v))
	}

	return c
}

func flattenChannelTargetInfo(channelTargetInfo *types.ChannelTargetInfo) []any {
	var result []any

	c := make(map[string]any)

	if v := channelTargetInfo.ContactChannelId; v != nil {
		c["contact_channel_id"] = aws.ToString(v)
	}

	if v := channelTargetInfo.RetryIntervalInMinutes; v != nil {
		c["retry_interval_in_minutes"] = aws.ToInt32(v)
	}

	result = append(result, c)

	return result
}

func expandContactTargetInfo(contactTargetInfo []any) *types.ContactTargetInfo {
	if len(contactTargetInfo) == 0 {
		return nil
	}

	c := &types.ContactTargetInfo{}

	contactTargetInfoData := contactTargetInfo[0].(map[string]any)

	if v, ok := contactTargetInfoData["is_essential"].(bool); ok {
		c.IsEssential = aws.Bool(v)
	}

	if v, ok := contactTargetInfoData["contact_id"].(string); ok && v != "" {
		c.ContactId = aws.String(v)
	}

	return c
}

func flattenContactTargetInfo(contactTargetInfo *types.ContactTargetInfo) []any {
	var result []any

	c := make(map[string]any)

	if v := contactTargetInfo.IsEssential; v != nil {
		c["is_essential"] = aws.ToBool(v)
	}

	if v := contactTargetInfo.ContactId; v != nil {
		c["contact_id"] = aws.ToString(v)
	}

	result = append(result, c)

	return result
}
