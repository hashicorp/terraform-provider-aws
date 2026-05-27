// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package opensearch

import (
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/opensearch/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func expandAdvancedSecurityOptions(m []any) *awstypes.AdvancedSecurityOptionsInput {
	config := awstypes.AdvancedSecurityOptionsInput{}
	group := m[0].(map[string]any)

	if advancedSecurityEnabled, ok := group[names.AttrEnabled]; ok {
		config.Enabled = aws.Bool(advancedSecurityEnabled.(bool))

		if advancedSecurityEnabled.(bool) {
			if v, ok := group["anonymous_auth_enabled"].(bool); ok {
				config.AnonymousAuthEnabled = aws.Bool(v)
			}

			if v, ok := group["internal_user_database_enabled"].(bool); ok {
				config.InternalUserDatabaseEnabled = aws.Bool(v)
			}

			if v, ok := group["jwt_options"].([]any); ok && len(v) > 0 && v[0] != nil {
				config.JWTOptions = expandJWTOptionsInput(v[0].(map[string]any))
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

func expandJWTOptionsInput(tfMap map[string]any) *awstypes.JWTOptionsInput {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.JWTOptionsInput{}

	if v, ok := tfMap[names.AttrEnabled].(bool); ok {
		apiObject.Enabled = aws.Bool(v)
	}

	if v, ok := tfMap[names.AttrPublicKey].(string); ok && v != "" {
		// AWS expects the public key without newlines
		publicKey := strings.ReplaceAll(v, "\n", "")
		apiObject.PublicKey = aws.String(publicKey)
	}

	if v, ok := tfMap["subject_key"].(string); ok && v != "" {
		apiObject.SubjectKey = aws.String(v)
	}

	if v, ok := tfMap["roles_key"].(string); ok && v != "" {
		apiObject.RolesKey = aws.String(v)
	}

	return apiObject
}

func expandAIMLOptionsInput(tfMap map[string]any) *awstypes.AIMLOptionsInput {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.AIMLOptionsInput{}

	if v, ok := tfMap["natural_language_query_generation_options"].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.NaturalLanguageQueryGenerationOptions = expandNaturalLanguageQueryGenerationOptionsInput(v[0].(map[string]any))
	}

	if v, ok := tfMap["s3_vectors_engine"].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.S3VectorsEngine = expandS3VectorsEngine(v[0].(map[string]any))
	}

	if v, ok := tfMap["serverless_vector_acceleration"].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.ServerlessVectorAcceleration = expandServerlessVectorAcceleration(v[0].(map[string]any))
	}

	return apiObject
}

func expandNaturalLanguageQueryGenerationOptionsInput(tfMap map[string]any) *awstypes.NaturalLanguageQueryGenerationOptionsInput {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.NaturalLanguageQueryGenerationOptionsInput{}

	if v, ok := tfMap["desired_state"].(string); ok && v != "" {
		apiObject.DesiredState = awstypes.NaturalLanguageQueryGenerationDesiredState(v)
	}

	return apiObject
}

func expandS3VectorsEngine(tfMap map[string]any) *awstypes.S3VectorsEngine {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.S3VectorsEngine{}

	if v, ok := tfMap[names.AttrEnabled].(bool); ok {
		apiObject.Enabled = aws.Bool(v)
	}

	return apiObject
}

func expandServerlessVectorAcceleration(tfMap map[string]any) *awstypes.ServerlessVectorAcceleration {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.ServerlessVectorAcceleration{}

	if v, ok := tfMap[names.AttrEnabled].(bool); ok {
		apiObject.Enabled = aws.Bool(v)
	}

	return apiObject
}

func expandAutoTuneOptions(tfMap map[string]any) *awstypes.AutoTuneOptions {
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

func expandAutoTuneOptionsInput(tfMap map[string]any) *awstypes.AutoTuneOptionsInput {
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

func expandOffPeakWindowOptions(tfMap map[string]any) *awstypes.OffPeakWindowOptions {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.OffPeakWindowOptions{}

	if v, ok := tfMap[names.AttrEnabled].(bool); ok {
		apiObject.Enabled = aws.Bool(v)
	}

	if v, ok := tfMap["off_peak_window"].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.OffPeakWindow = expandOffPeakWindow(v[0].(map[string]any))
	}

	return apiObject
}

func expandOffPeakWindow(tfMap map[string]any) *awstypes.OffPeakWindow {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.OffPeakWindow{}

	if v, ok := tfMap["window_start_time"].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.WindowStartTime = expandWindowStartTime(v[0].(map[string]any))
	}

	return apiObject
}

func expandWindowStartTime(tfMap map[string]any) *awstypes.WindowStartTime {
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

func flattenAdvancedSecurityOptions(advancedSecurityOptions *awstypes.AdvancedSecurityOptions) []map[string]any {
	if advancedSecurityOptions == nil {
		return []map[string]any{}
	}

	m := map[string]any{}
	m[names.AttrEnabled] = aws.ToBool(advancedSecurityOptions.Enabled)

	if aws.ToBool(advancedSecurityOptions.Enabled) && advancedSecurityOptions.AnonymousAuthEnabled != nil {
		m["anonymous_auth_enabled"] = aws.ToBool(advancedSecurityOptions.AnonymousAuthEnabled)
	}

	if aws.ToBool(advancedSecurityOptions.Enabled) && advancedSecurityOptions.InternalUserDatabaseEnabled != nil {
		m["internal_user_database_enabled"] = aws.ToBool(advancedSecurityOptions.InternalUserDatabaseEnabled)
	}

	if advancedSecurityOptions.JWTOptions != nil && aws.ToBool(advancedSecurityOptions.JWTOptions.Enabled) {
		m["jwt_options"] = flattenJWTOptionsOutput(advancedSecurityOptions.JWTOptions)
	}

	return []map[string]any{m}
}

func flattenJWTOptionsOutput(apiObject *awstypes.JWTOptionsOutput) []map[string]any {
	if apiObject == nil {
		return nil
	}

	m := map[string]any{}

	if apiObject.Enabled != nil {
		m[names.AttrEnabled] = aws.ToBool(apiObject.Enabled)
	}

	if apiObject.PublicKey != nil {
		m[names.AttrPublicKey] = aws.ToString(apiObject.PublicKey)
	}

	if apiObject.SubjectKey != nil {
		m["subject_key"] = aws.ToString(apiObject.SubjectKey)
	}

	if apiObject.RolesKey != nil {
		m["roles_key"] = aws.ToString(apiObject.RolesKey)
	}

	return []map[string]any{m}
}

func flattenAIMLOptionsOutput(apiObject *awstypes.AIMLOptionsOutput) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.NaturalLanguageQueryGenerationOptions; v != nil {
		tfMap["natural_language_query_generation_options"] = []any{flattenNaturalLanguageQueryGenerationOptionsOutput(v)}
	}

	if v := apiObject.S3VectorsEngine; v != nil {
		tfMap["s3_vectors_engine"] = []any{flattenS3VectorsEngine(v)}
	}

	if v := apiObject.ServerlessVectorAcceleration; v != nil {
		tfMap["serverless_vector_acceleration"] = []any{flattenServerlessVectorAcceleration(v)}
	}

	return tfMap
}

func flattenNaturalLanguageQueryGenerationOptionsOutput(apiObject *awstypes.NaturalLanguageQueryGenerationOptionsOutput) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		"desired_state": apiObject.DesiredState,
	}

	return tfMap
}

