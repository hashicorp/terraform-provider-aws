// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securityhub

import (
	"context"
	"errors"
	"log"
	"math"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/securityhub"
	"github.com/aws/aws-sdk-go-v2/service/securityhub/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_securityhub_configuration_policy", name="Configuration Policy")
func resourceConfigurationPolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceConfigurationPolicyCreate,
		ReadWithoutTimeout:   resourceConfigurationPolicyRead,
		UpdateWithoutTimeout: resourceConfigurationPolicyUpdate,
		DeleteWithoutTimeout: resourceConfigurationPolicyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		SchemaFunc: func() map[string]*schema.Schema {
			customParameterResource := func() *schema.Resource {
				return &schema.Resource{
					Schema: map[string]*schema.Schema{
						"bool": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrValue: {
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
									names.AttrValue: {
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
									names.AttrValue: {
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
									names.AttrValue: {
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
									names.AttrValue: {
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
									names.AttrValue: {
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
						names.AttrName: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},
						"string": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrValue: {
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
									names.AttrValue: {
										Required: true,
										Type:     schema.TypeList,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
								},
							},
						},
						"value_type": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[types.ParameterValueType](),
						},
					},
				}
			}

			return map[string]*schema.Schema{
				names.AttrARN: {
					Type:     schema.TypeString,
					Computed: true,
				},
				"configuration_policy": {
					Type:     schema.TypeList,
					Required: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"enabled_standard_arns": {
								Type:     schema.TypeSet,
								Optional: true,
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
											Type:     schema.TypeSet,
											Optional: true,
											Elem: &schema.Schema{
												Type:         schema.TypeString,
												ValidateFunc: validation.StringIsNotEmpty,
											},
											ConflictsWith: []string{
												"configuration_policy.0.security_controls_configuration.0.enabled_control_identifiers",
											},
										},
										"enabled_control_identifiers": {
											Type:     schema.TypeSet,
											Optional: true,
											Elem: &schema.Schema{
												Type:         schema.TypeString,
												ValidateFunc: validation.StringIsNotEmpty,
											},
											ConflictsWith: []string{
												"configuration_policy.0.security_controls_configuration.0.disabled_control_identifiers",
											},
										},
										"security_control_custom_parameter": {
											Type:     schema.TypeList,
											Optional: true,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													names.AttrParameter: {
														Type:     schema.TypeSet,
														Required: true,
														MinItems: 1,
														Elem:     customParameterResource(),
													},
													"security_control_id": {
														Type:         schema.TypeString,
														Required:     true,
														ValidateFunc: validation.StringIsNotEmpty,
													},
												},
											},
										},
									},
								},
							},
							"service_enabled": {
								Type:     schema.TypeBool,
								Required: true,
							},
						},
					},
				},
				names.AttrDescription: {
					Type:     schema.TypeString,
					Optional: true,
				},
				names.AttrName: {
					Type:     schema.TypeString,
					Required: true,
					ValidateFunc: validation.StringMatch(
						regexache.MustCompile(`[A-Za-z0-9\-\.!*/]+`),
						"Only alphanumeric characters and the following ASCII characters are permitted: -, ., !, *, /",
					),
				},
			}
		},
	}
}

func resourceConfigurationPolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &securityhub.CreateConfigurationPolicyInput{
		Name: aws.String(name),
	}

	if v, ok := d.GetOk("configuration_policy"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		policy := expandPolicyMemberSecurityHub(v.([]interface{})[0].(map[string]interface{}))
		if err := validatePolicyMemberSecurityHub(policy); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
		input.ConfigurationPolicy = policy
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	output, err := conn.CreateConfigurationPolicy(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Security Hub Configuration Policy (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.Id))

	return append(diags, resourceConfigurationPolicyRead(ctx, d, meta)...)
}

func resourceConfigurationPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubClient(ctx)

	output, err := findConfigurationPolicyByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Security Hub Configuration Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Security Hub Configuration Policy (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, output.Arn)
	if err := d.Set("configuration_policy", []interface{}{flattenPolicy(output.ConfigurationPolicy)}); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting configuration_policy: %s", err)
	}
	d.Set(names.AttrDescription, output.Description)
	d.Set(names.AttrName, output.Name)

	return diags
}

func resourceConfigurationPolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubClient(ctx)

	input := &securityhub.UpdateConfigurationPolicyInput{
		Identifier: aws.String(d.Id()),
		Name:       aws.String(d.Get(names.AttrName).(string)),
	}

	if v, ok := d.GetOk("configuration_policy"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		policy := expandPolicyMemberSecurityHub(v.([]interface{})[0].(map[string]interface{}))
		if err := validatePolicyMemberSecurityHub(policy); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
		input.ConfigurationPolicy = policy
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	_, err := conn.UpdateConfigurationPolicy(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Security Hub Configuration Policy (%s): %s", d.Id(), err)
	}

	return append(diags, resourceConfigurationPolicyRead(ctx, d, meta)...)
}

func resourceConfigurationPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubClient(ctx)

	log.Printf("[DEBUG] Deleting Security Hub Configuration Policy: %s", d.Id())
	_, err := conn.DeleteConfigurationPolicy(ctx, &securityhub.DeleteConfigurationPolicyInput{
		Identifier: aws.String(d.Id()),
	})

	if tfawserr.ErrMessageContains(err, errCodeAccessDeniedException, "Must be a Security Hub delegated administrator with Central Configuration enabled") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Security Hub Configuration Policy (%s): %s", d.Id(), err)
	}

	return diags
}

func findConfigurationPolicyByID(ctx context.Context, conn *securityhub.Client, id string) (*securityhub.GetConfigurationPolicyOutput, error) {
	input := &securityhub.GetConfigurationPolicyInput{
		Identifier: aws.String(id),
	}

	return findConfigurationPolicy(ctx, conn, input)
}

