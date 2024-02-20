// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securityhub

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"math"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/securityhub"
	"github.com/aws/aws-sdk-go-v2/service/securityhub/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_securityhub_configuration_policy")
func ResourceConfigurationPolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceConfigurationPolicyCreate,
		ReadWithoutTimeout:   resourceConfigurationPolicyRead,
		UpdateWithoutTimeout: resourceConfigurationPolicyUpdate,
		DeleteWithoutTimeout: resourceConfigurationPolicyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringMatch(
					regexache.MustCompile(`[A-Za-z0-9\-\.!*/]+`),
					"Only alphanumeric characters and the following ASCII characters are permitted: -, ., !, *, /",
				),
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"security_hub_policy": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"service_enabled": {
							Type:     schema.TypeBool,
							Required: true,
						},
						"enabled_standard_arns": {
							Type:     schema.TypeList,
							Required: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: verify.ValidARN,
							},
						},
						"security_controls_configuration": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"disabled_control_identifiers": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: validation.StringIsNotEmpty,
										},
										ConflictsWith: []string{
											"security_hub_policy.0.security_controls_configuration.0.enabled_control_identifiers",
										},
									},
									"enabled_control_identifiers": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: validation.StringIsNotEmpty,
										},
										ConflictsWith: []string{
											"security_hub_policy.0.security_controls_configuration.0.disabled_control_identifiers",
										},
									},
									"control_custom_parameter": {
										Type:        schema.TypeList,
										Optional:    true,
										Description: "https://docs.aws.amazon.com/securityhub/latest/userguide/securityhub-controls-reference.html",
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"control_identifier": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringIsNotEmpty,
												},
												"parameter": {
													Type:     schema.TypeSet,
													Required: true,
													MinItems: 1,
													Elem:     customParameterResource(),
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func customParameterResource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"value_type": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: enum.Validate[types.ParameterValueType](),
			},
			"bool": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"value": {
							Required: true,
							Type:     schema.TypeBool,
						},
					},
				},
			},
			"double": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"value": {
							Required: true,
							Type:     schema.TypeFloat,
						},
					},
				},
			},
			"enum": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"value": {
							Required: true,
							Type:     schema.TypeString,
						},
					},
				},
			},
			"enum_list": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"value": {
							Required: true,
							Type:     schema.TypeList,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
			"int": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"value": {
							Required:     true,
							Type:         schema.TypeInt,
							ValidateFunc: validation.IntAtMost(math.MaxInt32),
						},
					},
				},
			},
			"int_list": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"value": {
							Required: true,
							Type:     schema.TypeList,
							Elem: &schema.Schema{
								Type:         schema.TypeInt,
								ValidateFunc: validation.IntAtMost(math.MaxInt32),
							},
						},
					},
				},
			},
			"string": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"value": {
							Required: true,
							Type:     schema.TypeString,
						},
					},
				},
			},
			"string_list": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"value": {
							Required: true,
							Type:     schema.TypeList,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
		},
	}
}

func resourceConfigurationPolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubClient(ctx)

	input := &securityhub.CreateConfigurationPolicyInput{
		Name: aws.String(d.Get("name").(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.Get("security_hub_policy").([]interface{}); ok && len(v) > 0 {
		policy := expandSecurityHubPolicy(v[0].(map[string]interface{}))
		if err := validateSecurityHubPolicy(policy); err != nil {
			requestBody, _ := json.MarshalIndent(policy, "", " ")
			return sdkdiag.AppendErrorf(diags, "creating Security Hub Configuration Policy (%s): %s, %s", *input.Name, err, string(requestBody))
		}
		input.ConfigurationPolicy = policy
	}

	out, err := conn.CreateConfigurationPolicy(ctx, input)
	if err != nil {
		requestBody, _ := json.MarshalIndent(input, "", " ")
		return sdkdiag.AppendErrorf(diags, "creating Security Hub Configuration Policy (%s): %s for request %s", *input.Name, err, string(requestBody))
	}

	if d.IsNewResource() {
		d.SetId(*out.Arn)
	}

	return append(diags, resourceConfigurationPolicyRead(ctx, d, meta)...)
}

func resourceConfigurationPolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubClient(ctx)

	input := &securityhub.UpdateConfigurationPolicyInput{
		Identifier: aws.String(d.Id()),
		Name:       aws.String(d.Get("name").(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.Get("security_hub_policy").([]interface{}); ok && len(v) > 0 {
		policy := expandSecurityHubPolicy(v[0].(map[string]interface{}))
		if err := validateSecurityHubPolicy(policy); err != nil {
			requestBody, _ := json.MarshalIndent(policy, "", " ")
			return sdkdiag.AppendErrorf(diags, "updating Security Hub Configuration Policy (%s): %s, %s", d.Id(), err, requestBody)
		}
		input.ConfigurationPolicy = policy
	}

	_, err := conn.UpdateConfigurationPolicy(ctx, input)
	if err != nil {
		requestBody, _ := json.MarshalIndent(input, "", " ")
		return sdkdiag.AppendErrorf(diags, "updating Security Hub Configuration Policy (%s): %s from request:\n%s", d.Id(), err, string(requestBody))
	}

	return append(diags, resourceConfigurationPolicyRead(ctx, d, meta)...)
}

func resourceConfigurationPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubClient(ctx)

	input := &securityhub.GetConfigurationPolicyInput{
		Identifier: aws.String(d.Id()),
	}

	out, err := conn.GetConfigurationPolicy(ctx, input)
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Security Hub Configuration Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Security Hub Configuration Policy (%s): %s", d.Id(), err)
	}

	d.Set("name", out.Name)
	d.Set("description", out.Description)
	if err := d.Set("security_hub_policy", []interface{}{flattenSecurityHubPolicy(out.ConfigurationPolicy)}); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting security_hub_policy: %s", err)
	}

	return diags
}

func resourceConfigurationPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubClient(ctx)

	input := &securityhub.DeleteConfigurationPolicyInput{
		Identifier: aws.String(d.Id()),
	}

	_, err := conn.DeleteConfigurationPolicy(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Security Hub Configuration Policy (%s): %s", d.Id(), err)
	}

	return diags
}

// validateSecurityHubPolicy performs validation before running creates/updates to prevent certain issues with state.
func validateSecurityHubPolicy(apiPolicy *types.PolicyMemberSecurityHub) error {
	// security_controls_configuration can be specified in Creates/Updates and accepted by the APIs,
	// but the resources returned by subsequent Get API call will be nil instead of non-nil.
	// This leaves terraform in perpetual drift and so we prevent this case explicitly.
	if !*apiPolicy.Value.ServiceEnabled && apiPolicy.Value.SecurityControlsConfiguration != nil {
		return errors.New("security_controls_configuration cannot be defined when service_enabled is false")
	} else if *apiPolicy.Value.ServiceEnabled && apiPolicy.Value.SecurityControlsConfiguration == nil {
		return errors.New("security_controls_configuration must be defined when service_enabled is true")
	}

	// If ServiceEnabled is true, then Create/Update APIs require exactly one of enabled or disable control fields to be non-nil.
	// If terraform defaults are set for both, then we choose to set DisabledSecurityControlIdentifiers to the empty struct.
	if *apiPolicy.Value.ServiceEnabled && apiPolicy.Value.SecurityControlsConfiguration != nil &&
		apiPolicy.Value.SecurityControlsConfiguration.DisabledSecurityControlIdentifiers == nil &&
		apiPolicy.Value.SecurityControlsConfiguration.EnabledSecurityControlIdentifiers == nil {
		apiPolicy.Value.SecurityControlsConfiguration.DisabledSecurityControlIdentifiers = []string{}
	}

	return nil
}

func expandSecurityHubPolicy(tfMap map[string]interface{}) *types.PolicyMemberSecurityHub {
	if tfMap == nil {
		return nil
	}

	apiObject := types.SecurityHubPolicy{}
	apiObject.ServiceEnabled = aws.Bool(tfMap["service_enabled"].(bool))
	for _, s := range tfMap["enabled_standard_arns"].([]interface{}) {
		apiObject.EnabledStandardIdentifiers = append(apiObject.EnabledStandardIdentifiers, s.(string))
	}
	apiObject.SecurityControlsConfiguration = expandSecurityControlsConfiguration(tfMap["security_controls_configuration"])
	return &types.PolicyMemberSecurityHub{
		Value: apiObject,
	}
}

