// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
)

func expandScalingConfiguration(tfMap map[string]interface{}) *rds.ScalingConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &rds.ScalingConfiguration{}

	if v, ok := tfMap["auto_pause"].(bool); ok {
		apiObject.AutoPause = aws.Bool(v)
	}

	if v, ok := tfMap["max_capacity"].(int); ok {
		apiObject.MaxCapacity = aws.Int64(int64(v))
	}

	if v, ok := tfMap["min_capacity"].(int); ok {
		apiObject.MinCapacity = aws.Int64(int64(v))
	}

	if v, ok := tfMap["seconds_until_auto_pause"].(int); ok {
		apiObject.SecondsUntilAutoPause = aws.Int64(int64(v))
	}

	if v, ok := tfMap["timeout_action"].(string); ok && v != "" {
		apiObject.TimeoutAction = aws.String(v)
	}

	return apiObject
}

func flattenManagedMasterUserSecret(apiObject *rds.MasterUserSecret) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if v := apiObject.KmsKeyId; v != nil {
		tfMap["kms_key_id"] = aws.StringValue(v)
	}
	if v := apiObject.SecretArn; v != nil {
		tfMap["secret_arn"] = aws.StringValue(v)
	}
	if v := apiObject.SecretStatus; v != nil {
		tfMap["secret_status"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenScalingConfigurationInfo(apiObject *rds.ScalingConfigurationInfo) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AutoPause; v != nil {
		tfMap["auto_pause"] = aws.BoolValue(v)
	}

	if v := apiObject.MaxCapacity; v != nil {
		tfMap["max_capacity"] = aws.Int64Value(v)
	}

	if v := apiObject.MaxCapacity; v != nil {
		tfMap["max_capacity"] = aws.Int64Value(v)
	}

	if v := apiObject.MinCapacity; v != nil {
		tfMap["min_capacity"] = aws.Int64Value(v)
	}

	if v := apiObject.SecondsUntilAutoPause; v != nil {
		tfMap["seconds_until_auto_pause"] = aws.Int64Value(v)
	}

	if v := apiObject.TimeoutAction; v != nil {
		tfMap["timeout_action"] = aws.StringValue(v)
	}

	return tfMap
}

func expandServerlessV2ScalingConfiguration(tfMap map[string]interface{}) *rds.ServerlessV2ScalingConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &rds.ServerlessV2ScalingConfiguration{}

	if v, ok := tfMap["max_capacity"].(float64); ok && v != 0.0 {
		apiObject.MaxCapacity = aws.Float64(v)
	}

	if v, ok := tfMap["min_capacity"].(float64); ok && v != 0.0 {
		apiObject.MinCapacity = aws.Float64(v)
	}

	return apiObject
}

func flattenServerlessV2ScalingConfigurationInfo(apiObject *rds.ServerlessV2ScalingConfigurationInfo) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.MaxCapacity; v != nil {
		tfMap["max_capacity"] = aws.Float64Value(v)
	}

	if v := apiObject.MinCapacity; v != nil {
		tfMap["min_capacity"] = aws.Float64Value(v)
	}

	return tfMap
}

func expandOptionConfiguration(configured []interface{}) []*rds.OptionConfiguration {
	var option []*rds.OptionConfiguration

	for _, pRaw := range configured {
		data := pRaw.(map[string]interface{})

		o := &rds.OptionConfiguration{
			OptionName: aws.String(data["option_name"].(string)),
		}

		if raw, ok := data["port"]; ok {
			port := raw.(int)
			if port != 0 {
				o.Port = aws.Int64(int64(port))
			}
		}

		if raw, ok := data["db_security_group_memberships"]; ok {
			memberships := flex.ExpandStringSet(raw.(*schema.Set))
			if len(memberships) > 0 {
				o.DBSecurityGroupMemberships = memberships
			}
		}

		if raw, ok := data["vpc_security_group_memberships"]; ok {
			memberships := flex.ExpandStringSet(raw.(*schema.Set))
			if len(memberships) > 0 {
				o.VpcSecurityGroupMemberships = memberships
			}
		}

		if raw, ok := data["option_settings"]; ok {
			o.OptionSettings = expandOptionSetting(raw.(*schema.Set).List())
		}

		if raw, ok := data["version"]; ok && raw.(string) != "" {
			o.OptionVersion = aws.String(raw.(string))
		}

		option = append(option, o)
	}

	return option
}

