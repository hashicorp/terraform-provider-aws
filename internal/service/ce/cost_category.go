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
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	costCategoryRuleRootElementSchemaLevel = 3
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
				names.AttrARN: {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrDefaultValue: {
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
				names.AttrName: {
					Type:         schema.TypeString,
					Required:     true,
					ForceNew:     true,
					ValidateFunc: validation.StringLenBetween(1, 50),
				},
				names.AttrRule: {
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
							names.AttrRule: {
								Type:     schema.TypeList,
								MaxItems: 1,
								Optional: true,
								Elem:     expressionElem(costCategoryRuleRootElementSchemaLevel),
							},
							names.AttrType: {
								Type:             schema.TypeString,
								Optional:         true,
								ValidateDiagFunc: enum.Validate[awstypes.CostCategoryRuleType](),
							},
							names.AttrValue: {
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
							names.AttrParameter: {
								Type:     schema.TypeSet,
								Optional: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										names.AttrType: {
											Type:             schema.TypeString,
											Optional:         true,
											ValidateDiagFunc: enum.Validate[awstypes.CostCategorySplitChargeRuleParameterType](),
										},
										names.AttrValues: {
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
							names.AttrSource: {
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

func expressionElem(level int) *schema.Resource {
	// This is the non-recursive part of the schema.
	expressionSchema := map[string]*schema.Schema{
		"cost_category": {
			Type:     schema.TypeList,
			MaxItems: 1,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					names.AttrKey: {
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
					names.AttrValues: {
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
					names.AttrKey: {
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
					names.AttrValues: {
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
		names.AttrTags: {
			Type:     schema.TypeList,
			MaxItems: 1,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					names.AttrKey: {
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
					names.AttrValues: {
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
	}

	if level > 1 {
		// Add in the recursive part of the schema.
		expressionSchema["and"] = &schema.Schema{
			Type:     schema.TypeSet,
			Optional: true,
			Elem:     expressionElem(level - 1),
		}
		expressionSchema["not"] = &schema.Schema{
			Type:     schema.TypeList,
			MaxItems: 1,
			Optional: true,
			Elem:     expressionElem(level - 1),
		}
		expressionSchema["or"] = &schema.Schema{
			Type:     schema.TypeSet,
			Optional: true,
			Elem:     expressionElem(level - 1),
		}
	}

	return &schema.Resource{
		Schema: expressionSchema,
	}
}

func resourceCostCategoryCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CEClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &costexplorer.CreateCostCategoryDefinitionInput{
		Name:         aws.String(name),
		ResourceTags: getTagsIn(ctx),
		Rules:        expandCostCategoryRules(d.Get(names.AttrRule).(*schema.Set).List()),
		RuleVersion:  awstypes.CostCategoryRuleVersion(d.Get("rule_version").(string)),
	}

	if v, ok := d.GetOk(names.AttrDefaultValue); ok {
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

	d.Set(names.AttrARN, costCategory.CostCategoryArn)
	d.Set(names.AttrDefaultValue, costCategory.DefaultValue)
	d.Set("effective_end", costCategory.EffectiveEnd)
	d.Set("effective_start", costCategory.EffectiveStart)
	d.Set(names.AttrName, costCategory.Name)
	if err = d.Set(names.AttrRule, flattenCostCategoryRules(costCategory.Rules)); err != nil {
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

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &costexplorer.UpdateCostCategoryDefinitionInput{
			CostCategoryArn: aws.String(d.Id()),
			EffectiveStart:  aws.String(d.Get("effective_start").(string)),
			Rules:           expandCostCategoryRules(d.Get(names.AttrRule).(*schema.Set).List()),
			RuleVersion:     awstypes.CostCategoryRuleVersion(d.Get("rule_version").(string)),
		}

		if d.HasChange(names.AttrDefaultValue) {
			input.DefaultValue = aws.String(d.Get(names.AttrDefaultValue).(string))
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

func expandCostCategoryRule(tfMap map[string]interface{}) *awstypes.CostCategoryRule {
	apiObject := &awstypes.CostCategoryRule{}

	if v, ok := tfMap["inherited_value"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.InheritedValue = expandCostCategoryInheritedValueDimension(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap[names.AttrRule].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.Rule = expandExpression(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap[names.AttrType].(string); ok && v != "" {
		apiObject.Type = awstypes.CostCategoryRuleType(v)
	}

	if v, ok := tfMap[names.AttrValue].(string); ok && v != "" {
		apiObject.Value = aws.String(v)
	}

	return apiObject
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

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
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

func expandExpression(tfMap map[string]interface{}) *awstypes.Expression {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.Expression{}

	if v, ok := tfMap["and"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.And = expandExpressions(v.List())
	}

	if v, ok := tfMap["cost_category"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.CostCategories = expandCostCategoryValues(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["dimension"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.Dimensions = expandDimensionValues(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["not"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.Not = expandExpression(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["or"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.Or = expandExpressions(v.List())
	}

	if v, ok := tfMap[names.AttrTags].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.Tags = expandTagValues(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandExpressions(tfList []interface{}) []awstypes.Expression {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.Expression

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		apiObject := expandExpression(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandCostCategoryValues(tfMap map[string]interface{}) *awstypes.CostCategoryValues {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.CostCategoryValues{}

	if v, ok := tfMap[names.AttrKey].(string); ok && v != "" {
		apiObject.Key = aws.String(v)
	}

	if v, ok := tfMap["match_options"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.MatchOptions = flex.ExpandStringyValueSet[awstypes.MatchOption](v)
	}

	if v, ok := tfMap[names.AttrValues].(*schema.Set); ok && v.Len() > 0 {
		apiObject.Values = flex.ExpandStringValueSet(v)
	}

	return apiObject
}

func expandDimensionValues(tfMap map[string]interface{}) *awstypes.DimensionValues {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.DimensionValues{}

	if v, ok := tfMap[names.AttrKey].(string); ok && v != "" {
		apiObject.Key = awstypes.Dimension(v)
	}

	if v, ok := tfMap["match_options"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.MatchOptions = flex.ExpandStringyValueSet[awstypes.MatchOption](v)
	}

	if v, ok := tfMap[names.AttrValues].(*schema.Set); ok && v.Len() > 0 {
		apiObject.Values = flex.ExpandStringValueSet(v)
	}

	return apiObject
}

func expandTagValues(tfMap map[string]interface{}) *awstypes.TagValues {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.TagValues{}

	if v, ok := tfMap[names.AttrKey].(string); ok && v != "" {
		apiObject.Key = aws.String(v)
	}

	if v, ok := tfMap["match_options"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.MatchOptions = flex.ExpandStringyValueSet[awstypes.MatchOption](v)
	}

	if v, ok := tfMap[names.AttrValues].(*schema.Set); ok && v.Len() > 0 {
		apiObject.Values = flex.ExpandStringValueSet(v)
	}

	return apiObject
}

func expandCostCategorySplitChargeRule(tfMap map[string]interface{}) *awstypes.CostCategorySplitChargeRule {
	apiObject := &awstypes.CostCategorySplitChargeRule{
		Method:  awstypes.CostCategorySplitChargeMethod(tfMap["method"].(string)),
		Source:  aws.String(tfMap[names.AttrSource].(string)),
		Targets: flex.ExpandStringValueSet(tfMap["targets"].(*schema.Set)),
	}
	if v, ok := tfMap[names.AttrParameter]; ok {
		apiObject.Parameters = expandCostCategorySplitChargeRuleParameters(v.(*schema.Set).List())
	}

	return apiObject
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

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandCostCategorySplitChargeRuleParameter(tfMap map[string]interface{}) *awstypes.CostCategorySplitChargeRuleParameter {
	apiObject := &awstypes.CostCategorySplitChargeRuleParameter{
		Type:   awstypes.CostCategorySplitChargeRuleParameterType(tfMap[names.AttrType].(string)),
		Values: flex.ExpandStringValueList(tfMap[names.AttrValues].([]interface{})),
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

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func flattenCostCategoryRule(apiObject *awstypes.CostCategoryRule) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	tfMap["inherited_value"] = flattenCostCategoryInheritedValueDimension(apiObject.InheritedValue)

	if v := apiObject.Rule; v != nil {
		tfMap[names.AttrRule] = []interface{}{flattenExpression(apiObject.Rule)}
	}

	tfMap[names.AttrType] = string(apiObject.Type)

	if v := apiObject.Value; v != nil {
		tfMap[names.AttrValue] = aws.ToString(v)
	}

	return tfMap
}

func flattenCostCategoryRules(apiObjects []awstypes.CostCategoryRule) []map[string]interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []map[string]interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenCostCategoryRule(&apiObject))
	}

	return tfList
}

func flattenCostCategoryInheritedValueDimension(apiObject *awstypes.CostCategoryInheritedValueDimension) []map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	var tfList []map[string]interface{}
	tfMap := map[string]interface{}{}

	if v := apiObject.DimensionKey; v != nil {
		tfMap["dimension_key"] = aws.ToString(v)
	}

	tfMap["dimension_name"] = string(apiObject.DimensionName)

	tfList = append(tfList, tfMap)

	return tfList
}

func flattenExpression(apiObject *awstypes.Expression) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if len(apiObject.And) > 0 {
		tfMap["and"] = flattenExpressions(apiObject.And)
	}
	tfMap["cost_category"] = flattenCostCategoryValues(apiObject.CostCategories)
	tfMap["dimension"] = flattenDimensionValues(apiObject.Dimensions)
	if apiObject.Not != nil {
		tfMap["not"] = []interface{}{flattenExpression(apiObject.Not)}
	}
	if len(apiObject.Or) > 0 {
		tfMap["or"] = flattenExpressions(apiObject.Or)
	}
	tfMap[names.AttrTags] = flattenTagValues(apiObject.Tags)

	return tfMap
}

func flattenExpressions(apiObjects []awstypes.Expression) []map[string]interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []map[string]interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenExpression(&apiObject))
	}

	return tfList
}

func flattenCostCategoryValues(apiObject *awstypes.CostCategoryValues) []map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	var tfList []map[string]interface{}
	tfMap := map[string]interface{}{}

	if v := apiObject.Key; v != nil {
		tfMap[names.AttrKey] = aws.ToString(v)
	}

	tfMap["match_options"] = flex.FlattenStringyValueList(apiObject.MatchOptions)
	tfMap[names.AttrValues] = apiObject.Values

	tfList = append(tfList, tfMap)

	return tfList
}

func flattenDimensionValues(apiObject *awstypes.DimensionValues) []map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	var tfList []map[string]interface{}
	tfMap := map[string]interface{}{}

	tfMap[names.AttrKey] = string(apiObject.Key)
	tfMap["match_options"] = flex.FlattenStringyValueList(apiObject.MatchOptions)
	tfMap[names.AttrValues] = apiObject.Values

	tfList = append(tfList, tfMap)

	return tfList
}

func flattenTagValues(apiObject *awstypes.TagValues) []map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	var tfList []map[string]interface{}
	tfMap := map[string]interface{}{}

	if v := apiObject.Key; v != nil {
		tfMap[names.AttrKey] = aws.ToString(v)
	}

	tfMap["match_options"] = flex.FlattenStringyValueList(apiObject.MatchOptions)
	tfMap[names.AttrValues] = apiObject.Values

	tfList = append(tfList, tfMap)

	return tfList
}

func flattenCostCategorySplitChargeRule(apiObject *awstypes.CostCategorySplitChargeRule) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	tfMap["method"] = string(apiObject.Method)
	tfMap[names.AttrParameter] = flattenCostCategorySplitChargeRuleParameters(apiObject.Parameters)
	tfMap[names.AttrSource] = aws.ToString(apiObject.Source)
	tfMap["targets"] = apiObject.Targets

	return tfMap
}

func flattenCostCategorySplitChargeRules(apiObjects []awstypes.CostCategorySplitChargeRule) []map[string]interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []map[string]interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenCostCategorySplitChargeRule(&apiObject))
	}

	return tfList
}

func flattenCostCategorySplitChargeRuleParameter(apiObject *awstypes.CostCategorySplitChargeRuleParameter) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	tfMap[names.AttrType] = string(apiObject.Type)
	tfMap[names.AttrValues] = apiObject.Values

	return tfMap
}

func flattenCostCategorySplitChargeRuleParameters(apiObjects []awstypes.CostCategorySplitChargeRuleParameter) []map[string]interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []map[string]interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenCostCategorySplitChargeRuleParameter(&apiObject))
	}

	return tfList
}
