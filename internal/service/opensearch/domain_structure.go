// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package opensearch

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/opensearchservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func expandAdvancedSecurityOptions(m []interface{}) *opensearchservice.AdvancedSecurityOptionsInput_ {
	config := opensearchservice.AdvancedSecurityOptionsInput_{}
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
					muo := opensearchservice.MasterUserOptions{}
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

					config.SetMasterUserOptions(&muo)
				}
			}
		}
	}

	return &config
}

func expandAutoTuneOptions(tfMap map[string]interface{}) *opensearchservice.AutoTuneOptions {
	if tfMap == nil {
		return nil
	}

	options := &opensearchservice.AutoTuneOptions{}

	autoTuneOptionsInput := expandAutoTuneOptionsInput(tfMap)

	options.DesiredState = autoTuneOptionsInput.DesiredState
	options.MaintenanceSchedules = autoTuneOptionsInput.MaintenanceSchedules
	options.UseOffPeakWindow = autoTuneOptionsInput.UseOffPeakWindow

	if v, ok := tfMap["rollback_on_disable"].(string); ok && v != "" {
		options.RollbackOnDisable = aws.String(v)
	}

	return options
}

func expandAutoTuneOptionsInput(tfMap map[string]interface{}) *opensearchservice.AutoTuneOptionsInput_ {
	if tfMap == nil {
		return nil
	}

	options := &opensearchservice.AutoTuneOptionsInput_{}

	options.DesiredState = aws.String(tfMap["desired_state"].(string))

	if v, ok := tfMap["maintenance_schedule"].(*schema.Set); ok && v.Len() > 0 {
		options.MaintenanceSchedules = expandAutoTuneMaintenanceSchedules(v.List())
	}

	options.UseOffPeakWindow = aws.Bool(tfMap["use_off_peak_window"].(bool))

	return options
}

func expandAutoTuneMaintenanceSchedules(tfList []interface{}) []*opensearchservice.AutoTuneMaintenanceSchedule {
	var autoTuneMaintenanceSchedules []*opensearchservice.AutoTuneMaintenanceSchedule

	for _, tfMapRaw := range tfList {
		tfMap, _ := tfMapRaw.(map[string]interface{})

		autoTuneMaintenanceSchedule := &opensearchservice.AutoTuneMaintenanceSchedule{}

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

func expandAutoTuneMaintenanceScheduleDuration(tfMap map[string]interface{}) *opensearchservice.Duration {
	autoTuneMaintenanceScheduleDuration := &opensearchservice.Duration{
		Value: aws.Int64(int64(tfMap[names.AttrValue].(int))),
		Unit:  aws.String(tfMap[names.AttrUnit].(string)),
	}

	return autoTuneMaintenanceScheduleDuration
}

func expandESSAMLOptions(data []interface{}) *opensearchservice.SAMLOptionsInput_ {
	if len(data) == 0 {
		return nil
	}

	if data[0] == nil {
		return &opensearchservice.SAMLOptionsInput_{}
	}

	options := opensearchservice.SAMLOptionsInput_{}
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
				options.SessionTimeoutMinutes = aws.Int64(int64(v))
			}
			if v, ok := group["subject_key"].(string); ok {
				options.SubjectKey = aws.String(v)
			}
		}
	}

	return &options
}

func expandSAMLOptionsIdp(l []interface{}) *opensearchservice.SAMLIdp {
	if len(l) == 0 {
		return nil
	}

	if l[0] == nil {
		return &opensearchservice.SAMLIdp{}
	}

	m := l[0].(map[string]interface{})

	return &opensearchservice.SAMLIdp{
		EntityId:        aws.String(m["entity_id"].(string)),
		MetadataContent: aws.String(m["metadata_content"].(string)),
	}
}

