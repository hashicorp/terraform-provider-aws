package ce

import (
	"bytes"
	"context"
	"fmt"

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
	"github.com/hashicorp/terraform-provider-aws/names"
)

func ResourceCostCategory() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCostCategoryCreate,
		ReadContext:   resourceCostCategoryRead,
		UpdateContext: resourceCostCategoryUpdate,
		DeleteContext: resourceCostCategoryDelete,
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
							Elem:     schemaCostCategoryRule(),
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

func schemaCostCategoryRule() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"and": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     schemaCostCategoryRuleExpression(),
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
				Elem:     schemaCostCategoryRuleExpression(),
			},
			"or": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     schemaCostCategoryRuleExpression(),
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

func schemaCostCategoryRuleExpression() *schema.Resource {
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

func resourceCostCategoryCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).CEConn
	input := &costexplorer.CreateCostCategoryDefinitionInput{
		Name:        aws.String(d.Get("name").(string)),
		Rules:       expandCostCategoryRules(d.Get("rule").(*schema.Set).List()),
		RuleVersion: aws.String(d.Get("rule_version").(string)),
	}

	if v, ok := d.GetOk("default_value"); ok {
		input.DefaultValue = aws.String(v.(string))
	}

	if v, ok := d.GetOk("split_charge_rule"); ok {
		input.SplitChargeRules = expandCostCategorySplitChargeRules(v.(*schema.Set).List())
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
		return names.DiagError(names.CE, names.ErrActionCreating, ResCostCategory, d.Id(), err)
	}

	d.SetId(aws.StringValue(output.CostCategoryArn))

	return resourceCostCategoryRead(ctx, d, meta)
}

func resourceCostCategoryRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).CEConn

	resp, err := conn.DescribeCostCategoryDefinitionWithContext(ctx, &costexplorer.DescribeCostCategoryDefinitionInput{CostCategoryArn: aws.String(d.Id())})
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, costexplorer.ErrCodeResourceNotFoundException) {
		names.LogNotFoundRemoveState(names.CE, names.ErrActionReading, ResCostCategory, d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return names.DiagError(names.CE, names.ErrActionReading, ResCostCategory, d.Id(), err)
	}

	d.Set("arn", resp.CostCategory.CostCategoryArn)
	d.Set("default_value", resp.CostCategory.DefaultValue)
	d.Set("effective_end", resp.CostCategory.EffectiveEnd)
	d.Set("effective_start", resp.CostCategory.EffectiveStart)
	d.Set("name", resp.CostCategory.Name)
	if err = d.Set("rule", flattenCostCategoryRules(resp.CostCategory.Rules)); err != nil {
		return names.DiagError(names.CE, "setting rule", ResCostCategory, d.Id(), err)
	}
	d.Set("rule_version", resp.CostCategory.RuleVersion)
	if err = d.Set("split_charge_rule", flattenCostCategorySplitChargeRules(resp.CostCategory.SplitChargeRules)); err != nil {
		return names.DiagError(names.CE, "setting split_charge_rule", ResCostCategory, d.Id(), err)
	}

	return nil
}

func resourceCostCategoryUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).CEConn

	input := &costexplorer.UpdateCostCategoryDefinitionInput{
		CostCategoryArn: aws.String(d.Id()),
		Rules:           expandCostCategoryRules(d.Get("rule").(*schema.Set).List()),
		RuleVersion:     aws.String(d.Get("rule_version").(string)),
	}

	if d.HasChange("default_value") {
		input.DefaultValue = aws.String(d.Get("default_value").(string))
	}

	if d.HasChange("split_charge_rule") {
		input.SplitChargeRules = expandCostCategorySplitChargeRules(d.Get("split_charge_rule").(*schema.Set).List())
	}

	_, err := conn.UpdateCostCategoryDefinitionWithContext(ctx, input)

	if err != nil {
		return names.DiagError(names.CE, names.ErrActionUpdating, ResCostCategory, d.Id(), err)
	}

	return resourceCostCategoryRead(ctx, d, meta)
}

func resourceCostCategoryDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).CEConn

	_, err := conn.DeleteCostCategoryDefinitionWithContext(ctx, &costexplorer.DeleteCostCategoryDefinitionInput{
		CostCategoryArn: aws.String(d.Id()),
	})
	if err != nil && tfawserr.ErrCodeEquals(err, costexplorer.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return names.DiagError(names.CE, names.ErrActionDeleting, ResCostCategory, d.Id(), err)
	}

	return nil
}

func expandCostCategoryRule(tfMap map[string]interface{}) *costexplorer.CostCategoryRule {
	if tfMap == nil {
		return nil
	}

	apiObject := &costexplorer.CostCategoryRule{}
	if v, ok := tfMap["inherited_value"]; ok {
		apiObject.InheritedValue = expandCostCategoryInheritedValue(v.([]interface{}))
	}
	if v, ok := tfMap["rule"]; ok {
		apiObject.Rule = expandCostExpressions(v.([]interface{}))[0]
	}
	if v, ok := tfMap["type"]; ok {
		apiObject.Type = aws.String(v.(string))
	}
	if v, ok := tfMap["value"]; ok {
		apiObject.Value = aws.String(v.(string))
	}

	return apiObject
}

func expandCostCategoryInheritedValue(tfList []interface{}) *costexplorer.CostCategoryInheritedValueDimension {
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

func expandCostExpression(tfMap map[string]interface{}) *costexplorer.Expression {
	if tfMap == nil {
		return nil
	}

	apiObject := &costexplorer.Expression{}
	if v, ok := tfMap["and"]; ok {
		apiObject.And = expandCostExpressions(v.(*schema.Set).List())
	}
	if v, ok := tfMap["cost_category"]; ok {
		apiObject.CostCategories = expandCostExpressionCostCategory(v.([]interface{}))
	}
	if v, ok := tfMap["dimension"]; ok {
		apiObject.Dimensions = expandCostExpressionDimension(v.([]interface{}))
	}
	if v, ok := tfMap["not"]; ok && len(v.([]interface{})) > 0 {
		apiObject.Not = expandCostExpressions(v.([]interface{}))[0]
	}
	if v, ok := tfMap["or"]; ok {
		apiObject.Or = expandCostExpressions(v.(*schema.Set).List())
	}
	if v, ok := tfMap["tags"]; ok {
		apiObject.Tags = expandCostExpressionTag(v.([]interface{}))
	}

	return apiObject
}

func expandCostExpressionCostCategory(tfList []interface{}) *costexplorer.CostCategoryValues {
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

func expandCostExpressionDimension(tfList []interface{}) *costexplorer.DimensionValues {
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

func expandCostExpressionTag(tfList []interface{}) *costexplorer.TagValues {
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

func expandCostExpressions(tfList []interface{}) []*costexplorer.Expression {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*costexplorer.Expression

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandCostExpression(tfMap)

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandCostCategoryRules(tfList []interface{}) []*costexplorer.CostCategoryRule {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*costexplorer.CostCategoryRule

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

func expandCostCategorySplitChargeRule(tfMap map[string]interface{}) *costexplorer.CostCategorySplitChargeRule {
	if tfMap == nil {
		return nil
	}

	apiObject := &costexplorer.CostCategorySplitChargeRule{
		Method:  aws.String(tfMap["method"].(string)),
		Source:  aws.String(tfMap["source"].(string)),
		Targets: flex.ExpandStringSet(tfMap["targets"].(*schema.Set)),
	}
	if v, ok := tfMap["parameter"]; ok {
		apiObject.Parameters = expandCostCategorySplitChargeRuleParameters(v.(*schema.Set).List())
	}

	return apiObject
}

func expandCostCategorySplitChargeRuleParameter(tfMap map[string]interface{}) *costexplorer.CostCategorySplitChargeRuleParameter {
	if tfMap == nil {
		return nil
	}

	apiObject := &costexplorer.CostCategorySplitChargeRuleParameter{
		Type:   aws.String(tfMap["method"].(string)),
		Values: flex.ExpandStringSet(tfMap["values"].(*schema.Set)),
	}

	return apiObject
}

func expandCostCategorySplitChargeRuleParameters(tfList []interface{}) []*costexplorer.CostCategorySplitChargeRuleParameter {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*costexplorer.CostCategorySplitChargeRuleParameter

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

func expandCostCategorySplitChargeRules(tfList []interface{}) []*costexplorer.CostCategorySplitChargeRule {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*costexplorer.CostCategorySplitChargeRule

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

func flattenCostCategoryRule(apiObject *costexplorer.CostCategoryRule) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	var expressions []*costexplorer.Expression
	expressions = append(expressions, apiObject.Rule)

	tfMap["inherited_value"] = flattenCostCategoryRuleInheritedValue(apiObject.InheritedValue)
	tfMap["rule"] = flattenCostCategoryRuleExpressions(expressions)
	tfMap["type"] = aws.StringValue(apiObject.Type)
	tfMap["value"] = aws.StringValue(apiObject.Value)

	return tfMap
}

func flattenCostCategoryRuleInheritedValue(apiObject *costexplorer.CostCategoryInheritedValueDimension) []map[string]interface{} {
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

func flattenCostCategoryRuleExpression(apiObject *costexplorer.Expression) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	tfMap["and"] = flattenCostCategoryRuleOperandExpressions(apiObject.And)
	tfMap["cost_category"] = flattenCostCategoryRuleExpressionCostCategory(apiObject.CostCategories)
	tfMap["dimension"] = flattenCostCategoryRuleExpressionDimension(apiObject.Dimensions)
	tfMap["not"] = flattenCostCategoryRuleOperandExpressions([]*costexplorer.Expression{apiObject.Not})
	tfMap["or"] = flattenCostCategoryRuleOperandExpressions(apiObject.Or)
	tfMap["tags"] = flattenCostCategoryRuleExpressionTag(apiObject.Tags)

	return tfMap
}

func flattenCostCategoryRuleExpressionCostCategory(apiObject *costexplorer.CostCategoryValues) []map[string]interface{} {
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

func flattenCostCategoryRuleExpressionDimension(apiObject *costexplorer.DimensionValues) []map[string]interface{} {
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

func flattenCostCategoryRuleExpressionTag(apiObject *costexplorer.TagValues) []map[string]interface{} {
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

func flattenCostCategoryRuleExpressions(apiObjects []*costexplorer.Expression) []map[string]interface{} {
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

func flattenCostCategoryRuleOperandExpression(apiObject *costexplorer.Expression) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	tfMap["cost_category"] = flattenCostCategoryRuleExpressionCostCategory(apiObject.CostCategories)
	tfMap["dimension"] = flattenCostCategoryRuleExpressionDimension(apiObject.Dimensions)
	tfMap["tags"] = flattenCostCategoryRuleExpressionTag(apiObject.Tags)

	return tfMap
}

func flattenCostCategoryRuleOperandExpressions(apiObjects []*costexplorer.Expression) []map[string]interface{} {
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

func flattenCostCategoryRules(apiObjects []*costexplorer.CostCategoryRule) []map[string]interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []map[string]interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenCostCategoryRule(apiObject))
	}

	return tfList
}

func flattenCostCategorySplitChargeRule(apiObject *costexplorer.CostCategorySplitChargeRule) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	tfMap["method"] = aws.StringValue(apiObject.Method)
	tfMap["parameter"] = flattenCostCategorySplitChargeRuleParameters(apiObject.Parameters)
	tfMap["source"] = aws.StringValue(apiObject.Source)
	tfMap["targets"] = flex.FlattenStringList(apiObject.Targets)

	return tfMap
}

func flattenCostCategorySplitChargeRuleParameter(apiObject *costexplorer.CostCategorySplitChargeRuleParameter) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	tfMap["type"] = aws.StringValue(apiObject.Type)
	tfMap["values"] = flex.FlattenStringList(apiObject.Values)

	return tfMap
}

func flattenCostCategorySplitChargeRuleParameters(apiObjects []*costexplorer.CostCategorySplitChargeRuleParameter) []map[string]interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []map[string]interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenCostCategorySplitChargeRuleParameter(apiObject))
	}

	return tfList
}

func flattenCostCategorySplitChargeRules(apiObjects []*costexplorer.CostCategorySplitChargeRule) []map[string]interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []map[string]interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenCostCategorySplitChargeRule(apiObject))
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