func flattenS3VectorsEngine(apiObject *awstypes.S3VectorsEngine) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		names.AttrEnabled: aws.ToBool(apiObject.Enabled),
	}

	return tfMap
}

func flattenServerlessVectorAcceleration(apiObject *awstypes.ServerlessVectorAcceleration) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		names.AttrEnabled: aws.ToBool(apiObject.Enabled),
	}

	return tfMap
}

func flattenAutoTuneOptions(autoTuneOptions *awstypes.AutoTuneOptions) map[string]any {
	if autoTuneOptions == nil {
		return nil
	}

	m := map[string]any{}

	m["desired_state"] = autoTuneOptions.DesiredState

	if v := autoTuneOptions.MaintenanceSchedules; v != nil {
		m["maintenance_schedule"] = flattenAutoTuneMaintenanceSchedules(v)
	}

	m["rollback_on_disable"] = autoTuneOptions.RollbackOnDisable

	m["use_off_peak_window"] = aws.ToBool(autoTuneOptions.UseOffPeakWindow)

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
	m[names.AttrUnit] = autoTuneMaintenanceScheduleDuration.Unit

	return m
}

func flattenESSAMLOptions(d *schema.ResourceData, samlOptions *awstypes.SAMLOptionsOutput) []any {
	if samlOptions == nil {
		return nil
	}

	m := map[string]any{
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

	return []any{m}
}

func flattenESSAMLIdpOptions(SAMLIdp *awstypes.SAMLIdp) []any {
	if SAMLIdp == nil {
		return []any{}
	}

	m := map[string]any{
		"entity_id":        aws.ToString(SAMLIdp.EntityId),
		"metadata_content": aws.ToString(SAMLIdp.MetadataContent),
	}

	return []any{m}
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

func expandIdentityCenterOptions(tfList []any) *awstypes.IdentityCenterOptionsInput {
	if len(tfList) == 0 {
		return nil
	}

	if tfList[0] == nil {
		return &awstypes.IdentityCenterOptionsInput{}
	}

	apiObject := &awstypes.IdentityCenterOptionsInput{}
	tfMap := tfList[0].(map[string]any)

	if v, ok := tfMap["enabled_api_access"].(bool); ok {
		apiObject.EnabledAPIAccess = aws.Bool(v)
	}

	if apiObject.EnabledAPIAccess != nil && aws.ToBool(apiObject.EnabledAPIAccess) {
		if v, ok := tfMap["identity_center_instance_arn"].(string); ok && v != "" {
			apiObject.IdentityCenterInstanceARN = aws.String(v)
		}

		if v, ok := tfMap["roles_key"].(string); ok && v != "" {
			apiObject.RolesKey = awstypes.RolesKeyIdCOption(v)
		}

		if v, ok := tfMap["subject_key"].(string); ok && v != "" {
			apiObject.SubjectKey = awstypes.SubjectKeyIdCOption(v)
		}
	}

	return apiObject
}

func flattenIdentityCenterOptions(apiObject *awstypes.IdentityCenterOptions) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.EnabledAPIAccess; v != nil {
		tfMap["enabled_api_access"] = aws.ToBool(v)
	}

	if v := apiObject.IdentityCenterInstanceARN; v != nil {
		tfMap["identity_center_instance_arn"] = aws.ToString(v)
	}

	if v := apiObject.RolesKey; v != "" {
		tfMap["roles_key"] = v
	}

	if v := apiObject.SubjectKey; v != "" {
		tfMap["subject_key"] = v
	}

	return []any{tfMap}
}

func expandLogPublishingOptions(tfList []any) map[string]awstypes.LogPublishingOption {
	apiObjects := make(map[string]awstypes.LogPublishingOption)

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := awstypes.LogPublishingOption{
			Enabled: aws.Bool(tfMap[names.AttrEnabled].(bool)),
		}

		if v, ok := tfMap[names.AttrCloudWatchLogGroupARN].(string); ok && v != "" {
			apiObject.CloudWatchLogsLogGroupArn = aws.String(v)
		}

		apiObjects[tfMap["log_type"].(string)] = apiObject
	}

	return apiObjects
}

func flattenLogPublishingOptions(apiObjects map[string]awstypes.LogPublishingOption) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	tfList := make([]any, 0)

	for k, apiObject := range apiObjects {
		tfMap := map[string]any{
			names.AttrEnabled: aws.ToBool(apiObject.Enabled),
			"log_type":        k,
		}

		if v := apiObject.CloudWatchLogsLogGroupArn; v != nil {
			tfMap[names.AttrCloudWatchLogGroupARN] = aws.ToString(v)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenOffPeakWindowOptions(apiObject *awstypes.OffPeakWindowOptions) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.Enabled; v != nil {
		tfMap[names.AttrEnabled] = aws.ToBool(v)
	}

	if v := apiObject.OffPeakWindow; v != nil {
		tfMap["off_peak_window"] = []any{flattenOffPeakWindow(v)}
	}

	return tfMap
}

func flattenOffPeakWindow(apiObject *awstypes.OffPeakWindow) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.WindowStartTime; v != nil {
		tfMap["window_start_time"] = []any{flattenWindowStartTime(v)}
	}

	return tfMap
}

func flattenWindowStartTime(apiObject *awstypes.WindowStartTime) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		"hours":   apiObject.Hours,
		"minutes": apiObject.Minutes,
	}

	return tfMap
}