func findConfigurationPolicy(ctx context.Context, conn *securityhub.Client, input *securityhub.GetConfigurationPolicyInput) (*securityhub.GetConfigurationPolicyOutput, error) {
	output, err := conn.GetConfigurationPolicy(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeResourceNotFoundException) || tfawserr.ErrMessageContains(err, errCodeAccessDeniedException, "Must be a Security Hub delegated administrator with Central Configuration enabled") {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

// validatePolicyMemberSecurityHub performs validation before running creates/updates to prevent certain issues with state.
func validatePolicyMemberSecurityHub(apiPolicy *types.PolicyMemberSecurityHub) error { // nosemgrep:ci.securityhub-in-func-name
	// security_controls_configuration can be specified in Creates/Updates and accepted by the APIs,
	// but the resources returned by subsequent Get API call will be nil instead of non-nil.
	// This leaves terraform in perpetual drift and so we prevent this case explicitly.
	if !aws.ToBool(apiPolicy.Value.ServiceEnabled) && apiPolicy.Value.SecurityControlsConfiguration != nil {
		return errors.New("security_controls_configuration cannot be defined when service_enabled is false")
	} else if aws.ToBool(apiPolicy.Value.ServiceEnabled) && apiPolicy.Value.SecurityControlsConfiguration == nil {
		return errors.New("security_controls_configuration must be defined when service_enabled is true")
	}

	// If ServiceEnabled is true, then Create/Update APIs require exactly one of enabled or disable control fields to be non-nil.
	// If terraform defaults are set for both, then we choose to set DisabledSecurityControlIdentifiers to the empty struct.
	if aws.ToBool(apiPolicy.Value.ServiceEnabled) && apiPolicy.Value.SecurityControlsConfiguration != nil &&
		apiPolicy.Value.SecurityControlsConfiguration.DisabledSecurityControlIdentifiers == nil &&
		apiPolicy.Value.SecurityControlsConfiguration.EnabledSecurityControlIdentifiers == nil {
		apiPolicy.Value.SecurityControlsConfiguration.DisabledSecurityControlIdentifiers = []string{}
	}

	return nil
}

func expandPolicyMemberSecurityHub(tfMap map[string]interface{}) *types.PolicyMemberSecurityHub { // nosemgrep:ci.securityhub-in-func-name
	if tfMap == nil {
		return nil
	}

	apiObject := types.SecurityHubPolicy{
		SecurityControlsConfiguration: expandSecurityControlsConfiguration(tfMap["security_controls_configuration"]),
	}

	if v, ok := tfMap["service_enabled"].(bool); ok {
		apiObject.ServiceEnabled = aws.Bool(v)

		if v {
			if v, ok := tfMap["enabled_standard_arns"].(*schema.Set); ok {
				apiObject.EnabledStandardIdentifiers = flex.ExpandStringValueSet(v)
			}
		}
	}

	return &types.PolicyMemberSecurityHub{
		Value: apiObject,
	}
}

func expandSecurityControlsConfiguration(tfListRaw interface{}) *types.SecurityControlsConfiguration {
	var apiObject *types.SecurityControlsConfiguration

	if v, ok := tfListRaw.([]interface{}); ok && len(v) > 0 && v[0] != nil {
		tfMap := v[0].(map[string]interface{})
		apiObject = &types.SecurityControlsConfiguration{}

		if v, ok := tfMap["disabled_control_identifiers"].(*schema.Set); ok && v.Len() > 0 {
			apiObject.DisabledSecurityControlIdentifiers = flex.ExpandStringValueSet(v)
		}

		if v, ok := tfMap["enabled_control_identifiers"].(*schema.Set); ok && v.Len() > 0 {
			apiObject.EnabledSecurityControlIdentifiers = flex.ExpandStringValueSet(v)
		}

		if v, ok := tfMap["security_control_custom_parameter"].([]interface{}); ok && len(v) > 0 {
			for _, tfMapRaw := range v {
				tfMap, ok := tfMapRaw.(map[string]interface{})
				if !ok {
					continue
				}

				apiObject.SecurityControlCustomParameters = append(apiObject.SecurityControlCustomParameters, expandSecurityControlCustomParameter(tfMap))
			}
		}
	} else if ok && len(v) > 0 && v[0] == nil { // resource defined, but with defaults
		apiObject = &types.SecurityControlsConfiguration{}
	} // else resource undefined yields nil

	return apiObject
}

func expandSecurityControlCustomParameter(tfMap map[string]interface{}) types.SecurityControlCustomParameter {
	apiObject := types.SecurityControlCustomParameter{
		Parameters: make(map[string]types.ParameterConfiguration),
	}

	if v, ok := tfMap["security_control_id"].(string); ok {
		apiObject.SecurityControlId = aws.String(v)
	}

	if v, ok := tfMap[names.AttrParameter].(*schema.Set); ok && v.Len() > 0 {
		for _, tfMapRaw := range v.List() {
			tfMap, ok := tfMapRaw.(map[string]interface{})
			if !ok {
				continue
			}

			parameterConfiguration := types.ParameterConfiguration{}

			if v, ok := tfMap["value_type"].(string); ok {
				parameterConfiguration.ValueType = types.ParameterValueType(v)
			}

			var parameterValue types.ParameterValue

			if v, ok := tfMap["bool"].([]interface{}); ok && len(v) > 0 { // block defined
				parameterValue = &types.ParameterValueMemberBoolean{}
				if v[0] != nil { // block defined with non-defaults
					val := v[0].(map[string]interface{})[names.AttrValue]
					parameterValue = &types.ParameterValueMemberBoolean{Value: val.(bool)}
				}
			} else if v, ok := tfMap["double"].([]interface{}); ok && len(v) > 0 {
				parameterValue = &types.ParameterValueMemberDouble{}
				if v[0] != nil {
					val := v[0].(map[string]interface{})[names.AttrValue]
					parameterValue = &types.ParameterValueMemberDouble{Value: val.(float64)}
				}
			} else if v, ok := tfMap["enum"].([]interface{}); ok && len(v) > 0 {
				parameterValue = &types.ParameterValueMemberEnum{}
				if v[0] != nil {
					val := v[0].(map[string]interface{})[names.AttrValue]
					parameterValue = &types.ParameterValueMemberEnum{Value: val.(string)}
				}
			} else if v, ok := tfMap["string"].([]interface{}); ok && len(v) > 0 {
				parameterValue = &types.ParameterValueMemberString{}
				if v[0] != nil {
					val := v[0].(map[string]interface{})[names.AttrValue]
					parameterValue = &types.ParameterValueMemberString{Value: val.(string)}
				}
			} else if v, ok := tfMap["int"].([]interface{}); ok && len(v) > 0 {
				parameterValue = &types.ParameterValueMemberInteger{}
				if v[0] != nil {
					val := v[0].(map[string]interface{})[names.AttrValue]
					parameterValue = &types.ParameterValueMemberInteger{Value: int32(val.(int))}
				}
			} else if v, ok := tfMap["int_list"].([]interface{}); ok && len(v) > 0 {
				parameterValue = &types.ParameterValueMemberIntegerList{}
				if v[0] != nil {
					val := v[0].(map[string]interface{})[names.AttrValue]
					var vals []int32
					for _, s := range val.([]interface{}) {
						vals = append(vals, int32(s.(int)))
					}
					parameterValue = &types.ParameterValueMemberIntegerList{Value: vals}
				}
			} else if v, ok := tfMap["enum_list"].([]interface{}); ok && len(v) > 0 {
				parameterValue = &types.ParameterValueMemberEnumList{}
				if v[0] != nil {
					val := v[0].(map[string]interface{})[names.AttrValue]
					var vals []string
					for _, s := range val.([]interface{}) {
						vals = append(vals, s.(string))
					}
					parameterValue = &types.ParameterValueMemberEnumList{Value: vals}
				}
			} else if v, ok := tfMap["string_list"].([]interface{}); ok && len(v) > 0 {
				parameterValue = &types.ParameterValueMemberStringList{}
				if v[0] != nil {
					val := v[0].(map[string]interface{})[names.AttrValue]
					var vals []string
					for _, s := range val.([]interface{}) {
						vals = append(vals, s.(string))
					}
					parameterValue = &types.ParameterValueMemberStringList{Value: vals}
				}
			}

			parameterConfiguration.Value = parameterValue

			if v, ok := tfMap[names.AttrName].(string); ok && len(v) > 0 {
				apiObject.Parameters[v] = parameterConfiguration
			}
		}
	}

	return apiObject
}

func flattenPolicy(apiObject types.Policy) map[string]interface{} {
	switch apiObject := apiObject.(type) {
	case *types.PolicyMemberSecurityHub:
		return flattenPolicyMemberSecurityHub(apiObject)
	}

	return nil
}

func flattenPolicyMemberSecurityHub(apiObject *types.PolicyMemberSecurityHub) map[string]interface{} { // nosemgrep:ci.securityhub-in-func-name
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"enabled_standard_arns":           apiObject.Value.EnabledStandardIdentifiers,
		"security_controls_configuration": flattenSecurityControlsConfiguration(apiObject.Value.SecurityControlsConfiguration),
	}

	if v := apiObject.Value.ServiceEnabled; v != nil {
		tfMap["service_enabled"] = aws.ToBool(v)
	}

	return tfMap
}

func flattenSecurityControlsConfiguration(apiObject *types.SecurityControlsConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"disabled_control_identifiers": apiObject.DisabledSecurityControlIdentifiers,
		"enabled_control_identifiers":  apiObject.EnabledSecurityControlIdentifiers,
	}

	var tfList []interface{}

	for _, apiObject := range apiObject.SecurityControlCustomParameters {
		tfList = append(tfList, flattenSecurityControlCustomParameter(apiObject))
	}

	tfMap["security_control_custom_parameter"] = tfList

	return []interface{}{tfMap}
}

