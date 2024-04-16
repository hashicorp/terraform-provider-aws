// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ce

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
	awstypes "github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ce_cost_category", name="Cost Category")
// @Tags(identifierAttribute="id")
func resourceCostCategory() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceCostCategoryCreate,
		ReadWithoutTimeout:   resourceCostCategoryRead,
		UpdateWithoutTimeout: resourceCostCategoryUpdate,
		DeleteWithoutTimeout: resourceCostCategoryDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: customdiff.Sequence(verify.SetTagsDiff),

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				"arn": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"default_value": {
					Type:         schema.TypeString,
					Optional:     true,
					ValidateFunc: validation.StringLenBetween(1, 50),
				},
				"effective_end": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"effective_start": {
					Type:     schema.TypeString,
					Optional: true,
					Computed: true,
				},
				"name": {
					Type:         schema.TypeString,
					Required:     true,
					ForceNew:     true,
					ValidateFunc: validation.StringLenBetween(1, 50),
				},
				"rule": {
					Type:     schema.TypeSet,
					Required: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"inherited_value": {
								Type:     schema.TypeList,
								MaxItems: 1,
								Optional: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"dimension_key": {
											Type:         schema.TypeString,
											Optional:     true,
											ValidateFunc: validation.StringLenBetween(0, 1024),
										},
										"dimension_name": {
											Type:             schema.TypeString,
											Optional:         true,
											ValidateDiagFunc: enum.Validate[awstypes.CostCategoryInheritedValueDimensionName](),
										},
									},
								},
							},
							"rule": {
								Type:     schema.TypeList,
								MaxItems: 1,
								Optional: true,
								Elem:     elemExpression(),
							},
							"type": {
								Type:             schema.TypeString,
								Optional:         true,
								ValidateDiagFunc: enum.Validate[awstypes.CostCategoryRuleType](),
							},
							"value": {
								Type:         schema.TypeString,
								Optional:     true,
								ValidateFunc: validation.StringLenBetween(1, 50),
							},
						},
					},
				},
				"rule_version": {
					Type:         schema.TypeString,
					Required:     true,
					ForceNew:     true,
					ValidateFunc: validation.StringLenBetween(0, 100),
				},
				"split_charge_rule": {
					Type:     schema.TypeSet,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"method": {
								Type:             schema.TypeString,
								Required:         true,
								ValidateDiagFunc: enum.Validate[awstypes.CostCategorySplitChargeMethod](),
							},
							"parameter": {
								Type:     schema.TypeSet,
								Optional: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"type": {
											Type:             schema.TypeString,
											Optional:         true,
											ValidateDiagFunc: enum.Validate[awstypes.CostCategorySplitChargeRuleParameterType](),
										},
										"values": {
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 1,
											MaxItems: 500,
											Elem: &schema.Schema{
												Type:         schema.TypeString,
												ValidateFunc: validation.StringLenBetween(0, 1024),
											},
										},
									},
								},
							},
							"source": {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: validation.StringLenBetween(0, 1024),
							},
							"targets": {
								Type:     schema.TypeSet,
								Required: true,
								MinItems: 1,
								MaxItems: 500,
								Elem: &schema.Schema{
									Type:         schema.TypeString,
									ValidateFunc: validation.StringLenBetween(0, 1024),
								},
							},
						},
					},
				},
				names.AttrTags:    tftags.TagsSchema(),
				names.AttrTagsAll: tftags.TagsSchemaComputed(),
			}
		},
	}
}

