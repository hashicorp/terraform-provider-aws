// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package configservice

import (
	"context"
	"log"
	"slices"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/configservice"
	"github.com/aws/aws-sdk-go-v2/service/configservice/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_config_remediation_configuration", name="Remediation Configuration")
func resourceRemediationConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRemediationConfigurationPut,
		ReadWithoutTimeout:   resourceRemediationConfigurationRead,
		UpdateWithoutTimeout: resourceRemediationConfigurationPut,
		DeleteWithoutTimeout: resourceRemediationConfigurationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"automatic": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"config_rule_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
			"execution_controls": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ssm_controls": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"concurrent_execution_rate_percentage": {
										Type:         schema.TypeInt,
										Optional:     true,
										ValidateFunc: validation.IntBetween(1, 100),
									},
									"error_percentage": {
										Type:         schema.TypeInt,
										Optional:     true,
										ValidateFunc: validation.IntBetween(1, 100),
									},
								},
							},
						},
					},
				},
			},
			"maximum_automatic_attempts": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(1, 25),
			},
			names.AttrParameter: {
				Type:     schema.TypeList,
				MaxItems: 25,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrName: {
							Type:     schema.TypeString,
							Required: true,
						},
						"resource_value": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(0, 256),
						},
						"static_value": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"static_values": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			names.AttrResourceType: {
				Type:     schema.TypeString,
				Optional: true,
			},
			"retry_attempt_seconds": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(1, 2678000),
			},
			"target_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 256),
			},
			"target_type": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: enum.Validate[types.RemediationTargetType](),
			},
			"target_version": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceRemediationConfigurationPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConfigServiceClient(ctx)

	name := d.Get("config_rule_name").(string)
	remediationConfiguration := types.RemediationConfiguration{
		ConfigRuleName: aws.String(name),
	}

	if v, ok := d.GetOk("automatic"); ok {
		remediationConfiguration.Automatic = v.(bool)
	}

	if v, ok := d.GetOk("execution_controls"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		remediationConfiguration.ExecutionControls = expandExecutionControls(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("maximum_automatic_attempts"); ok {
		remediationConfiguration.MaximumAutomaticAttempts = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk(names.AttrParameter); ok && len(v.([]interface{})) > 0 {
		remediationConfiguration.Parameters = expandRemediationParameterValues(v.([]interface{}))
	}

	if v, ok := d.GetOk(names.AttrResourceType); ok {
		remediationConfiguration.ResourceType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("retry_attempt_seconds"); ok {
		remediationConfiguration.RetryAttemptSeconds = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("target_id"); ok {
		remediationConfiguration.TargetId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("target_type"); ok {
		remediationConfiguration.TargetType = types.RemediationTargetType(v.(string))
	}

	if v, ok := d.GetOk("target_version"); ok {
		remediationConfiguration.TargetVersion = aws.String(v.(string))
	}

	_, err := conn.PutRemediationConfigurations(ctx, &configservice.PutRemediationConfigurationsInput{
		RemediationConfigurations: []types.RemediationConfiguration{remediationConfiguration},
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "putting ConfigService Remediation Configuration (%s): %s", name, err)
	}

	if d.IsNewResource() {
		d.SetId(name)
	}

	return append(diags, resourceRemediationConfigurationRead(ctx, d, meta)...)
}

func resourceRemediationConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConfigServiceClient(ctx)

	remediationConfiguration, err := findRemediationConfigurationByConfigRuleName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ConfigService Remediation Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ConfigService Remediation Configuration (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, remediationConfiguration.Arn)
	d.Set("automatic", remediationConfiguration.Automatic)
	d.Set("config_rule_name", remediationConfiguration.ConfigRuleName)
	if err := d.Set("execution_controls", flattenExecutionControls(remediationConfiguration.ExecutionControls)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting execution_controls: %s", err)
	}
	d.Set("maximum_automatic_attempts", remediationConfiguration.MaximumAutomaticAttempts)
	if err := d.Set(names.AttrParameter, flattenRemediationParameterValues(remediationConfiguration.Parameters)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting parameter: %s", err)
	}
	d.Set(names.AttrResourceType, remediationConfiguration.ResourceType)
	d.Set("retry_attempt_seconds", remediationConfiguration.RetryAttemptSeconds)
	d.Set("target_id", remediationConfiguration.TargetId)
	d.Set("target_type", remediationConfiguration.TargetType)
	d.Set("target_version", remediationConfiguration.TargetVersion)

	return diags
}

func resourceRemediationConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConfigServiceClient(ctx)

	input := &configservice.DeleteRemediationConfigurationInput{
		ConfigRuleName: aws.String(d.Id()),
	}

	if v, ok := d.GetOk(names.AttrResourceType); ok {
		input.ResourceType = aws.String(v.(string))
	}

	const (
		timeout = 2 * time.Minute
	)
	log.Printf("[DEBUG] Deleting ConfigService Remediation Configuration: %s", d.Id())
	_, err := tfresource.RetryWhenIsA[*types.ResourceInUseException](ctx, timeout, func() (interface{}, error) {
		return conn.DeleteRemediationConfiguration(ctx, input)
	})

	if errs.IsA[*types.NoSuchRemediationConfigurationException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting ConfigService Remediation Configuration (%s): %s", d.Id(), err)
	}

	return diags
}

func findRemediationConfigurationByConfigRuleName(ctx context.Context, conn *configservice.Client, name string) (*types.RemediationConfiguration, error) {
	input := &configservice.DescribeRemediationConfigurationsInput{
		ConfigRuleNames: []string{name},
	}

	return findRemediationConfiguration(ctx, conn, input)
}

func findRemediationConfiguration(ctx context.Context, conn *configservice.Client, input *configservice.DescribeRemediationConfigurationsInput) (*types.RemediationConfiguration, error) {
	output, err := findRemediationConfigurations(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findRemediationConfigurations(ctx context.Context, conn *configservice.Client, input *configservice.DescribeRemediationConfigurationsInput) ([]types.RemediationConfiguration, error) {
	output, err := conn.DescribeRemediationConfigurations(ctx, input)

	if errs.IsA[*types.NoSuchConfigRuleException](err) {
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

	return output.RemediationConfigurations, nil
}

func expandRemediationParameterValue(tfMap map[string]interface{}) types.RemediationParameterValue {
	apiObject := types.RemediationParameterValue{}

	if v, ok := tfMap["resource_value"].(string); ok && v != "" {
		apiObject.ResourceValue = &types.ResourceValue{
			Value: types.ResourceValueType(v),
		}
	}

	if v, ok := tfMap["static_value"].(string); ok && v != "" {
		apiObject.StaticValue = &types.StaticValue{
			Values: []string{v},
		}
	}

	if v, ok := tfMap["static_values"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.StaticValue = &types.StaticValue{
			Values: flex.ExpandStringValueList(v),
		}
	}

	return apiObject
}

func expandRemediationParameterValues(tfList []interface{}) map[string]types.RemediationParameterValue {
	if len(tfList) == 0 {
		return nil
	}

	apiObjects := make(map[string]types.RemediationParameterValue)

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		if v, ok := tfMap[names.AttrName].(string); !ok || v == "" {
			continue
		}

		apiObjects[tfMap[names.AttrName].(string)] = expandRemediationParameterValue(tfMap)
	}

	return apiObjects
}

func expandSSMControls(tfMap map[string]interface{}) *types.SsmControls {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.SsmControls{}

	if v, ok := tfMap["concurrent_execution_rate_percentage"].(int); ok && v != 0 {
		apiObject.ConcurrentExecutionRatePercentage = aws.Int32(int32(v))
	}

	if v, ok := tfMap["error_percentage"].(int); ok && v != 0 {
		apiObject.ErrorPercentage = aws.Int32(int32(v))
	}

	return apiObject
}

func expandExecutionControls(tfMap map[string]interface{}) *types.ExecutionControls {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.ExecutionControls{}

	if v, ok := tfMap["ssm_controls"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.SsmControls = expandSSMControls(v[0].(map[string]interface{}))
	}

	return apiObject
}

func flattenRemediationParameterValues(apiObjects map[string]types.RemediationParameterValue) []interface{} {
	var tfList []interface{}

	for key, value := range apiObjects {
		tfMap := map[string]interface{}{
			names.AttrName: key,
		}

		if v := value.ResourceValue; v != nil {
			tfMap["resource_value"] = v.Value
		}

		if v := value.StaticValue; v != nil {
			if len(v.Values) == 1 {
				tfMap["static_value"] = v.Values[0]
			} else if len(v.Values) > 1 {
				tfMap["static_values"] = v.Values
			}
		} else {
			tfMap["static_values"] = make([]interface{}, 0)
		}

		tfList = append(tfList, tfMap)
	}

	slices.SortFunc(tfList, func(a, b interface{}) int {
		if a.(map[string]interface{})[names.AttrName].(string) < b.(map[string]interface{})[names.AttrName].(string) {
			return -1
		}

		if a.(map[string]interface{})[names.AttrName].(string) > b.(map[string]interface{})[names.AttrName].(string) {
			return 1
		}

		return 0
	})

	return tfList
}

func flattenExecutionControls(apiObject *types.ExecutionControls) []interface{} {
	if apiObject == nil {
		return nil
	}

	return []interface{}{map[string]interface{}{
		"ssm_controls": flattenSSMControls(apiObject.SsmControls),
	}}
}

func flattenSSMControls(apiObject *types.SsmControls) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.ConcurrentExecutionRatePercentage != nil {
		tfMap["concurrent_execution_rate_percentage"] = apiObject.ConcurrentExecutionRatePercentage
	}

	if apiObject.ErrorPercentage != nil {
		tfMap["error_percentage"] = apiObject.ErrorPercentage
	}
	return []interface{}{tfMap}
}
