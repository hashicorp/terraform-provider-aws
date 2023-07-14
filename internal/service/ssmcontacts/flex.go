// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssmcontacts

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssmcontacts/types"
)

func expandContactChannelAddress(deliveryAddress []interface{}) *types.ContactChannelAddress {
	if len(deliveryAddress) == 0 || deliveryAddress[0] == nil {
		return nil
	}

	m := deliveryAddress[0].(map[string]interface{})

	contactChannelAddress := &types.ContactChannelAddress{}

	if v, ok := m["simple_address"].(string); ok {
		contactChannelAddress.SimpleAddress = aws.String(v)
	}

	return contactChannelAddress
}

func flattenContactChannelAddress(contactChannelAddress *types.ContactChannelAddress) []interface{} {
	m := map[string]interface{}{}

	if v := contactChannelAddress.SimpleAddress; v != nil {
		m["simple_address"] = aws.ToString(v)
	}

	return []interface{}{m}
}

func expandStages(stages []interface{}) []types.Stage {
	var stageList []types.Stage

	for _, stage := range stages {
		s := types.Stage{}

		stageData := stage.(map[string]interface{})

		if v, ok := stageData["duration_in_minutes"].(int); ok {
			s.DurationInMinutes = aws.Int32(int32(v))
		}

		if v, ok := stageData["target"].([]interface{}); ok {
			s.Targets = expandTargets(v)
		}

		stageList = append(stageList, s)
	}

	return stageList
}

func flattenStages(stages []types.Stage) []interface{} {
	var result []interface{}

	for _, stage := range stages {
		s := map[string]interface{}{}

		if v := stage.DurationInMinutes; v != nil {
			s["duration_in_minutes"] = aws.ToInt32(v)
		}

		if v := stage.Targets; v != nil {
			s["target"] = flattenTargets(v)
		}

		result = append(result, s)
	}

	return result
}

func expandTargets(targets []interface{}) []types.Target {
	targetList := make([]types.Target, 0)

	for _, target := range targets {
		if target == nil {
			continue
		}

		t := types.Target{}

		targetData := target.(map[string]interface{})

		if v, ok := targetData["channel_target_info"].([]interface{}); ok {
			t.ChannelTargetInfo = expandChannelTargetInfo(v)
		}

		if v, ok := targetData["contact_target_info"].([]interface{}); ok {
			t.ContactTargetInfo = expandContactTargetInfo(v)
		}

		targetList = append(targetList, t)
	}

	return targetList
}