func elemExpression() *schema.Resource {
	elemNestedExpression := func() *schema.Resource {
		return &schema.Resource{
			Schema: map[string]*schema.Schema{
				"cost_category": {
					Type:     schema.TypeList,
					MaxItems: 1,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"key": {
								Type:         schema.TypeString,
								Optional:     true,
								ValidateFunc: validation.StringLenBetween(1, 50),
							},
							"match_options": {
								Type:     schema.TypeSet,
								Optional: true,
								Elem: &schema.Schema{
									Type:             schema.TypeString,
									ValidateDiagFunc: enum.Validate[awstypes.MatchOption](),
								},
							},
							"values": {
								Type:     schema.TypeSet,
								Optional: true,
								Elem: &schema.Schema{
									Type:         schema.TypeString,
									ValidateFunc: validation.StringLenBetween(0, 1024),
								},
							},
						},
					},
				},
				"dimension": {
					Type:     schema.TypeList,
					MaxItems: 1,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"key": {
								Type:             schema.TypeString,
								Optional:         true,
								ValidateDiagFunc: enum.Validate[awstypes.Dimension](),
							},
							"match_options": {
								Type:     schema.TypeSet,
								Optional: true,
								Elem: &schema.Schema{
									Type:             schema.TypeString,
									ValidateDiagFunc: enum.Validate[awstypes.MatchOption](),
								},
							},
							"values": {
								Type:     schema.TypeSet,
								Optional: true,
								Elem: &schema.Schema{
									Type:         schema.TypeString,
									ValidateFunc: validation.StringLenBetween(0, 1024),
								},
							},
						},
					},
				},
				"tags": {
					Type:     schema.TypeList,
					MaxItems: 1,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"key": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"match_options": {
								Type:     schema.TypeSet,
								Optional: true,
								Elem: &schema.Schema{
									Type:             schema.TypeString,
									ValidateDiagFunc: enum.Validate[awstypes.MatchOption](),
								},
							},
							"values": {
								Type:     schema.TypeSet,
								Optional: true,
								Elem: &schema.Schema{
									Type:         schema.TypeString,
									ValidateFunc: validation.StringLenBetween(0, 1024),
								},
							},
						},
					},
				},
			},
		}
	}

	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"and": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     elemNestedExpression(),
			},
			"cost_category": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(1, 50),
						},
						"match_options": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type:             schema.TypeString,
								ValidateDiagFunc: enum.Validate[awstypes.MatchOption](),
							},
						},
						"values": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringLenBetween(0, 1024),
							},
						},
					},
				},
			},
			"dimension": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[awstypes.Dimension](),
						},
						"match_options": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type:             schema.TypeString,
								ValidateDiagFunc: enum.Validate[awstypes.MatchOption](),
							},
						},
						"values": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringLenBetween(0, 1024),
							},
						},
					},
				},
			},
			"not": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem:     elemNestedExpression(),
			},
			"or": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     elemNestedExpression(),
			},
			"tags": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"match_options": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type:             schema.TypeString,
								ValidateDiagFunc: enum.Validate[awstypes.MatchOption](),
							},
						},
						"values": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringLenBetween(0, 1024),
							},
						},
					},
				},
			},
		},
	}
}

func resourceCostCategoryCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CEClient(ctx)

	name := d.Get("name").(string)
	input := &costexplorer.CreateCostCategoryDefinitionInput{
		Name:         aws.String(name),
		ResourceTags: getTagsIn(ctx),
		Rules:        expandCostCategoryRules(d.Get("rule").(*schema.Set).List()),
		RuleVersion:  awstypes.CostCategoryRuleVersion(d.Get("rule_version").(string)),
	}

	if v, ok := d.GetOk("default_value"); ok {
		input.DefaultValue = aws.String(v.(string))
	}

	if v, ok := d.GetOk("effective_start"); ok {
		input.EffectiveStart = aws.String(v.(string))
	}

	if v, ok := d.GetOk("split_charge_rule"); ok {
		input.SplitChargeRules = expandCostCategorySplitChargeRules(v.(*schema.Set).List())
	}

	outputRaw, err := tfresource.RetryWhenIsA[*awstypes.ResourceNotFoundException](ctx, d.Timeout(schema.TimeoutCreate),
		func() (interface{}, error) {
			return conn.CreateCostCategoryDefinition(ctx, input)
		})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Cost Explorer Cost Category (%s): %s", name, err)
	}

	d.SetId(aws.ToString(outputRaw.(*costexplorer.CreateCostCategoryDefinitionOutput).CostCategoryArn))

	return append(diags, resourceCostCategoryRead(ctx, d, meta)...)
}

func resourceCostCategoryRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CEClient(ctx)

	costCategory, err := findCostCategoryByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Cost Explorer Cost Category (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Cost Explorer Cost Category (%s): %s", d.Id(), err)
	}

	d.Set("arn", costCategory.CostCategoryArn)
	d.Set("default_value", costCategory.DefaultValue)
	d.Set("effective_end", costCategory.EffectiveEnd)
	d.Set("effective_start", costCategory.EffectiveStart)
	d.Set("name", costCategory.Name)
	if err = d.Set("rule", flattenCostCategoryRules(costCategory.Rules)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting rule: %s", err)
	}
	d.Set("rule_version", costCategory.RuleVersion)
	if err = d.Set("split_charge_rule", flattenCostCategorySplitChargeRules(costCategory.SplitChargeRules)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting split_charge_rule: %s", err)
	}

	return diags
}

func resourceCostCategoryUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CEClient(ctx)

	if d.HasChangesExcept("tags", "tags_all") {
		input := &costexplorer.UpdateCostCategoryDefinitionInput{
			CostCategoryArn: aws.String(d.Id()),
			EffectiveStart:  aws.String(d.Get("effective_start").(string)),
			Rules:           expandCostCategoryRules(d.Get("rule").(*schema.Set).List()),
			RuleVersion:     awstypes.CostCategoryRuleVersion(d.Get("rule_version").(string)),
		}

		if d.HasChange("default_value") {
			input.DefaultValue = aws.String(d.Get("default_value").(string))
		}

		if d.HasChange("split_charge_rule") {
			input.SplitChargeRules = expandCostCategorySplitChargeRules(d.Get("split_charge_rule").(*schema.Set).List())
		}

		_, err := conn.UpdateCostCategoryDefinition(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Cost Explorer Cost Category (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceCostCategoryRead(ctx, d, meta)...)
}

func resourceCostCategoryDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).CEClient(ctx)

	log.Printf("[DEBUG] Deleting Cost Explorer Cost Category: %s", d.Id())
	_, err := conn.DeleteCostCategoryDefinition(ctx, &costexplorer.DeleteCostCategoryDefinitionInput{
		CostCategoryArn: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Cost Explorer Cost Category (%s): %s", d.Id(), err)
	}

	return diags
}

