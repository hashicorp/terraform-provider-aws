package ce

import (
	"bytes"
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/costexplorer"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceCECostCategory() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCECostCategoryCreate,
		ReadContext:   resourceCECostCategoryRead,
		UpdateContext: resourceCECostCategoryUpdate,
		DeleteContext: resourceCECostCategoryDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
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
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice(costexplorer.CostCategoryInheritedValueDimensionName_Values(), false),
									},
								},
							},
						},
						"rule": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Elem:     schemaCECostCategoryRule(),
						},
						"type": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(costexplorer.CostCategoryRuleType_Values(), false),
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
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(costexplorer.CostCategorySplitChargeMethod_Values(), false),
						},
						"parameter": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"type": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice(costexplorer.CostCategorySplitChargeRuleParameterType_Values(), false),
									},
									"values": {
										Type:     schema.TypeSet,
										Optional: true,
										MinItems: 1,
										MaxItems: 500,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: validation.StringLenBetween(0, 1024),
										},
										Set: schema.HashString,
									},
								},
							},
							Set: costExplorerCostCategorySplitChargesParameter,
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
							Set: schema.HashString,
						},
					},
				},
				Set: costExplorerCostCategorySplitCharges,
			},
		},
	}
}

func schemaCECostCategoryRule() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"and": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     schemaCECostCategoryRuleExpression(),
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
								Type:         schema.TypeString,
								ValidateFunc: validation.StringInSlice(costexplorer.MatchOption_Values(), false),
							},
							Set: schema.HashString,
						},
						"values": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringLenBetween(0, 1024),
							},
							Set: schema.HashString,
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
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(costexplorer.Dimension_Values(), false),
						},
						"match_options": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringInSlice(costexplorer.MatchOption_Values(), false),
							},
							Set: schema.HashString,
						},
						"values": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringLenBetween(0, 1024),
							},
							Set: schema.HashString,
						},
					},
				},
			},
			"not": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem:     schemaCECostCategoryRuleExpression(),
			},
			"or": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     schemaCECostCategoryRuleExpression(),
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
								Type:         schema.TypeString,
								ValidateFunc: validation.StringInSlice(costexplorer.MatchOption_Values(), false),
							},
							Set: schema.HashString,
						},
						"values": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringLenBetween(0, 1024),
							},
							Set: schema.HashString,
						},
					},
				},
			},
		},
	}
}

func schemaCECostCategoryRuleExpression() *schema.Resource {
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
								Type:         schema.TypeString,
								ValidateFunc: validation.StringInSlice(costexplorer.MatchOption_Values(), false),
							},
							Set: schema.HashString,
						},
						"values": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringLenBetween(0, 1024),
							},
							Set: schema.HashString,
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
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(costexplorer.Dimension_Values(), false),
						},
						"match_options": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringInSlice(costexplorer.MatchOption_Values(), false),
							},
							Set: schema.HashString,
						},
						"values": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringLenBetween(0, 1024),
							},
							Set: schema.HashString,
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
								Type:         schema.TypeString,
								ValidateFunc: validation.StringInSlice(costexplorer.MatchOption_Values(), false),
							},
							Set: schema.HashString,
						},
						"values": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringLenBetween(0, 1024),
							},
							Set: schema.HashString,
						},
					},
				},
			},
		},
	}
}

func resourceCECostCategoryCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).CEConn
	input := &costexplorer.CreateCostCategoryDefinitionInput{
		Name:        aws.String(d.Get("name").(string)),
		Rules:       expandCECostCategoryRules(d.Get("rule").(*schema.Set).List()),
		RuleVersion: aws.String(d.Get("rule_version").(string)),
	}

	if v, ok := d.GetOk("default_value"); ok {
		input.DefaultValue = aws.String(v.(string))
	}

	if v, ok := d.GetOk("split_charge_rule"); ok {
		input.SplitChargeRules = expandCECostCategorySplitChargeRules(v.(*schema.Set).List())
	}

	var err error
	var output *costexplorer.CreateCostCategoryDefinitionOutput
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		output, err = conn.CreateCostCategoryDefinition(input)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, costexplorer.ErrCodeResourceNotFoundException) {
				return resource.RetryableError(err)
			}

			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		output, err = conn.CreateCostCategoryDefinition(input)
	}

	if err != nil {
		return diag.Errorf("error creating CE Cost Category Definition (%s): %s", d.Id(), err)
	}

	d.SetId(aws.StringValue(output.CostCategoryArn))

	return resourceCECostCategoryRead(ctx, d, meta)
}

func resourceCECostCategoryRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).CEConn

	resp, err := conn.DescribeCostCategoryDefinitionWithContext(ctx, &costexplorer.DescribeCostCategoryDefinitionInput{CostCategoryArn: aws.String(d.Id())})
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, costexplorer.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] CE Cost Category Definition (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("error reading CE Cost Category Definition (%s): %s", d.Id(), err)
	}

	d.Set("arn", resp.CostCategory.CostCategoryArn)
	d.Set("default_value", resp.CostCategory.DefaultValue)
	d.Set("effective_end", resp.CostCategory.EffectiveEnd)
	d.Set("effective_start", resp.CostCategory.EffectiveStart)
	d.Set("name", resp.CostCategory.Name)
	if err = d.Set("rule", flattenCECostCategoryRules(resp.CostCategory.Rules)); err != nil {
		return diag.Errorf("error setting `%s` for CE Cost Category Definition (%s): %s", "rule", d.Id(), err)
	}
	d.Set("rule_version", resp.CostCategory.RuleVersion)
	if err = d.Set("split_charge_rule", flattenCECostCategorySplitChargeRules(resp.CostCategory.SplitChargeRules)); err != nil {
		return diag.Errorf("error setting `%s` for CE Cost Category Definition (%s): %s", "split_charge_rule", d.Id(), err)
	}

	return nil
}

func resourceCECostCategoryUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).CEConn

	input := &costexplorer.UpdateCostCategoryDefinitionInput{
		CostCategoryArn: aws.String(d.Id()),
		Rules:           expandCECostCategoryRules(d.Get("rule").(*schema.Set).List()),
		RuleVersion:     aws.String(d.Get("rule_version").(string)),
	}

	if d.HasChange("default_value") {
		input.DefaultValue = aws.String(d.Get("default_value").(string))
	}

	if d.HasChange("split_charge_rule") {
		input.SplitChargeRules = expandCECostCategorySplitChargeRules(d.Get("split_charge_rule").(*schema.Set).List())
	}

	_, err := conn.UpdateCostCategoryDefinitionWithContext(ctx, input)

	if err != nil {
		diag.Errorf("error updating CE Cost Category Definition (%s): %s", d.Id(), err)
	}

	return resourceCECostCategoryRead(ctx, d, meta)
}

func resourceCECostCategoryDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).CEConn

	_, err := conn.DeleteCostCategoryDefinitionWithContext(ctx, &costexplorer.DeleteCostCategoryDefinitionInput{
		CostCategoryArn: aws.String(d.Id()),
	})
	if err != nil {
		if tfawserr.ErrCodeEquals(err, costexplorer.ErrCodeResourceNotFoundException) {
			return nil
		}
		return diag.Errorf("error deleting CE Cost Category Definition (%s): %s", d.Id(), err)
	}

	return nil
}

func expandCECostCategoryRule(tfMap map[string]interface{}) *costexplorer.CostCategoryRule {
	if tfMap == nil {
		return nil
	}

	apiObject := &costexplorer.CostCategoryRule{}
	if v, ok := tfMap["inherited_value"]; ok {
		apiObject.InheritedValue = expandCECostCategoryInheritedValue(v.([]interface{}))
	}
	if v, ok := tfMap["rule"]; ok {
		apiObject.Rule = expandCECostExpressions(v.([]interface{}))[0]
	}
	if v, ok := tfMap["type"]; ok {
		apiObject.Type = aws.String(v.(string))
	}
	if v, ok := tfMap["value"]; ok {
		apiObject.Value = aws.String(v.(string))
	}

	return apiObject
}

