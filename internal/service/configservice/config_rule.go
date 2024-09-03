// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package configservice

import (
	"context"
	"log"
	"time"

	"github.com/YakDriver/regexache"
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
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_config_config_rule", name="Config Rule")
// @Tags(identifierAttribute="arn")
func resourceConfigRule() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceConfigRulePut,
		ReadWithoutTimeout:   resourceConfigRuleRead,
		UpdateWithoutTimeout: resourceConfigRulePut,
		DeleteWithoutTimeout: resourceConfigRuleDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 256),
			},
			"evaluation_mode": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrMode: {
							Type:             schema.TypeString,
							Optional:         true,
							Computed:         true,
							ValidateDiagFunc: enum.Validate[types.EvaluationMode](),
						},
					},
				},
			},
			"input_parameters": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringIsJSON,
			},
			"maximum_execution_frequency": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: enum.Validate[types.MaximumExecutionFrequency](),
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 128),
			},
			"rule_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrScope: {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"compliance_resource_id": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(0, 256),
						},
						"compliance_resource_types": {
							Type:     schema.TypeSet,
							Optional: true,
							MaxItems: 100,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringLenBetween(0, 256),
							},
							Set: schema.HashString,
						},
						"tag_key": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(0, 128),
						},
						"tag_value": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(0, 256),
						},
					},
				},
			},
			names.AttrSource: {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"custom_policy_details": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"enable_debug_log_delivery": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  false,
									},
									"policy_runtime": {
										Type:     schema.TypeString,
										Required: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(0, 64),
											validation.StringMatch(regexache.MustCompile(`^guard\-2\.x\.x$`), "Must match cloudformation-guard version"),
										),
									},
									"policy_text": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringLenBetween(0, 10000),
									},
								},
							},
						},
						names.AttrOwner: {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[types.Owner](),
						},
						"source_detail": {
							Type:     schema.TypeSet,
							Optional: true,
							MaxItems: 25,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"event_source": {
										Type:             schema.TypeString,
										Optional:         true,
										Default:          types.EventSourceAwsConfig,
										ValidateDiagFunc: enum.Validate[types.EventSource](),
									},
									"maximum_execution_frequency": {
										Type:             schema.TypeString,
										Optional:         true,
										ValidateDiagFunc: enum.Validate[types.MaximumExecutionFrequency](),
									},
									"message_type": {
										Type:             schema.TypeString,
										Optional:         true,
										ValidateDiagFunc: enum.Validate[types.MessageType](),
									},
								},
							},
							Set: sdkv2.SimpleSchemaSetFunc("message_type", "event_source", "maximum_execution_frequency"),
						},
						"source_identifier": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(0, 256),
						},
					},
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceConfigRulePut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConfigServiceClient(ctx)

	if d.IsNewResource() || d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		name := d.Get(names.AttrName).(string)
		configRule := &types.ConfigRule{
			ConfigRuleName: aws.String(name),
		}

		if v, ok := d.GetOk(names.AttrDescription); ok {
			configRule.Description = aws.String(v.(string))
		}

		if v, ok := d.Get("evaluation_mode").(*schema.Set); ok && v.Len() > 0 {
			configRule.EvaluationModes = expandEvaluationModeConfigurations(v.List())
		}

		if v, ok := d.GetOk("input_parameters"); ok {
			configRule.InputParameters = aws.String(v.(string))
		}

		if v, ok := d.GetOk("maximum_execution_frequency"); ok {
			configRule.MaximumExecutionFrequency = types.MaximumExecutionFrequency(v.(string))
		}

		if v, ok := d.GetOk(names.AttrScope); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			configRule.Scope = expandScope(v.([]interface{})[0].(map[string]interface{}))
		}

		if v, ok := d.GetOk(names.AttrSource); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			configRule.Source = expandSource(v.([]interface{})[0].(map[string]interface{}))
		}

		input := &configservice.PutConfigRuleInput{
			ConfigRule: configRule,
			Tags:       getTagsIn(ctx),
		}

		_, err := tfresource.RetryWhenIsA[*types.InsufficientPermissionsException](ctx, propagationTimeout, func() (interface{}, error) {
			return conn.PutConfigRule(ctx, input)
		})

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "putting ConfigService Config Rule (%s): %s", name, err)
		}

		if d.IsNewResource() {
			d.SetId(name)
		}
	}

	return append(diags, resourceConfigRuleRead(ctx, d, meta)...)
}

func resourceConfigRuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConfigServiceClient(ctx)

	rule, err := findConfigRuleByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ConfigService Config Rule (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ConfigService Config Rule (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, rule.ConfigRuleArn)
	d.Set(names.AttrDescription, rule.Description)
	if err := d.Set("evaluation_mode", flattenEvaluationModeConfigurations(rule.EvaluationModes)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting evaluation_mode: %s", err)
	}
	d.Set("input_parameters", rule.InputParameters)
	d.Set("maximum_execution_frequency", rule.MaximumExecutionFrequency)
	d.Set(names.AttrName, rule.ConfigRuleName)
	d.Set("rule_id", rule.ConfigRuleId)
	if rule.Scope != nil {
		if err := d.Set(names.AttrScope, flattenScope(rule.Scope)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting scope: %s", err)
		}
	}
	if rule.Source != nil && rule.Source.CustomPolicyDetails != nil && aws.ToString(rule.Source.CustomPolicyDetails.PolicyText) == "" {
		// Source.CustomPolicyDetails.PolicyText is not returned by the API, so copy from state.
		if v, ok := d.GetOk("source.0.custom_policy_details.0.policy_text"); ok {
			rule.Source.CustomPolicyDetails.PolicyText = aws.String(v.(string))
		}
	}
	if err := d.Set(names.AttrSource, flattenSource(rule.Source)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting source: %s", err)
	}

	return diags
}

func resourceConfigRuleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConfigServiceClient(ctx)

	const (
		timeout = 2 * time.Minute
	)
	log.Printf("[DEBUG] Deleting ConfigService Config Rule: %s", d.Id())
	_, err := tfresource.RetryWhenIsA[*types.ResourceInUseException](ctx, timeout, func() (interface{}, error) {
		return conn.DeleteConfigRule(ctx, &configservice.DeleteConfigRuleInput{
			ConfigRuleName: aws.String(d.Id()),
		})
	})

	if errs.IsA[*types.NoSuchConfigRuleException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting ConfigService Config Rule (%s): %s", d.Id(), err)
	}

	if _, err := waitConfigRuleDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for ConfigService Config Rule (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findConfigRuleByName(ctx context.Context, conn *configservice.Client, name string) (*types.ConfigRule, error) {
	input := &configservice.DescribeConfigRulesInput{
		ConfigRuleNames: []string{name},
	}

	return findConfigRule(ctx, conn, input)
}

func findConfigRule(ctx context.Context, conn *configservice.Client, input *configservice.DescribeConfigRulesInput) (*types.ConfigRule, error) {
	output, err := findConfigRules(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findConfigRules(ctx context.Context, conn *configservice.Client, input *configservice.DescribeConfigRulesInput) ([]types.ConfigRule, error) {
	var output []types.ConfigRule

	pages := configservice.NewDescribeConfigRulesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*types.NoSuchConfigRuleException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.ConfigRules...)
	}

	return output, nil
}

func statusConfigRule(ctx context.Context, conn *configservice.Client, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findConfigRuleByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.ConfigRuleState), nil
	}
}