func findCostCategoryByARN(ctx context.Context, conn *costexplorer.Client, arn string) (*awstypes.CostCategory, error) {
	input := &costexplorer.DescribeCostCategoryDefinitionInput{
		CostCategoryArn: aws.String(arn),
	}

	output, err := conn.DescribeCostCategoryDefinition(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.CostCategory == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.CostCategory, nil
}

func expandCostCategoryRule(tfMap map[string]interface{}) awstypes.CostCategoryRule {
	apiObject := awstypes.CostCategoryRule{}

	if v, ok := tfMap["inherited_value"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.InheritedValue = expandCostCategoryInheritedValueDimension(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["rule"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.Rule = expandCostExpression(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["type"].(string); ok && v != "" {
		apiObject.Type = awstypes.CostCategoryRuleType(v)
	}

	if v, ok := tfMap["value"].(string); ok && v != "" {
		apiObject.Value = aws.String(v)
	}

	return apiObject
}

func expandCostCategoryInheritedValueDimension(tfMap map[string]interface{}) *awstypes.CostCategoryInheritedValueDimension {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.CostCategoryInheritedValueDimension{}

	if v, ok := tfMap["dimension_key"].(string); ok && v != "" {
		apiObject.DimensionKey = aws.String(v)
	}

	if v, ok := tfMap["dimension_name"].(string); ok && v != "" {
		apiObject.DimensionName = awstypes.CostCategoryInheritedValueDimensionName(v)
	}

	return apiObject
}

func expandCostExpression(tfMap map[string]interface{}) *awstypes.Expression {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.Expression{}

	if v, ok := tfMap["and"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.And = expandCostExpressions(v.List())
	}

	if v, ok := tfMap["cost_category"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.CostCategories = expandCostCategoryValues(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["dimension"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.Dimensions = expandDimensionValues(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["not"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.Not = expandCostExpression(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["or"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.Or = expandCostExpressions(v.List())
	}

	if v, ok := tfMap["tags"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.Tags = expandTagValues(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandCostCategoryValues(tfMap map[string]interface{}) *awstypes.CostCategoryValues {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.CostCategoryValues{}

	if v, ok := tfMap["key"].(string); ok && v != "" {
		apiObject.Key = aws.String(v)
	}

	if v, ok := tfMap["match_options"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.MatchOptions = flex.ExpandStringyValueSet[awstypes.MatchOption](v)
	}

	if v, ok := tfMap["values"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.Values = flex.ExpandStringValueSet(v)
	}

	return apiObject
}

func expandDimensionValues(tfMap map[string]interface{}) *awstypes.DimensionValues {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.DimensionValues{}

	if v, ok := tfMap["key"].(string); ok && v != "" {
		apiObject.Key = awstypes.Dimension(v)
	}

	if v, ok := tfMap["match_options"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.MatchOptions = flex.ExpandStringyValueSet[awstypes.MatchOption](v)
	}

	if v, ok := tfMap["values"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.Values = flex.ExpandStringValueSet(v)
	}

	return apiObject
}

func expandTagValues(tfMap map[string]interface{}) *awstypes.TagValues {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.TagValues{}

	if v, ok := tfMap["key"].(string); ok && v != "" {
		apiObject.Key = aws.String(v)
	}

	if v, ok := tfMap["match_options"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.MatchOptions = flex.ExpandStringyValueSet[awstypes.MatchOption](v)
	}

	if v, ok := tfMap["values"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.Values = flex.ExpandStringValueSet(v)
	}

	return apiObject
}

func expandCostExpressions(tfList []interface{}) []awstypes.Expression {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.Expression

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		apiObject := expandCostExpression(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandCostCategoryRules(tfList []interface{}) []awstypes.CostCategoryRule {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.CostCategoryRule

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandCostCategoryRule(tfMap)

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandCostCategorySplitChargeRule(tfMap map[string]interface{}) awstypes.CostCategorySplitChargeRule {
	apiObject := awstypes.CostCategorySplitChargeRule{
		Method:  awstypes.CostCategorySplitChargeMethod(tfMap["method"].(string)),
		Source:  aws.String(tfMap["source"].(string)),
		Targets: flex.ExpandStringValueSet(tfMap["targets"].(*schema.Set)),
	}
	if v, ok := tfMap["parameter"]; ok {
		apiObject.Parameters = expandCostCategorySplitChargeRuleParameters(v.(*schema.Set).List())
	}

	return apiObject
}

func expandCostCategorySplitChargeRuleParameter(tfMap map[string]interface{}) awstypes.CostCategorySplitChargeRuleParameter {
	apiObject := awstypes.CostCategorySplitChargeRuleParameter{
		Type:   awstypes.CostCategorySplitChargeRuleParameterType(tfMap["type"].(string)),
		Values: flex.ExpandStringValueList(tfMap["values"].([]interface{})),
	}

	return apiObject
}

func expandCostCategorySplitChargeRuleParameters(tfList []interface{}) []awstypes.CostCategorySplitChargeRuleParameter {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.CostCategorySplitChargeRuleParameter

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandCostCategorySplitChargeRuleParameter(tfMap)

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandCostCategorySplitChargeRules(tfList []interface{}) []awstypes.CostCategorySplitChargeRule {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.CostCategorySplitChargeRule

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandCostCategorySplitChargeRule(tfMap)

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenCostCategoryRule(apiObject awstypes.CostCategoryRule) map[string]interface{} {
	tfMap := map[string]interface{}{}

	var expressions []*awstypes.Expression
	expressions = append(expressions, apiObject.Rule)

	tfMap["inherited_value"] = flattenCostCategoryRuleInheritedValue(apiObject.InheritedValue)
	tfMap["rule"] = flattenCostCategoryRuleExpressions(expressions)
	tfMap["type"] = string(apiObject.Type)
	tfMap["value"] = aws.ToString(apiObject.Value)

	return tfMap
}

func flattenCostCategoryRuleInheritedValue(apiObject *awstypes.CostCategoryInheritedValueDimension) []map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	var tfList []map[string]interface{}
	tfMap := map[string]interface{}{}

	tfMap["dimension_key"] = aws.ToString(apiObject.DimensionKey)
	tfMap["dimension_name"] = string(apiObject.DimensionName)

	tfList = append(tfList, tfMap)

	return tfList
}

func flattenCostCategoryRuleExpression(apiObject *awstypes.Expression) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	tfMap["and"] = flattenCostCategoryRuleOperandExpressions(tfslices.ToPointers[[]awstypes.Expression](apiObject.And))
	tfMap["cost_category"] = flattenCostCategoryRuleExpressionCostCategory(apiObject.CostCategories)
	tfMap["dimension"] = flattenCostCategoryRuleExpressionDimension(apiObject.Dimensions)
	tfMap["not"] = flattenCostCategoryRuleOperandExpressions([]*awstypes.Expression{apiObject.Not})
	tfMap["or"] = flattenCostCategoryRuleOperandExpressions(tfslices.ToPointers[[]awstypes.Expression](apiObject.Or))
	tfMap["tags"] = flattenCostCategoryRuleExpressionTag(apiObject.Tags)

	return tfMap
}

func flattenCostCategoryRuleExpressionCostCategory(apiObject *awstypes.CostCategoryValues) []map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	var tfList []map[string]interface{}
	tfMap := map[string]interface{}{}

	tfMap["key"] = aws.ToString(apiObject.Key)
	tfMap["match_options"] = flex.FlattenStringyValueList(apiObject.MatchOptions)
	tfMap["values"] = apiObject.Values

	tfList = append(tfList, tfMap)

	return tfList
}

func flattenCostCategoryRuleExpressionDimension(apiObject *awstypes.DimensionValues) []map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	var tfList []map[string]interface{}
	tfMap := map[string]interface{}{}

	tfMap["key"] = string(apiObject.Key)
	tfMap["match_options"] = flex.FlattenStringyValueList(apiObject.MatchOptions)
	tfMap["values"] = apiObject.Values

	tfList = append(tfList, tfMap)

	return tfList
}

func flattenCostCategoryRuleExpressionTag(apiObject *awstypes.TagValues) []map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	var tfList []map[string]interface{}
	tfMap := map[string]interface{}{}

	tfMap["key"] = aws.ToString(apiObject.Key)
	tfMap["match_options"] = flex.FlattenStringyValueList(apiObject.MatchOptions)
	tfMap["values"] = apiObject.Values

	tfList = append(tfList, tfMap)

	return tfList
}

func flattenCostCategoryRuleExpressions(apiObjects []*awstypes.Expression) []map[string]interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []map[string]interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenCostCategoryRuleExpression(apiObject))
	}

	return tfList
}

func flattenCostCategoryRuleOperandExpression(apiObject *awstypes.Expression) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	tfMap["cost_category"] = flattenCostCategoryRuleExpressionCostCategory(apiObject.CostCategories)
	tfMap["dimension"] = flattenCostCategoryRuleExpressionDimension(apiObject.Dimensions)
	tfMap["tags"] = flattenCostCategoryRuleExpressionTag(apiObject.Tags)

	return tfMap
}

func flattenCostCategoryRuleOperandExpressions(apiObjects []*awstypes.Expression) []map[string]interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []map[string]interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenCostCategoryRuleOperandExpression(apiObject))
	}

	return tfList
}

func flattenCostCategoryRules(apiObjects []awstypes.CostCategoryRule) []map[string]interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []map[string]interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenCostCategoryRule(apiObject))
	}

	return tfList
}

func flattenCostCategorySplitChargeRule(apiObject awstypes.CostCategorySplitChargeRule) map[string]interface{} {
	tfMap := map[string]interface{}{}
	tfMap["method"] = string(apiObject.Method)
	tfMap["parameter"] = flattenCostCategorySplitChargeRuleParameters(apiObject.Parameters)
	tfMap["source"] = aws.ToString(apiObject.Source)
	tfMap["targets"] = apiObject.Targets

	return tfMap
}

func flattenCostCategorySplitChargeRuleParameter(apiObject awstypes.CostCategorySplitChargeRuleParameter) map[string]interface{} {
	tfMap := map[string]interface{}{}
	tfMap["type"] = string(apiObject.Type)
	tfMap["values"] = apiObject.Values

	return tfMap
}

func flattenCostCategorySplitChargeRuleParameters(apiObjects []awstypes.CostCategorySplitChargeRuleParameter) []map[string]interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []map[string]interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenCostCategorySplitChargeRuleParameter(apiObject))
	}

	return tfList
}

func flattenCostCategorySplitChargeRules(apiObjects []awstypes.CostCategorySplitChargeRule) []map[string]interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []map[string]interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenCostCategorySplitChargeRule(apiObject))
	}

	return tfList
}