func expandCECostCategoryInheritedValue(tfList []interface{}) *costexplorer.CostCategoryInheritedValueDimension {
	if len(tfList) == 0 {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})

	apiObject := &costexplorer.CostCategoryInheritedValueDimension{}
	if v, ok := tfMap["dimension_key"]; ok {
		apiObject.DimensionKey = aws.String(v.(string))
	}
	if v, ok := tfMap["dimension_name"]; ok {
		apiObject.DimensionName = aws.String(v.(string))
	}

	return apiObject
}

func expandCECostExpression(tfMap map[string]interface{}) *costexplorer.Expression {
	if tfMap == nil {
		return nil
	}

	apiObject := &costexplorer.Expression{}
	if v, ok := tfMap["and"]; ok {
		apiObject.And = expandCECostExpressions(v.(*schema.Set).List())
	}
	if v, ok := tfMap["cost_category"]; ok {
		apiObject.CostCategories = expandCECostExpressionCostCategory(v.([]interface{}))
	}
	if v, ok := tfMap["dimension"]; ok {
		apiObject.Dimensions = expandCECostExpressionDimension(v.([]interface{}))
	}
	if v, ok := tfMap["not"]; ok && len(v.([]interface{})) > 0 {
		apiObject.Not = expandCECostExpressions(v.([]interface{}))[0]
	}
	if v, ok := tfMap["or"]; ok {
		apiObject.Or = expandCECostExpressions(v.(*schema.Set).List())
	}
	if v, ok := tfMap["tags"]; ok {
		apiObject.Tags = expandCECostExpressionTag(v.([]interface{}))
	}

	return apiObject
}

func expandCECostExpressionCostCategory(tfList []interface{}) *costexplorer.CostCategoryValues {
	if len(tfList) == 0 {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})

	apiObject := &costexplorer.CostCategoryValues{}
	if v, ok := tfMap["key"]; ok {
		apiObject.Key = aws.String(v.(string))
	}
	if v, ok := tfMap["match_options"]; ok {
		apiObject.MatchOptions = flex.ExpandStringSet(v.(*schema.Set))
	}
	if v, ok := tfMap["values"]; ok {
		apiObject.Values = flex.ExpandStringSet(v.(*schema.Set))
	}

	return apiObject
}

func expandCECostExpressionDimension(tfList []interface{}) *costexplorer.DimensionValues {
	if len(tfList) == 0 {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})

	apiObject := &costexplorer.DimensionValues{}
	if v, ok := tfMap["key"]; ok {
		apiObject.Key = aws.String(v.(string))
	}
	if v, ok := tfMap["match_options"]; ok {
		apiObject.MatchOptions = flex.ExpandStringSet(v.(*schema.Set))
	}
	if v, ok := tfMap["values"]; ok {
		apiObject.Values = flex.ExpandStringSet(v.(*schema.Set))
	}

	return apiObject
}

func expandCECostExpressionTag(tfList []interface{}) *costexplorer.TagValues {
	if len(tfList) == 0 {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})

	apiObject := &costexplorer.TagValues{}
	if v, ok := tfMap["key"]; ok {
		apiObject.Key = aws.String(v.(string))
	}
	if v, ok := tfMap["match_options"]; ok {
		apiObject.MatchOptions = flex.ExpandStringSet(v.(*schema.Set))
	}
	if v, ok := tfMap["values"]; ok {
		apiObject.Values = flex.ExpandStringSet(v.(*schema.Set))
	}

	return apiObject
}

