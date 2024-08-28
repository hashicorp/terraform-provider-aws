// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package opensearch

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/opensearch/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func expandAdvancedSecurityOptions(m []interface{}) *awstypes.AdvancedSecurityOptionsInput {
	config := awstypes.AdvancedSecurityOptionsInput{}
	group := m[0].(map[string]interface{})

	if advancedSecurityEnabled, ok := group[names.AttrEnabled]; ok {
		config.Enabled = aws.Bool(advancedSecurityEnabled.(bool))

		if advancedSecurityEnabled.(bool) {
			if v, ok := group["anonymous_auth_enabled"].(bool); ok {
				config.AnonymousAuthEnabled = aws.Bool(v)
			}

			if v, ok := group["internal_user_database_enabled"].(bool); ok {
				config.InternalUserDatabaseEnabled = aws.Bool(v)
			}

			if v, ok := group["master_user_options"].([]interface{}); ok {
				if len(v) > 0 && v[0] != nil {
					muo := awstypes.MasterUserOptions{}
					masterUserOptions := v[0].(map[string]interface{})

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

func expandAutoTuneOptions(tfMap map[string]interface{}) *awstypes.AutoTuneOptions {
	if tfMap == nil {
		return nil
	}

	options := &awstypes.AutoTuneOptions{}

	autoTuneOptionsInput := expandAutoTuneOptionsInput(tfMap)

	options.DesiredState = autoTuneOptionsInput.DesiredState
	options.MaintenanceSchedules = autoTuneOptionsInput.MaintenanceSchedules
	options.UseOffPeakWindow = autoTuneOptionsInput.UseOffPeakWindow

	if v, ok := tfMap["rollback_on_disable"].(string); ok && v != "" {
		options.RollbackOnDisable = awstypes.RollbackOnDisable(v)
	}

	return options
}

func expandAutoTuneOptionsInput(tfMap map[string]interface{}) *awstypes.AutoTuneOptionsInput {
	if tfMap == nil {
		return nil
	}

	options := &awstypes.AutoTuneOptionsInput{}

	options.DesiredState = awstypes.AutoTuneDesiredState(tfMap["desired_state"].(string))

	if v, ok := tfMap["maintenance_schedule"].(*schema.Set); ok && v.Len() > 0 {
		options.MaintenanceSchedules = expandAutoTuneMaintenanceSchedules(v.List())
	}

	options.UseOffPeakWindow = aws.Bool(tfMap["use_off_peak_window"].(bool))

	return options
}

func expandAutoTuneMaintenanceSchedules(tfList []interface{}) []awstypes.AutoTuneMaintenanceSchedule {
	var autoTuneMaintenanceSchedules []awstypes.AutoTuneMaintenanceSchedule

	for _, tfMapRaw := range tfList {
		tfMap, _ := tfMapRaw.(map[string]interface{})

		autoTuneMaintenanceSchedule := awstypes.AutoTuneMaintenanceSchedule{}

		startAt, _ := time.Parse(time.RFC3339, tfMap["start_at"].(string))
		autoTuneMaintenanceSchedule.StartAt = aws.Time(startAt)

		if v, ok := tfMap[names.AttrDuration].([]interface{}); ok {
			autoTuneMaintenanceSchedule.Duration = expandAutoTuneMaintenanceScheduleDuration(v[0].(map[string]interface{}))
		}

		autoTuneMaintenanceSchedule.CronExpressionForRecurrence = aws.String(tfMap["cron_expression_for_recurrence"].(string))

		autoTuneMaintenanceSchedules = append(autoTuneMaintenanceSchedules, autoTuneMaintenanceSchedule)
	}

	return autoTuneMaintenanceSchedules
}

func expandAutoTuneMaintenanceScheduleDuration(tfMap map[string]interface{}) *awstypes.Duration {
	autoTuneMaintenanceScheduleDuration := &awstypes.Duration{
		Value: aws.Int64(int64(tfMap[names.AttrValue].(int))),
		Unit:  awstypes.TimeUnit(tfMap[names.AttrUnit].(string)),
	}

	return autoTuneMaintenanceScheduleDuration
}

func expandESSAMLOptions(data []interface{}) *awstypes.SAMLOptionsInput {
	if len(data) == 0 {
		return nil
	}

	if data[0] == nil {
		return &awstypes.SAMLOptionsInput{}
	}

	options := awstypes.SAMLOptionsInput{}
	group := data[0].(map[string]interface{})

	if SAMLEnabled, ok := group[names.AttrEnabled]; ok {
		options.Enabled = aws.Bool(SAMLEnabled.(bool))

		if SAMLEnabled.(bool) {
			options.Idp = expandSAMLOptionsIdp(group["idp"].([]interface{}))
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

func expandSAMLOptionsIdp(l []interface{}) *awstypes.SAMLIdp {
	if len(l) == 0 {
		return nil
	}

	if l[0] == nil {
		return &awstypes.SAMLIdp{}
	}

	m := l[0].(map[string]interface{})

	return &awstypes.SAMLIdp{
		EntityId:        aws.String(m["entity_id"].(string)),
		MetadataContent: aws.String(m["metadata_content"].(string)),
	}
}

func expandOffPeakWindowOptions(tfMap map[string]interface{}) *awstypes.OffPeakWindowOptions {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.OffPeakWindowOptions{}

	if v, ok := tfMap[names.AttrEnabled].(bool); ok {
		apiObject.Enabled = aws.Bool(v)
	}

	if v, ok := tfMap["off_peak_window"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.OffPeakWindow = expandOffPeakWindow(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandOffPeakWindow(tfMap map[string]interface{}) *awstypes.OffPeakWindow {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.OffPeakWindow{}

	if v, ok := tfMap["window_start_time"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.WindowStartTime = expandWindowStartTime(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandWindowStartTime(tfMap map[string]interface{}) *awstypes.WindowStartTime {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.WindowStartTime{}

	if v, ok := tfMap["hours"].(int); ok {
		apiObject.Hours = int64(v)
	}

	if v, ok := tfMap["minutes"].(int); ok {
		apiObject.Minutes = int64(v)
	}

	return apiObject
}

func flattenAdvancedSecurityOptions(advancedSecurityOptions *awstypes.AdvancedSecurityOptions) []map[string]interface{} {
	if advancedSecurityOptions == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{}
	m[names.AttrEnabled] = aws.ToBool(advancedSecurityOptions.Enabled)

	if aws.ToBool(advancedSecurityOptions.Enabled) && advancedSecurityOptions.AnonymousAuthEnabled != nil {
		m["anonymous_auth_enabled"] = aws.ToBool(advancedSecurityOptions.AnonymousAuthEnabled)
	}

	if aws.ToBool(advancedSecurityOptions.Enabled) && advancedSecurityOptions.InternalUserDatabaseEnabled != nil {
		m["internal_user_database_enabled"] = aws.ToBool(advancedSecurityOptions.InternalUserDatabaseEnabled)
	}

	return []map[string]interface{}{m}
}

func flattenAutoTuneOptions(autoTuneOptions *awstypes.AutoTuneOptions) map[string]interface{} {
	if autoTuneOptions == nil {
		return nil
	}

	m := map[string]interface{}{}

	m["desired_state"] = autoTuneOptions.DesiredState

	if v := autoTuneOptions.MaintenanceSchedules; v != nil {
		m["maintenance_schedule"] = flattenAutoTuneMaintenanceSchedules(v)
	}

	m["rollback_on_disable"] = autoTuneOptions.RollbackOnDisable

	m["use_off_peak_window"] = aws.ToBool(autoTuneOptions.UseOffPeakWindow)

	return m
}

func flattenAutoTuneMaintenanceSchedules(autoTuneMaintenanceSchedules []awstypes.AutoTuneMaintenanceSchedule) []interface{} {
	if len(autoTuneMaintenanceSchedules) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, autoTuneMaintenanceSchedule := range autoTuneMaintenanceSchedules {
		m := map[string]interface{}{}

		m["start_at"] = aws.ToTime(autoTuneMaintenanceSchedule.StartAt).Format(time.RFC3339)

		m[names.AttrDuration] = []interface{}{flattenAutoTuneMaintenanceScheduleDuration(autoTuneMaintenanceSchedule.Duration)}

		m["cron_expression_for_recurrence"] = aws.ToString(autoTuneMaintenanceSchedule.CronExpressionForRecurrence)

		tfList = append(tfList, m)
	}

	return tfList
}

func flattenAutoTuneMaintenanceScheduleDuration(autoTuneMaintenanceScheduleDuration *awstypes.Duration) map[string]interface{} {
	m := map[string]interface{}{}

	m[names.AttrValue] = aws.ToInt64(autoTuneMaintenanceScheduleDuration.Value)
	m[names.AttrUnit] = autoTuneMaintenanceScheduleDuration.Unit

	return m
}

func flattenESSAMLOptions(d *schema.ResourceData, samlOptions *awstypes.SAMLOptionsOutput) []interface{} {
	if samlOptions == nil {
		return nil
	}

	m := map[string]interface{}{
		names.AttrEnabled: aws.ToBool(samlOptions.Enabled),
		"idp":             flattenESSAMLIdpOptions(samlOptions.Idp),
	}

	m["roles_key"] = aws.ToString(samlOptions.RolesKey)
	m["session_timeout_minutes"] = aws.ToInt32(samlOptions.SessionTimeoutMinutes)
	m["subject_key"] = aws.ToString(samlOptions.SubjectKey)

	// samlOptions.master_backend_role and samlOptions.master_user_name will be added to the
	// all_access role in kibana's security manager.  These values cannot be read or
	// modified by the opensearch API.  So, we ignore it on read and let persist
	// the value already in the state.
	m["master_backend_role"] = d.Get("saml_options.0.master_backend_role").(string)
	m["master_user_name"] = d.Get("saml_options.0.master_user_name").(string)

	return []interface{}{m}
}

func flattenESSAMLIdpOptions(SAMLIdp *awstypes.SAMLIdp) []interface{} {
	if SAMLIdp == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"entity_id":        aws.ToString(SAMLIdp.EntityId),
		"metadata_content": aws.ToString(SAMLIdp.MetadataContent),
	}

	return []interface{}{m}
}

func getMasterUserOptions(d *schema.ResourceData) []interface{} {
	if v, ok := d.GetOk("advanced_security_options"); ok {
		options := v.([]interface{})
		if len(options) > 0 && options[0] != nil {
			m := options[0].(map[string]interface{})
			if opts, ok := m["master_user_options"]; ok {
				return opts.([]interface{})
			}
		}
	}
	return []interface{}{}
}

func expandLogPublishingOptions(m *schema.Set) map[string]awstypes.LogPublishingOption {
	options := make(map[string]awstypes.LogPublishingOption)

	for _, vv := range m.List() {
		lo := vv.(map[string]interface{})
		options[lo["log_type"].(string)] = awstypes.LogPublishingOption{
			CloudWatchLogsLogGroupArn: aws.String(lo[names.AttrCloudWatchLogGroupARN].(string)),
			Enabled:                   aws.Bool(lo[names.AttrEnabled].(bool)),
		}
	}

	return options
}

func flattenLogPublishingOptions(o map[string]awstypes.LogPublishingOption) []map[string]interface{} {
	m := make([]map[string]interface{}, 0)
	for logType, val := range o {
		mm := map[string]interface{}{
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

func flattenOffPeakWindowOptions(apiObject *awstypes.OffPeakWindowOptions) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Enabled; v != nil {
		tfMap[names.AttrEnabled] = aws.ToBool(v)
	}

	if v := apiObject.OffPeakWindow; v != nil {
		tfMap["off_peak_window"] = []interface{}{flattenOffPeakWindow(v)}
	}

	return tfMap
}

func flattenOffPeakWindow(apiObject *awstypes.OffPeakWindow) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.WindowStartTime; v != nil {
		tfMap["window_start_time"] = []interface{}{flattenWindowStartTime(v)}
	}

	return tfMap
}

func flattenWindowStartTime(apiObject *awstypes.WindowStartTime) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"hours":   apiObject.Hours,
		"minutes": apiObject.Minutes,
	}

	return tfMap
}
