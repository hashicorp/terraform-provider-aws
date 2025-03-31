// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticsearch

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticsearchservice/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func expandAdvancedSecurityOptions(m []any) *awstypes.AdvancedSecurityOptionsInput {
	config := awstypes.AdvancedSecurityOptionsInput{}
	group := m[0].(map[string]any)

	if advancedSecurityEnabled, ok := group[names.AttrEnabled]; ok {
		config.Enabled = aws.Bool(advancedSecurityEnabled.(bool))

		if advancedSecurityEnabled.(bool) {
			if v, ok := group["internal_user_database_enabled"].(bool); ok {
				config.InternalUserDatabaseEnabled = aws.Bool(v)
			}

			if v, ok := group["master_user_options"].([]any); ok {
				if len(v) > 0 && v[0] != nil {
					muo := awstypes.MasterUserOptions{}
					masterUserOptions := v[0].(map[string]any)

					if v, ok := masterUserOptions["master_user_arn"].(string); ok && v != "" {
						muo.MasterUserARN = aws.String(v)
					}

					if v, ok := masterUserOptions["master_user_name"].(string); ok && v != "" {
						muo.MasterUserName = aws.String(v)
					}

					if v, ok := masterUserOptions["master_user_password"].(string); ok && v != "" {
						muo.MasterUserPassword = aws.String(v)
					}

					config.MasterUserOptions = &muo
				}
			}
		}
	}

	return &config
}

func expandAutoTuneOptions(tfMap map[string]any) *awstypes.AutoTuneOptions {
	if tfMap == nil {
		return nil
	}

	options := &awstypes.AutoTuneOptions{}

	autoTuneOptionsInput := expandAutoTuneOptionsInput(tfMap)

	options.DesiredState = autoTuneOptionsInput.DesiredState
	options.MaintenanceSchedules = autoTuneOptionsInput.MaintenanceSchedules

	if v, ok := tfMap["rollback_on_disable"].(string); ok && v != "" {
		options.RollbackOnDisable = awstypes.RollbackOnDisable(v)
	}

	return options
}

func expandAutoTuneOptionsInput(tfMap map[string]any) *awstypes.AutoTuneOptionsInput {
	if tfMap == nil {
		return nil
	}

	options := &awstypes.AutoTuneOptionsInput{}

	options.DesiredState = awstypes.AutoTuneDesiredState(tfMap["desired_state"].(string))

	if v, ok := tfMap["maintenance_schedule"].(*schema.Set); ok && v.Len() > 0 {
		options.MaintenanceSchedules = expandAutoTuneMaintenanceSchedules(v.List())
	}

	return options
}

func expandAutoTuneMaintenanceSchedules(tfList []any) []awstypes.AutoTuneMaintenanceSchedule {
	var autoTuneMaintenanceSchedules []awstypes.AutoTuneMaintenanceSchedule

	for _, tfMapRaw := range tfList {
		tfMap, _ := tfMapRaw.(map[string]any)

		autoTuneMaintenanceSchedule := awstypes.AutoTuneMaintenanceSchedule{}

		startAt, _ := time.Parse(time.RFC3339, tfMap["start_at"].(string))
		autoTuneMaintenanceSchedule.StartAt = aws.Time(startAt)

		if v, ok := tfMap[names.AttrDuration].([]any); ok {
			autoTuneMaintenanceSchedule.Duration = expandAutoTuneMaintenanceScheduleDuration(v[0].(map[string]any))
		}

		autoTuneMaintenanceSchedule.CronExpressionForRecurrence = aws.String(tfMap["cron_expression_for_recurrence"].(string))

		autoTuneMaintenanceSchedules = append(autoTuneMaintenanceSchedules, autoTuneMaintenanceSchedule)
	}

	return autoTuneMaintenanceSchedules
}

func expandAutoTuneMaintenanceScheduleDuration(tfMap map[string]any) *awstypes.Duration {
	autoTuneMaintenanceScheduleDuration := &awstypes.Duration{
		Value: aws.Int64(int64(tfMap[names.AttrValue].(int))),
		Unit:  awstypes.TimeUnit(tfMap[names.AttrUnit].(string)),
	}

	return autoTuneMaintenanceScheduleDuration
}

func expandESSAMLOptions(data []any) *awstypes.SAMLOptionsInput {
	if len(data) == 0 {
		return nil
	}

	if data[0] == nil {
		return &awstypes.SAMLOptionsInput{}
	}

	options := awstypes.SAMLOptionsInput{}
	group := data[0].(map[string]any)

	if SAMLEnabled, ok := group[names.AttrEnabled]; ok {
		options.Enabled = aws.Bool(SAMLEnabled.(bool))

		if SAMLEnabled.(bool) {
			options.Idp = expandSAMLOptionsIdp(group["idp"].([]any))
			if v, ok := group["master_backend_role"].(string); ok && v != "" {
				options.MasterBackendRole = aws.String(v)
			}
			if v, ok := group["master_user_name"].(string); ok && v != "" {
				options.MasterUserName = aws.String(v)
			}
			if v, ok := group["roles_key"].(string); ok {
				options.RolesKey = aws.String(v)
			}
			if v, ok := group["session_timeout_minutes"].(int); ok {
				options.SessionTimeoutMinutes = aws.Int32(int32(v))
			}
			if v, ok := group["subject_key"].(string); ok {
				options.SubjectKey = aws.String(v)
			}
		}
	}

	return &options
}

func expandSAMLOptionsIdp(l []any) *awstypes.SAMLIdp {
	if len(l) == 0 {
		return nil
	}

	if l[0] == nil {
		return &awstypes.SAMLIdp{}
	}

	m := l[0].(map[string]any)

	return &awstypes.SAMLIdp{
		EntityId:        aws.String(m["entity_id"].(string)),
		MetadataContent: aws.String(m["metadata_content"].(string)),
	}
}