func waitConfigRuleDeleted(ctx context.Context, conn *configservice.Client, name string) (*types.ConfigRule, error) {
	const (
		timeout = 5 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(
			types.ConfigRuleStateActive,
			types.ConfigRuleStateDeleting,
			types.ConfigRuleStateDeletingResults,
			types.ConfigRuleStateEvaluating,
		),
		Target:  []string{},
		Refresh: statusConfigRule(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*types.ConfigRule); ok {
		return v, err
	}

	return nil, err
}

func expandEvaluationModeConfigurations(tfList []interface{}) []types.EvaluationModeConfiguration {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []types.EvaluationModeConfiguration

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		apiObject := types.EvaluationModeConfiguration{}

		if v, ok := tfMap[names.AttrMode].(string); ok && v != "" {
			apiObject.Mode = types.EvaluationMode(v)
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandScope(tfMap map[string]interface{}) *types.Scope {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.Scope{}

	if v, ok := tfMap["compliance_resource_id"].(string); ok && v != "" {
		apiObject.ComplianceResourceId = aws.String(v)
	}

	if v, ok := tfMap["compliance_resource_types"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.ComplianceResourceTypes = flex.ExpandStringValueSet(v)
	}

	if v, ok := tfMap["tag_key"].(string); ok && v != "" {
		apiObject.TagKey = aws.String(v)
	}

	if v, ok := tfMap["tag_value"].(string); ok && v != "" {
		apiObject.TagValue = aws.String(v)
	}

	return apiObject
}

func expandSource(tfMap map[string]interface{}) *types.Source {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.Source{
		Owner: types.Owner(tfMap[names.AttrOwner].(string)),
	}

	if v, ok := tfMap["custom_policy_details"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.CustomPolicyDetails = expandCustomPolicyDetails(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["source_detail"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.SourceDetails = expandSourceDetails(v.List())
	}

	if v, ok := tfMap["source_identifier"].(string); ok && v != "" {
		apiObject.SourceIdentifier = aws.String(v)
	}

	return apiObject
}

func expandCustomPolicyDetails(tfMap map[string]interface{}) *types.CustomPolicyDetails {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.CustomPolicyDetails{
		EnableDebugLogDelivery: tfMap["enable_debug_log_delivery"].(bool),
		PolicyRuntime:          aws.String(tfMap["policy_runtime"].(string)),
		PolicyText:             aws.String(tfMap["policy_text"].(string)),
	}

	return apiObject
}

func expandSourceDetails(tfList []interface{}) []types.SourceDetail {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []types.SourceDetail

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		apiObject := types.SourceDetail{}

		if v, ok := tfMap["event_source"].(string); ok && v != "" {
			apiObject.EventSource = types.EventSource(v)
		}

		if v, ok := tfMap["maximum_execution_frequency"].(string); ok && v != "" {
			apiObject.MaximumExecutionFrequency = types.MaximumExecutionFrequency(v)
		}

		if v, ok := tfMap["message_type"].(string); ok && v != "" {
			apiObject.MessageType = types.MessageType(v)
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenEvaluationModeConfigurations(apiObjects []types.EvaluationModeConfiguration) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfMap := map[string]interface{}{
			names.AttrMode: apiObject.Mode,
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenScope(apiObject *types.Scope) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.ComplianceResourceId != nil {
		tfMap["compliance_resource_id"] = aws.ToString(apiObject.ComplianceResourceId)
	}

	if apiObject.ComplianceResourceTypes != nil {
		tfMap["compliance_resource_types"] = apiObject.ComplianceResourceTypes
	}

	if apiObject.TagKey != nil {
		tfMap["tag_key"] = aws.ToString(apiObject.TagKey)
	}

	if apiObject.TagValue != nil {
		tfMap["tag_value"] = aws.ToString(apiObject.TagValue)
	}

	return []interface{}{tfMap}
}

func flattenSource(apiObject *types.Source) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		names.AttrOwner:     apiObject.Owner,
		"source_identifier": aws.ToString(apiObject.SourceIdentifier),
	}

	if apiObject.CustomPolicyDetails != nil {
		tfMap["custom_policy_details"] = flattenCustomPolicyDetails(apiObject.CustomPolicyDetails)
	}

	if len(apiObject.SourceDetails) > 0 {
		tfMap["source_detail"] = flattenSourceDetails(apiObject.SourceDetails)
	}

	return []interface{}{tfMap}
}

func flattenCustomPolicyDetails(apiObject *types.CustomPolicyDetails) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"enable_debug_log_delivery": apiObject.EnableDebugLogDelivery,
		"policy_runtime":            aws.ToString(apiObject.PolicyRuntime),
		"policy_text":               aws.ToString(apiObject.PolicyText),
	}

	return []interface{}{tfMap}
}

func flattenSourceDetails(apiObjects []types.SourceDetail) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfMap := map[string]interface{}{
			"event_source":                apiObject.EventSource,
			"maximum_execution_frequency": apiObject.MaximumExecutionFrequency,
			"message_type":                apiObject.MessageType,
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}