func flattenSecurityControlCustomParameter(apiObject types.SecurityControlCustomParameter) map[string]interface{} {
	tfMap := map[string]interface{}{}

	if v := apiObject.SecurityControlId; v != nil {
		tfMap["security_control_id"] = aws.ToString(v)
	}

	var tfList []interface{}

	for name, apiObject := range apiObject.Parameters {
		tfMap := map[string]interface{}{
			names.AttrName: name,
			"value_type":   apiObject.ValueType,
		}

		switch apiObject := apiObject.Value.(type) {
		case *types.ParameterValueMemberBoolean:
			tfMap["bool"] = []interface{}{
				map[string]interface{}{
					names.AttrValue: apiObject.Value,
				},
			}
		case *types.ParameterValueMemberDouble:
			tfMap["double"] = []interface{}{
				map[string]interface{}{
					names.AttrValue: apiObject.Value,
				},
			}
		case *types.ParameterValueMemberEnum:
			tfMap["enum"] = []interface{}{
				map[string]interface{}{
					names.AttrValue: apiObject.Value,
				},
			}
		case *types.ParameterValueMemberEnumList:
			tfMap["enum_list"] = []interface{}{
				map[string]interface{}{
					names.AttrValue: apiObject.Value,
				},
			}
		case *types.ParameterValueMemberInteger:
			tfMap["int"] = []interface{}{
				map[string]interface{}{
					names.AttrValue: apiObject.Value,
				},
			}
		case *types.ParameterValueMemberIntegerList:
			tfMap["int_list"] = []interface{}{
				map[string]interface{}{
					names.AttrValue: apiObject.Value,
				},
			}
		case *types.ParameterValueMemberString:
			tfMap["string"] = []interface{}{
				map[string]interface{}{
					names.AttrValue: apiObject.Value,
				},
			}
		case *types.ParameterValueMemberStringList:
			tfMap["string_list"] = []interface{}{
				map[string]interface{}{
					names.AttrValue: apiObject.Value,
				},
			}
		}

		tfList = append(tfList, tfMap)
	}

	tfMap[names.AttrParameter] = tfList

	return tfMap
}