func flattenTargets(targets []types.Target) []interface{} {
	result := make([]interface{}, 0)

	for _, target := range targets {
		t := map[string]interface{}{}

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

func expandChannelTargetInfo(channelTargetInfo []interface{}) *types.ChannelTargetInfo {
	if len(channelTargetInfo) == 0 {
		return nil
	}

	c := &types.ChannelTargetInfo{}

	channelTargetInfoData := channelTargetInfo[0].(map[string]interface{})

	if v, ok := channelTargetInfoData["contact_channel_id"].(string); ok && v != "" {
		c.ContactChannelId = aws.String(v)
	}

	if v, ok := channelTargetInfoData["retry_interval_in_minutes"].(int); ok {
		c.RetryIntervalInMinutes = aws.Int32(int32(v))
	}

	return c
}

func flattenChannelTargetInfo(channelTargetInfo *types.ChannelTargetInfo) []interface{} {
	var result []interface{}

	c := make(map[string]interface{})

	if v := channelTargetInfo.ContactChannelId; v != nil {
		c["contact_channel_id"] = aws.ToString(v)
	}

	if v := channelTargetInfo.RetryIntervalInMinutes; v != nil {
		c["retry_interval_in_minutes"] = aws.ToInt32(v)
	}

	result = append(result, c)

	return result
}

func expandContactTargetInfo(contactTargetInfo []interface{}) *types.ContactTargetInfo {
	if len(contactTargetInfo) == 0 {
		return nil
	}

	c := &types.ContactTargetInfo{}

	contactTargetInfoData := contactTargetInfo[0].(map[string]interface{})

	if v, ok := contactTargetInfoData["is_essential"].(bool); ok {
		c.IsEssential = aws.Bool(v)
	}

	if v, ok := contactTargetInfoData["contact_id"].(string); ok && v != "" {
		c.ContactId = aws.String(v)
	}

	return c
}

func flattenContactTargetInfo(contactTargetInfo *types.ContactTargetInfo) []interface{} {
	var result []interface{}

	c := make(map[string]interface{})

	if v := contactTargetInfo.IsEssential; v != nil {
		c["is_essential"] = aws.ToBool(v)
	}

	if v := contactTargetInfo.ContactId; v != nil {
		c["contact_id"] = aws.ToString(v)
	}

	result = append(result, c)

	return result
}

func expandHandOffTime(handOffTime string) *types.HandOffTime {
	split := strings.Split(handOffTime, ":")
	hour, _ := strconv.Atoi(split[0])
	minute, _ := strconv.Atoi(split[1])

	return &types.HandOffTime{
		HourOfDay:    int32(hour),
		MinuteOfHour: int32(minute),
	}
}

func flattenHandOffTime(handOffTime *types.HandOffTime) string {
	return fmt.Sprintf("%02d:%02d", handOffTime.HourOfDay, handOffTime.MinuteOfHour)
}

func expandRecurrence(recurrence []interface{}, ctx context.Context) *types.RecurrenceSettings {
	c := &types.RecurrenceSettings{}

	recurrenceSettings := recurrence[0].(map[string]interface{})

	if v, ok := recurrenceSettings["daily_settings"].([]interface{}); ok && v != nil {
		c.DailySettings = expandDailySettings(v)
	}

	if v, ok := recurrenceSettings["monthly_settings"].([]interface{}); ok && v != nil {
		c.MonthlySettings = expandMonthlySettings(v)
	}

	if v, ok := recurrenceSettings["number_of_on_calls"].(int); ok {
		c.NumberOfOnCalls = aws.Int32(int32(v))
	}

	if v, ok := recurrenceSettings["recurrence_multiplier"].(int); ok {
		c.RecurrenceMultiplier = aws.Int32(int32(v))
	}

	if v, ok := recurrenceSettings["shift_coverages"].([]interface{}); ok && v != nil {
		c.ShiftCoverages = expandShiftCoverages(v)
	}

	if v, ok := recurrenceSettings["weekly_settings"].([]interface{}); ok && v != nil {
		c.WeeklySettings = expandWeeklySettings(v)
	}

	return c
}

func flattenRecurrence(recurrence *types.RecurrenceSettings, ctx context.Context) []interface{} {
	var result []interface{}

	c := make(map[string]interface{})

	if v := recurrence.DailySettings; v != nil {
		c["daily_settings"] = flattenDailySettings(v)
	}

	if v := recurrence.MonthlySettings; v != nil {
		c["monthly_settings"] = flattenMonthlySettings(v)
	}

	if v := recurrence.NumberOfOnCalls; v != nil {
		c["number_of_on_calls"] = aws.ToInt32(v)
	}

	if v := recurrence.RecurrenceMultiplier; v != nil {
		c["recurrence_multiplier"] = aws.ToInt32(v)
	}

	if v := recurrence.ShiftCoverages; v != nil {
		c["shift_coverages"] = flattenShiftCoverages(v)
	}

	if v := recurrence.WeeklySettings; v != nil {
		c["weekly_settings"] = flattenWeeklySettings(v)
	}

	result = append(result, c)

	return result
}

func expandDailySettings(dailySettings []interface{}) []types.HandOffTime {
	if len(dailySettings) == 0 {
		return nil
	}

	var result []types.HandOffTime

	for _, dailySetting := range dailySettings {
		result = append(result, *expandHandOffTime(dailySetting.(string)))
	}

	return result
}

func flattenDailySettings(dailySettings []types.HandOffTime) []interface{} {
	if len(dailySettings) == 0 {
		return nil
	}

	var result []interface{}

	for _, handOffTime := range dailySettings {
		result = append(result, flattenHandOffTime(&handOffTime))
	}

	return result
}

func expandMonthlySettings(monthlySettings []interface{}) []types.MonthlySetting {
	if len(monthlySettings) == 0 {
		return nil
	}

	var result []types.MonthlySetting

	for _, monthlySetting := range monthlySettings {
		monthlySettingData := monthlySetting.(map[string]interface{})

		c := types.MonthlySetting{
			DayOfMonth:  aws.Int32(int32(monthlySettingData["day_of_month"].(int))),
			HandOffTime: expandHandOffTime(monthlySettingData["hand_off_time"].(string)),
		}

		result = append(result, c)
	}

	return result
}

func flattenMonthlySettings(monthlySettings []types.MonthlySetting) []interface{} {
	if len(monthlySettings) == 0 {
		return nil
	}

	var result []interface{}

	for _, monthlySetting := range monthlySettings {
		c := make(map[string]interface{})

		if v := monthlySetting.DayOfMonth; v != nil {
			c["day_of_month"] = aws.ToInt32(v)
		}

		if v := monthlySetting.HandOffTime; v != nil {
			c["hand_off_time"] = flattenHandOffTime(v)
		}

		result = append(result, c)
	}

	return result
}

func expandShiftCoverages(shiftCoverages []interface{}) map[string][]types.CoverageTime {
	if len(shiftCoverages) == 0 {
		return nil
	}

	result := make(map[string][]types.CoverageTime)

	for _, shiftCoverage := range shiftCoverages {
		shiftCoverageData := shiftCoverage.(map[string]interface{})

		dayOfWeek := shiftCoverageData["day_of_week"].(string)
		coverageTimes := expandCoverageTimes(shiftCoverageData["coverage_times"].([]interface{}))

		result[dayOfWeek] = coverageTimes
	}

	return result
}

func flattenShiftCoverages(shiftCoverages map[string][]types.CoverageTime) []interface{} {
	if len(shiftCoverages) == 0 {
		return nil
	}

	var result []interface{}

	for coverageDay, coverageTime := range shiftCoverages {
		c := make(map[string]interface{})

		c["coverage_times"] = flattenCoverageTimes(coverageTime)
		c["day_of_week"] = coverageDay

		result = append(result, c)
	}

	// API doesn't return in any consistent order. This causes flakes during testing, so we sort to always return a consistent order
	sortShiftCoverages(result)

	return result
}

func expandCoverageTimes(coverageTimes []interface{}) []types.CoverageTime {
	var result []types.CoverageTime

	for _, coverageTime := range coverageTimes {
		coverageTimeData := coverageTime.(map[string]interface{})

		c := types.CoverageTime{}

		if v, ok := coverageTimeData["end_time"].(string); ok {
			c.End = expandHandOffTime(v)
		}

		if v, ok := coverageTimeData["start_time"].(string); ok {
			c.Start = expandHandOffTime(v)
		}

		result = append(result, c)
	}

	return result
}

func flattenCoverageTimes(coverageTimes []types.CoverageTime) []interface{} {
	var result []interface{}

	c := make(map[string]interface{})

	for _, coverageTime := range coverageTimes {
		if v := coverageTime.End; v != nil {
			c["end_time"] = flattenHandOffTime(v)
		}

		if v := coverageTime.Start; v != nil {
			c["start_time"] = flattenHandOffTime(v)
		}

		result = append(result, c)
	}

	return result
}

func expandWeeklySettings(weeklySettings []interface{}) []types.WeeklySetting {
	var result []types.WeeklySetting

	for _, weeklySetting := range weeklySettings {
		weeklySettingData := weeklySetting.(map[string]interface{})

		c := types.WeeklySetting{
			DayOfWeek:   types.DayOfWeek(weeklySettingData["day_of_week"].(string)),
			HandOffTime: expandHandOffTime(weeklySettingData["hand_off_time"].(string)),
		}

		result = append(result, c)

	}

	return result
}

func flattenWeeklySettings(weeklySettings []types.WeeklySetting) []interface{} {
	if len(weeklySettings) == 0 {
		return nil
	}

	var result []interface{}

	for _, weeklySetting := range weeklySettings {
		c := make(map[string]interface{})

		if v := string(weeklySetting.DayOfWeek); v != "" {
			c["day_of_week"] = v
		}

		if v := weeklySetting.HandOffTime; v != nil {
			c["hand_off_time"] = flattenHandOffTime(v)
		}

		result = append(result, c)
	}

	return result
}