func expandCECostExpressions(tfList []interface{}) []*costexplorer.Expression {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*costexplorer.Expression

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandCECostExpression(tfMap)

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandCECostCategoryRules(tfList []interface{}) []*costexplorer.CostCategoryRule {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*costexplorer.CostCategoryRule

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandCECostCategoryRule(tfMap)

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandCECostCategorySplitChargeRule(tfMap map[string]interface{}) *costexplorer.CostCategorySplitChargeRule {
	if tfMap == nil {
		return nil
	}

	apiObject := &costexplorer.CostCategorySplitChargeRule{
		Method:  aws.String(tfMap["method"].(string)),
		Source:  aws.String(tfMap["source"].(string)),
		Targets: flex.ExpandStringSet(tfMap["targets"].(*schema.Set)),
	}
	if v, ok := tfMap["parameter"]; ok {
		apiObject.Parameters = expandCECostCategorySplitChargeRuleParameters(v.(*schema.Set).List())
	}

	return apiObject
}

func expandCECostCategorySplitChargeRuleParameter(tfMap map[string]interface{}) *costexplorer.CostCategorySplitChargeRuleParameter {
	if tfMap == nil {
		return nil
	}

	apiObject := &costexplorer.CostCategorySplitChargeRuleParameter{
		Type:   aws.String(tfMap["method"].(string)),
		Values: flex.ExpandStringSet(tfMap["values"].(*schema.Set)),
	}

	return apiObject
}

func expandCECostCategorySplitChargeRuleParameters(tfList []interface{}) []*costexplorer.CostCategorySplitChargeRuleParameter {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*costexplorer.CostCategorySplitChargeRuleParameter

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandCECostCategorySplitChargeRuleParameter(tfMap)

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandCECostCategorySplitChargeRules(tfList []interface{}) []*costexplorer.CostCategorySplitChargeRule {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*costexplorer.CostCategorySplitChargeRule

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandCECostCategorySplitChargeRule(tfMap)

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenCECostCategoryRule(apiObject *costexplorer.CostCategoryRule) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	var expressions []*costexplorer.Expression
	expressions = append(expressions, apiObject.Rule)

	tfMap["inherited_value"] = flattenCECostCategoryRuleInheritedValue(apiObject.InheritedValue)
	tfMap["rule"] = flattenCECostCategoryRuleExpressions(expressions)
	tfMap["type"] = aws.StringValue(apiObject.Type)
	tfMap["value"] = aws.StringValue(apiObject.Value)

	return tfMap
}

func flattenCECostCategoryRuleInheritedValue(apiObject *costexplorer.CostCategoryInheritedValueDimension) []map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	var tfList []map[string]interface{}
	tfMap := map[string]interface{}{}

	tfMap["dimension_key"] = aws.StringValue(apiObject.DimensionKey)
	tfMap["dimension_name"] = aws.StringValue(apiObject.DimensionName)

	tfList = append(tfList, tfMap)

	return tfList
}

func flattenCECostCategoryRuleExpression(apiObject *costexplorer.Expression) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	tfMap["and"] = flattenCECostCategoryRuleOperandExpressions(apiObject.And)
	tfMap["cost_category"] = flattenCECostCategoryRuleExpressionCostCategory(apiObject.CostCategories)
	tfMap["dimension"] = flattenCECostCategoryRuleExpressionDimension(apiObject.Dimensions)
	tfMap["not"] = flattenCECostCategoryRuleOperandExpressions([]*costexplorer.Expression{apiObject.Not})
	tfMap["or"] = flattenCECostCategoryRuleOperandExpressions(apiObject.Or)
	tfMap["tags"] = flattenCECostCategoryRuleExpressionTag(apiObject.Tags)

	return tfMap
}

func flattenCECostCategoryRuleExpressionCostCategory(apiObject *costexplorer.CostCategoryValues) []map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	var tfList []map[string]interface{}
	tfMap := map[string]interface{}{}

	tfMap["key"] = aws.StringValue(apiObject.Key)
	tfMap["match_options"] = flex.FlattenStringList(apiObject.MatchOptions)
	tfMap["values"] = flex.FlattenStringList(apiObject.Values)

	tfList = append(tfList, tfMap)

	return tfList
}