func expandOffPeakWindowOptions(tfMap map[string]interface{}) *opensearchservice.OffPeakWindowOptions {
	if tfMap == nil {
		return nil
	}

	apiObject := &opensearchservice.OffPeakWindowOptions{}

	if v, ok := tfMap[names.AttrEnabled].(bool); ok {
		apiObject.Enabled = aws.Bool(v)
	}

	if v, ok := tfMap["off_peak_window"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.OffPeakWindow = expandOffPeakWindow(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandOffPeakWindow(tfMap map[string]interface{}) *opensearchservice.OffPeakWindow {
	if tfMap == nil {
		return nil
	}

	apiObject := &opensearchservice.OffPeakWindow{}

	if v, ok := tfMap["window_start_time"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.WindowStartTime = expandWindowStartTime(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandWindowStartTime(tfMap map[string]interface{}) *opensearchservice.WindowStartTime {
	if tfMap == nil {
		return nil
	}

	apiObject := &opensearchservice.WindowStartTime{}

	if v, ok := tfMap["hours"].(int); ok {
		apiObject.Hours = aws.Int64(int64(v))
	}

	if v, ok := tfMap["minutes"].(int); ok {
		apiObject.Minutes = aws.Int64(int64(v))
	}

	return apiObject
}

func flattenAdvancedSecurityOptions(advancedSecurityOptions *opensearchservice.AdvancedSecurityOptions) []map[string]interface{} {
	if advancedSecurityOptions == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{}
	m[names.AttrEnabled] = aws.BoolValue(advancedSecurityOptions.Enabled)

	if aws.BoolValue(advancedSecurityOptions.Enabled) && advancedSecurityOptions.AnonymousAuthEnabled != nil {
		m["anonymous_auth_enabled"] = aws.BoolValue(advancedSecurityOptions.AnonymousAuthEnabled)
	}

	if aws.BoolValue(advancedSecurityOptions.Enabled) && advancedSecurityOptions.InternalUserDatabaseEnabled != nil {
		m["internal_user_database_enabled"] = aws.BoolValue(advancedSecurityOptions.InternalUserDatabaseEnabled)
	}

	return []map[string]interface{}{m}
}

func flattenAutoTuneOptions(autoTuneOptions *opensearchservice.AutoTuneOptions) map[string]interface{} {
	if autoTuneOptions == nil {
		return nil
	}

	m := map[string]interface{}{}

	m["desired_state"] = aws.StringValue(autoTuneOptions.DesiredState)

	if v := autoTuneOptions.MaintenanceSchedules; v != nil {
		m["maintenance_schedule"] = flattenAutoTuneMaintenanceSchedules(v)
	}

	m["rollback_on_disable"] = aws.StringValue(autoTuneOptions.RollbackOnDisable)

	m["use_off_peak_window"] = aws.BoolValue(autoTuneOptions.UseOffPeakWindow)

	return m
}

func flattenAutoTuneMaintenanceSchedules(autoTuneMaintenanceSchedules []*opensearchservice.AutoTuneMaintenanceSchedule) []interface{} {
	if len(autoTuneMaintenanceSchedules) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, autoTuneMaintenanceSchedule := range autoTuneMaintenanceSchedules {
		m := map[string]interface{}{}

		m["start_at"] = aws.TimeValue(autoTuneMaintenanceSchedule.StartAt).Format(time.RFC3339)

		m[names.AttrDuration] = []interface{}{flattenAutoTuneMaintenanceScheduleDuration(autoTuneMaintenanceSchedule.Duration)}

		m["cron_expression_for_recurrence"] = aws.StringValue(autoTuneMaintenanceSchedule.CronExpressionForRecurrence)

		tfList = append(tfList, m)
	}

	return tfList
}

func flattenAutoTuneMaintenanceScheduleDuration(autoTuneMaintenanceScheduleDuration *opensearchservice.Duration) map[string]interface{} {
	m := map[string]interface{}{}

	m[names.AttrValue] = aws.Int64Value(autoTuneMaintenanceScheduleDuration.Value)
	m[names.AttrUnit] = aws.StringValue(autoTuneMaintenanceScheduleDuration.Unit)

	return m
}

func flattenESSAMLOptions(d *schema.ResourceData, samlOptions *opensearchservice.SAMLOptionsOutput_) []interface{} {
	if samlOptions == nil {
		return nil
	}

	m := map[string]interface{}{
		names.AttrEnabled: aws.BoolValue(samlOptions.Enabled),
		"idp":             flattenESSAMLIdpOptions(samlOptions.Idp),
	}

	m["roles_key"] = aws.StringValue(samlOptions.RolesKey)
	m["session_timeout_minutes"] = aws.Int64Value(samlOptions.SessionTimeoutMinutes)
	m["subject_key"] = aws.StringValue(samlOptions.SubjectKey)

	// samlOptions.master_backend_role and samlOptions.master_user_name will be added to the
	// all_access role in kibana's security manager.  These values cannot be read or
	// modified by the opensearchservice API.  So, we ignore it on read and let persist
	// the value already in the state.
	m["master_backend_role"] = d.Get("saml_options.0.master_backend_role").(string)
	m["master_user_name"] = d.Get("saml_options.0.master_user_name").(string)

	return []interface{}{m}
}

func flattenESSAMLIdpOptions(SAMLIdp *opensearchservice.SAMLIdp) []interface{} {
	if SAMLIdp == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"entity_id":        aws.StringValue(SAMLIdp.EntityId),
		"metadata_content": aws.StringValue(SAMLIdp.MetadataContent),
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

func expandLogPublishingOptions(m *schema.Set) map[string]*opensearchservice.LogPublishingOption {
	options := make(map[string]*opensearchservice.LogPublishingOption)

	for _, vv := range m.List() {
		lo := vv.(map[string]interface{})
		options[lo["log_type"].(string)] = &opensearchservice.LogPublishingOption{
			CloudWatchLogsLogGroupArn: aws.String(lo[names.AttrCloudWatchLogGroupARN].(string)),
			Enabled:                   aws.Bool(lo[names.AttrEnabled].(bool)),
		}
	}

	return options
}

func flattenLogPublishingOptions(o map[string]*opensearchservice.LogPublishingOption) []map[string]interface{} {
	m := make([]map[string]interface{}, 0)
	for logType, val := range o {
		mm := map[string]interface{}{
			"log_type":        logType,
			names.AttrEnabled: aws.BoolValue(val.Enabled),
		}

		if val.CloudWatchLogsLogGroupArn != nil {
			mm[names.AttrCloudWatchLogGroupARN] = aws.StringValue(val.CloudWatchLogsLogGroupArn)
		}

		m = append(m, mm)
	}
	return m
}

func flattenOffPeakWindowOptions(apiObject *opensearchservice.OffPeakWindowOptions) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Enabled; v != nil {
		tfMap[names.AttrEnabled] = aws.BoolValue(v)
	}

	if v := apiObject.OffPeakWindow; v != nil {
		tfMap["off_peak_window"] = []interface{}{flattenOffPeakWindow(v)}
	}

	return tfMap
}

func flattenOffPeakWindow(apiObject *opensearchservice.OffPeakWindow) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.WindowStartTime; v != nil {
		tfMap["window_start_time"] = []interface{}{flattenWindowStartTime(v)}
	}

	return tfMap
}

func flattenWindowStartTime(apiObject *opensearchservice.WindowStartTime) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Hours; v != nil {
		tfMap["hours"] = aws.Int64Value(v)
	}

	if v := apiObject.Minutes; v != nil {
		tfMap["minutes"] = aws.Int64Value(v)
	}

	return tfMap
}