// Flattens an array of Options into a []map[string]interface{}
func flattenOptions(apiOptions []*rds.Option, optionConfigurations []*rds.OptionConfiguration) []map[string]interface{} {
	result := make([]map[string]interface{}, 0)

	for _, apiOption := range apiOptions {
		if apiOption == nil || apiOption.OptionName == nil {
			continue
		}

		var configuredOption *rds.OptionConfiguration

		for _, optionConfiguration := range optionConfigurations {
			if aws.StringValue(apiOption.OptionName) == aws.StringValue(optionConfiguration.OptionName) {
				configuredOption = optionConfiguration
				break
			}
		}

		dbSecurityGroupMemberships := make([]interface{}, 0)
		for _, db := range apiOption.DBSecurityGroupMemberships {
			if db != nil {
				dbSecurityGroupMemberships = append(dbSecurityGroupMemberships, aws.StringValue(db.DBSecurityGroupName))
			}
		}

		optionSettings := make([]interface{}, 0)
		for _, apiOptionSetting := range apiOption.OptionSettings {
			// The RDS API responds with all settings. Omit settings that match default value,
			// but only if unconfigured. This is to prevent operators from continually needing
			// to continually update their Terraform configurations to match new option settings
			// when added by the API.
			var configuredOptionSetting *rds.OptionSetting

			if configuredOption != nil {
				for _, configuredOptionOptionSetting := range configuredOption.OptionSettings {
					if aws.StringValue(apiOptionSetting.Name) == aws.StringValue(configuredOptionOptionSetting.Name) {
						configuredOptionSetting = configuredOptionOptionSetting
						break
					}
				}
			}

			if configuredOptionSetting == nil && aws.StringValue(apiOptionSetting.Value) == aws.StringValue(apiOptionSetting.DefaultValue) {
				continue
			}

			optionSetting := map[string]interface{}{
				"name":  aws.StringValue(apiOptionSetting.Name),
				"value": aws.StringValue(apiOptionSetting.Value),
			}

			// Some values, like passwords, are sent back from the API as ****.
			// Set the response to match the configuration to prevent an unexpected difference
			if configuredOptionSetting != nil && aws.StringValue(apiOptionSetting.Value) == "****" {
				optionSetting["value"] = aws.StringValue(configuredOptionSetting.Value)
			}

			optionSettings = append(optionSettings, optionSetting)
		}
		optionSettingsResource := &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name": {
					Type:     schema.TypeString,
					Required: true,
				},
				"value": {
					Type:     schema.TypeString,
					Required: true,
				},
			},
		}

		vpcSecurityGroupMemberships := make([]interface{}, 0)
		for _, vpc := range apiOption.VpcSecurityGroupMemberships {
			if vpc != nil {
				vpcSecurityGroupMemberships = append(vpcSecurityGroupMemberships, aws.StringValue(vpc.VpcSecurityGroupId))
			}
		}

		r := map[string]interface{}{
			"db_security_group_memberships":  schema.NewSet(schema.HashString, dbSecurityGroupMemberships),
			"option_name":                    aws.StringValue(apiOption.OptionName),
			"option_settings":                schema.NewSet(schema.HashResource(optionSettingsResource), optionSettings),
			"port":                           aws.Int64Value(apiOption.Port),
			"version":                        aws.StringValue(apiOption.OptionVersion),
			"vpc_security_group_memberships": schema.NewSet(schema.HashString, vpcSecurityGroupMemberships),
		}

		result = append(result, r)
	}

	return result
}

func expandOptionSetting(list []interface{}) []*rds.OptionSetting {
	options := make([]*rds.OptionSetting, 0, len(list))

	for _, oRaw := range list {
		data := oRaw.(map[string]interface{})

		o := &rds.OptionSetting{
			Name:  aws.String(data["name"].(string)),
			Value: aws.String(data["value"].(string)),
		}

		options = append(options, o)
	}

	return options
}

// Takes the result of flatmap.Expand for an array of parameters and
// returns Parameter API compatible objects
func expandParameters(configured []interface{}) []*rds.Parameter {
	var parameters []*rds.Parameter

	// Loop over our configured parameters and create
	// an array of aws-sdk-go compatible objects
	for _, pRaw := range configured {
		data := pRaw.(map[string]interface{})

		if data["name"].(string) == "" {
			continue
		}

		p := &rds.Parameter{
			ParameterName:  aws.String(strings.ToLower(data["name"].(string))),
			ParameterValue: aws.String(data["value"].(string)),
		}

		if data["apply_method"].(string) != "" {
			p.ApplyMethod = aws.String(strings.ToLower(data["apply_method"].(string)))
		}

		parameters = append(parameters, p)
	}

	return parameters
}

// Flattens an array of Parameters into a []map[string]interface{}
func flattenParameters(list []*rds.Parameter) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(list))
	for _, i := range list {
		if i.ParameterName != nil {
			r := make(map[string]interface{})
			if i.ApplyMethod != nil {
				r["apply_method"] = strings.ToLower(aws.StringValue(i.ApplyMethod))
			}

			r["name"] = strings.ToLower(aws.StringValue(i.ParameterName))

			// Default empty string, guard against nil parameter values
			r["value"] = ""
			if i.ParameterValue != nil {
				r["value"] = aws.StringValue(i.ParameterValue)
			}

			result = append(result, r)
		}
	}

	return result
}