func flattenCECostCategoryRuleExpressionDimension(apiObject *costexplorer.DimensionValues) []map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	var tfList []map[string]interface{}
	tfMap := map[string]interface{}{}

	tfMap["key"] = aws.StringValue(apiObject.Key)
	tfMap["match_options"] = flex.FlattenStringList(apiObject.MatchOptions)
	tfMap["values"] = flex.FlattenStringList(apiObject.Values)

	tfList = append(tfList, tfMap)

	return tfList
}

func flattenCECostCategoryRuleExpressionTag(apiObject *costexplorer.TagValues) []map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	var tfList []map[string]interface{}
	tfMap := map[string]interface{}{}

	tfMap["key"] = aws.StringValue(apiObject.Key)
	tfMap["match_options"] = flex.FlattenStringList(apiObject.MatchOptions)
	tfMap["values"] = flex.FlattenStringList(apiObject.Values)

	tfList = append(tfList, tfMap)

	return tfList
}

func flattenCECostCategoryRuleExpressions(apiObjects []*costexplorer.Expression) []map[string]interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []map[string]interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenCECostCategoryRuleExpression(apiObject))
	}

	return tfList
}

func flattenCECostCategoryRuleOperandExpression(apiObject *costexplorer.Expression) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	tfMap["cost_category"] = flattenCECostCategoryRuleExpressionCostCategory(apiObject.CostCategories)
	tfMap["dimension"] = flattenCECostCategoryRuleExpressionDimension(apiObject.Dimensions)
	tfMap["tags"] = flattenCECostCategoryRuleExpressionTag(apiObject.Tags)

	return tfMap
}

func flattenCECostCategoryRuleOperandExpressions(apiObjects []*costexplorer.Expression) []map[string]interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []map[string]interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenCECostCategoryRuleOperandExpression(apiObject))
	}

	return tfList
}

func flattenCECostCategoryRules(apiObjects []*costexplorer.CostCategoryRule) []map[string]interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []map[string]interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenCECostCategoryRule(apiObject))
	}

	return tfList
}

func flattenCECostCategorySplitChargeRule(apiObject *costexplorer.CostCategorySplitChargeRule) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	tfMap["method"] = aws.StringValue(apiObject.Method)
	tfMap["parameter"] = flattenCECostCategorySplitChargeRuleParameters(apiObject.Parameters)
	tfMap["source"] = aws.StringValue(apiObject.Source)
	tfMap["targets"] = flex.FlattenStringList(apiObject.Targets)

	return tfMap
}

func flattenCECostCategorySplitChargeRuleParameter(apiObject *costexplorer.CostCategorySplitChargeRuleParameter) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	tfMap["type"] = aws.StringValue(apiObject.Type)
	tfMap["values"] = flex.FlattenStringList(apiObject.Values)

	return tfMap
}

func flattenCECostCategorySplitChargeRuleParameters(apiObjects []*costexplorer.CostCategorySplitChargeRuleParameter) []map[string]interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []map[string]interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenCECostCategorySplitChargeRuleParameter(apiObject))
	}

	return tfList
}

func flattenCECostCategorySplitChargeRules(apiObjects []*costexplorer.CostCategorySplitChargeRule) []map[string]interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []map[string]interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenCECostCategorySplitChargeRule(apiObject))
	}

	return tfList
}

func costExplorerCostCategorySplitCharges(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(m["method"].(string))
	buf.WriteString(fmt.Sprintf("%+v", m["parameter"].(*schema.Set)))
	buf.WriteString(m["source"].(string))
	buf.WriteString(fmt.Sprintf("%+v", m["targets"].(*schema.Set)))
	return schema.HashString(buf.String())
}

func costExplorerCostCategorySplitChargesParameter(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(m["type"].(string))
	buf.WriteString(fmt.Sprintf("%+v", m["values"].(*schema.Set)))
	return schema.HashString(buf.String())
}