func flattenAdvancedSecurityOptions(advancedSecurityOptions *awstypes.AdvancedSecurityOptions) []map[string]any {
	if advancedSecurityOptions == nil {
		return []map[string]any{}
	}

	m := map[string]any{}
	m[names.AttrEnabled] = aws.ToBool(advancedSecurityOptions.Enabled)
	if aws.ToBool(advancedSecurityOptions.Enabled) {
		m["internal_user_database_enabled"] = aws.ToBool(advancedSecurityOptions.InternalUserDatabaseEnabled)
	}

	return []map[string]any{m}
}

func flattenAutoTuneOptions(autoTuneOptions *awstypes.AutoTuneOptions) map[string]any {
	if autoTuneOptions == nil {
		return nil
	}

	m := map[string]any{}

	m["desired_state"] = string(autoTuneOptions.DesiredState)

	if v := autoTuneOptions.MaintenanceSchedules; v != nil {
		m["maintenance_schedule"] = flattenAutoTuneMaintenanceSchedules(v)
	}

	m["rollback_on_disable"] = string(autoTuneOptions.RollbackOnDisable)

	return m
}

func flattenAutoTuneMaintenanceSchedules(autoTuneMaintenanceSchedules []awstypes.AutoTuneMaintenanceSchedule) []any {
	if len(autoTuneMaintenanceSchedules) == 0 {
		return nil
	}

	var tfList []any

	for _, autoTuneMaintenanceSchedule := range autoTuneMaintenanceSchedules {
		m := map[string]any{}

		m["start_at"] = aws.ToTime(autoTuneMaintenanceSchedule.StartAt).Format(time.RFC3339)

		m[names.AttrDuration] = []any{flattenAutoTuneMaintenanceScheduleDuration(autoTuneMaintenanceSchedule.Duration)}

		m["cron_expression_for_recurrence"] = aws.ToString(autoTuneMaintenanceSchedule.CronExpressionForRecurrence)

		tfList = append(tfList, m)
	}

	return tfList
}

func flattenAutoTuneMaintenanceScheduleDuration(autoTuneMaintenanceScheduleDuration *awstypes.Duration) map[string]any {
	m := map[string]any{}

	m[names.AttrValue] = aws.ToInt64(autoTuneMaintenanceScheduleDuration.Value)
	m[names.AttrUnit] = string(autoTuneMaintenanceScheduleDuration.Unit)

	return m
}

func flattenSAMLOptionsOutput(d *schema.ResourceData, apiObject *awstypes.SAMLOptionsOutput) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		names.AttrEnabled: aws.ToBool(apiObject.Enabled),
		"idp":             flattenSAMLIdp(apiObject.Idp),
	}

	tfMap["roles_key"] = aws.ToString(apiObject.RolesKey)
	tfMap["session_timeout_minutes"] = aws.ToInt32(apiObject.SessionTimeoutMinutes)
	tfMap["subject_key"] = aws.ToString(apiObject.SubjectKey)

	// samlOptions.master_backend_role and samlOptions.master_user_name will be added to the
	// all_access role in kibana's security manager.  These values cannot be read or
	// modified by the elasticsearch API.  So, we ignore it on read and let persist
	// the value already in the state.
	tfMap["master_backend_role"] = d.Get("saml_options.0.master_backend_role").(string)
	tfMap["master_user_name"] = d.Get("saml_options.0.master_user_name").(string)

	return []any{tfMap}
}

func flattenSAMLIdp(apiObject *awstypes.SAMLIdp) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		"entity_id":        aws.ToString(apiObject.EntityId),
		"metadata_content": aws.ToString(apiObject.MetadataContent),
	}

	return []any{tfMap}
}

func getMasterUserOptions(d *schema.ResourceData) []any {
	if v, ok := d.GetOk("advanced_security_options"); ok {
		options := v.([]any)
		if len(options) > 0 && options[0] != nil {
			m := options[0].(map[string]any)
			if opts, ok := m["master_user_options"]; ok {
				return opts.([]any)
			}
		}
	}
	return []any{}
}

func getUserDBEnabled(d *schema.ResourceData) bool {
	if v, ok := d.GetOk("advanced_security_options"); ok {
		options := v.([]any)
		if len(options) > 0 && options[0] != nil {
			m := options[0].(map[string]any)
			if enabled, ok := m["internal_user_database_enabled"]; ok {
				return enabled.(bool)
			}
		}
	}
	return false
}

func expandLogPublishingOptions(m *schema.Set) map[string]awstypes.LogPublishingOption {
	options := make(map[string]awstypes.LogPublishingOption)

	for _, vv := range m.List() {
		lo := vv.(map[string]any)
		options[lo["log_type"].(string)] = awstypes.LogPublishingOption{
			CloudWatchLogsLogGroupArn: aws.String(lo[names.AttrCloudWatchLogGroupARN].(string)),
			Enabled:                   aws.Bool(lo[names.AttrEnabled].(bool)),
		}
	}

	return options
}

func flattenLogPublishingOptions(o map[string]awstypes.LogPublishingOption) []map[string]any {
	m := make([]map[string]any, 0)
	for logType, val := range o {
		mm := map[string]any{
			"log_type":        logType,
			names.AttrEnabled: aws.ToBool(val.Enabled),
		}

		if val.CloudWatchLogsLogGroupArn != nil {
			mm[names.AttrCloudWatchLogGroupARN] = aws.ToString(val.CloudWatchLogsLogGroupArn)
		}

		m = append(m, mm)
	}
	return m
}