func expandSecurityControlsConfiguration(tfSecurityControlsConfig interface{}) *types.SecurityControlsConfiguration {
	var apiSecurityControlsConfig *types.SecurityControlsConfiguration
	if v, ok := tfSecurityControlsConfig.([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiControlsConfig := &types.SecurityControlsConfiguration{}

		tfControlsConfig := v[0].(map[string]interface{})
		if v, ok := tfControlsConfig["disabled_control_identifiers"]; ok && v != nil {
			for _, c := range v.([]interface{}) {
				apiControlsConfig.DisabledSecurityControlIdentifiers = append(apiControlsConfig.DisabledSecurityControlIdentifiers, c.(string))
			}
		}
		if v, ok := tfControlsConfig["enabled_control_identifiers"]; ok && v != nil {
			for _, c := range v.([]interface{}) {
				apiControlsConfig.EnabledSecurityControlIdentifiers = append(apiControlsConfig.EnabledSecurityControlIdentifiers, c.(string))
			}
		}

		if v, ok := tfControlsConfig["control_custom_parameter"].([]interface{}); ok && len(v) > 0 {
			for _, param := range v {
				apiControlsConfig.SecurityControlCustomParameters = append(apiControlsConfig.SecurityControlCustomParameters, expandControlCustomParameter(param.(map[string]interface{})))
			}
		}
		apiSecurityControlsConfig = apiControlsConfig
	} else if ok && len(v) > 0 && v[0] == nil { // resource defined, but with defaults
		apiSecurityControlsConfig = &types.SecurityControlsConfiguration{}
	} // else resource undefined yields nil
	return apiSecurityControlsConfig
}

func expandControlCustomParameter(tfCustomParam map[string]interface{}) types.SecurityControlCustomParameter {
	apiCustomParam := types.SecurityControlCustomParameter{
		Parameters: make(map[string]types.ParameterConfiguration),
	}
	if v, ok := tfCustomParam["control_identifier"].(string); ok {
		apiCustomParam.SecurityControlId = aws.String(v)
	}
	if v, ok := tfCustomParam["parameter"].(*schema.Set); ok && v.Len() > 0 {
		for _, vp := range v.List() {
			param, ok := vp.(map[string]interface{})
			if !ok {
				continue
			}
			apiParamConfig := types.ParameterConfiguration{}
			if v, ok := param["value_type"].(string); ok && len(v) > 0 {
				apiParamConfig.ValueType = types.ParameterValueType(v)
			}

			var apiParamValue types.ParameterValue
			if v, ok := param["bool"].([]interface{}); ok && len(v) > 0 { // block defined
				apiParamValue = &types.ParameterValueMemberBoolean{}
				if v[0] != nil { // block defined with non-defaults
					val := v[0].(map[string]interface{})["value"]
					apiParamValue = &types.ParameterValueMemberBoolean{Value: val.(bool)}
				}
			} else if v, ok := param["double"].([]interface{}); ok && len(v) > 0 {
				apiParamValue = &types.ParameterValueMemberDouble{}
				if v[0] != nil {
					val := v[0].(map[string]interface{})["value"]
					apiParamValue = &types.ParameterValueMemberDouble{Value: val.(float64)}
				}
			} else if v, ok := param["enum"].([]interface{}); ok && len(v) > 0 {
				apiParamValue = &types.ParameterValueMemberEnum{}
				if v[0] != nil {
					val := v[0].(map[string]interface{})["value"]
					apiParamValue = &types.ParameterValueMemberEnum{Value: val.(string)}
				}
			} else if v, ok := param["string"].([]interface{}); ok && len(v) > 0 {
				apiParamValue = &types.ParameterValueMemberString{}
				if v[0] != nil {
					val := v[0].(map[string]interface{})["value"]
					apiParamValue = &types.ParameterValueMemberString{Value: val.(string)}
				}
			} else if v, ok := param["int"].([]interface{}); ok && len(v) > 0 {
				apiParamValue = &types.ParameterValueMemberInteger{}
				if v[0] != nil {
					val := v[0].(map[string]interface{})["value"]
					apiParamValue = &types.ParameterValueMemberInteger{Value: int32(val.(int))}
				}
			} else if v, ok := param["int_list"].([]interface{}); ok && len(v) > 0 {
				apiParamValue = &types.ParameterValueMemberIntegerList{}
				if v[0] != nil {
					val := v[0].(map[string]interface{})["value"]
					var vals []int32
					for _, s := range val.([]interface{}) {
						vals = append(vals, int32(s.(int)))
					}
					apiParamValue = &types.ParameterValueMemberIntegerList{Value: vals}
				}
			} else if v, ok := param["enum_list"].([]interface{}); ok && len(v) > 0 {
				apiParamValue = &types.ParameterValueMemberEnumList{}
				if v[0] != nil {
					val := v[0].(map[string]interface{})["value"]
					var vals []string
					for _, s := range val.([]interface{}) {
						vals = append(vals, s.(string))
					}
					apiParamValue = &types.ParameterValueMemberEnumList{Value: vals}
				}
			} else if v, ok := param["string_list"].([]interface{}); ok && len(v) > 0 {
				apiParamValue = &types.ParameterValueMemberStringList{}
				if v[0] != nil {
					val := v[0].(map[string]interface{})["value"]
					var vals []string
					for _, s := range val.([]interface{}) {
						vals = append(vals, s.(string))
					}
					apiParamValue = &types.ParameterValueMemberStringList{Value: vals}
				}
			}
			apiParamConfig.Value = apiParamValue
			if key, ok := param["name"].(string); ok && len(key) > 0 {
				apiCustomParam.Parameters[key] = apiParamConfig
			}
		}
	}
	return apiCustomParam
}

func flattenSecurityHubPolicy(policy types.Policy) map[string]interface{} {
	apiObject, ok := policy.(*types.PolicyMemberSecurityHub)
	if !ok || apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	tfMap["service_enabled"] = apiObject.Value.ServiceEnabled
	tfMap["enabled_standard_arns"] = apiObject.Value.EnabledStandardIdentifiers
	tfMap["security_controls_configuration"] = flattenSecurityControlsConfiguration(apiObject.Value.SecurityControlsConfiguration)
	return tfMap
}

func flattenSecurityControlsConfiguration(apiSecurityControlsConfig *types.SecurityControlsConfiguration) []interface{} {
	if apiSecurityControlsConfig == nil {
		return nil
	}
	tfSecurityControlsConfig := map[string]interface{}{}
	tfSecurityControlsConfig["disabled_control_identifiers"] = apiSecurityControlsConfig.DisabledSecurityControlIdentifiers
	tfSecurityControlsConfig["enabled_control_identifiers"] = apiSecurityControlsConfig.EnabledSecurityControlIdentifiers
	tfControlCustomParams := []interface{}{}
	for _, apiControlCustomParam := range apiSecurityControlsConfig.SecurityControlCustomParameters {
		tfControlCustomParams = append(tfControlCustomParams, flattenControlCustomParameter(apiControlCustomParam))
	}
	tfSecurityControlsConfig["control_custom_parameter"] = tfControlCustomParams
	return []interface{}{tfSecurityControlsConfig}
}

func flattenControlCustomParameter(apiControlCustomParam types.SecurityControlCustomParameter) map[string]interface{} {
	tfControlCustomParam := map[string]interface{}{}
	tfControlCustomParam["control_identifier"] = apiControlCustomParam.SecurityControlId
	tfParametersForControl := []interface{}{}
	for paramName, param := range apiControlCustomParam.Parameters {
		tfParameter := map[string]interface{}{
			"name":       paramName,
			"value_type": string(param.ValueType),
		}
		if param.Value != nil {
			switch casted := param.Value.(type) {
			case *types.ParameterValueMemberBoolean:
				tfParameter["bool"] = []interface{}{
					map[string]interface{}{
						"value": casted.Value,
					},
				}
			case *types.ParameterValueMemberDouble:
				tfParameter["double"] = []interface{}{
					map[string]interface{}{
						"value": casted.Value,
					},
				}
			case *types.ParameterValueMemberEnum:
				tfParameter["enum"] = []interface{}{
					map[string]interface{}{
						"value": casted.Value,
					},
				}
			case *types.ParameterValueMemberEnumList:
				tfParameter["enum_list"] = []interface{}{
					map[string]interface{}{
						"value": casted.Value,
					},
				}
			case *types.ParameterValueMemberInteger:
				tfParameter["int"] = []interface{}{
					map[string]interface{}{
						"value": casted.Value,
					},
				}
			case *types.ParameterValueMemberIntegerList:
				tfParameter["int_list"] = []interface{}{
					map[string]interface{}{
						"value": casted.Value,
					},
				}
			case *types.ParameterValueMemberString:
				tfParameter["string"] = []interface{}{
					map[string]interface{}{
						"value": casted.Value,
					},
				}
			case *types.ParameterValueMemberStringList:
				tfParameter["string_list"] = []interface{}{
					map[string]interface{}{
						"value": casted.Value,
					},
				}
			default:
				log.Printf("[WARN] Security Hub Configuration Policy (%T) unknown type of parameter value", casted)
			}
		}

		tfParametersForControl = append(tfParametersForControl, tfParameter)
	}
	tfControlCustomParam["parameter"] = tfParametersForControl
	return tfControlCustomParam
}
