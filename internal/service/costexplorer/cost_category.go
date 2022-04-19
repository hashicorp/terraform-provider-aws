package costexplorer

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

func ResourceCostExplorerCostCategory() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCostExplorerCostCategoryCreate,
		ReadContext:   resourceCostExplorerCostCategoryRead,
		UpdateContext: resourceCostExplorerCostCategoryUpdate,
		DeleteContext: resourceCostExplorerCostCategoryDelete,
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
							Elem:     schemaCostExplorerCostCategoryRule(),
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

func schemaCostExplorerCostCategoryRule() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"and": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     schemaCostExplorerCostCategoryRuleExpression(),
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
				Elem:     schemaCostExplorerCostCategoryRuleExpression(),
			},
			"or": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     schemaCostExplorerCostCategoryRuleExpression(),
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

func schemaCostExplorerCostCategoryRuleExpression() *schema.Resource {
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

func resourceCostExplorerCostCategoryCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).CostExplorerConn
	input := &costexplorer.CreateCostCategoryDefinitionInput{
		Name:        aws.String(d.Get("name").(string)),
		Rules:       expandCostExplorerCostCategoryRules(d.Get("rule").(*schema.Set).List()),
		RuleVersion: aws.String(d.Get("rule_version").(string)),
	}

	if v, ok := d.GetOk("default_value"); ok {
		input.DefaultValue = aws.String(v.(string))
	}

	if v, ok := d.GetOk("split_charge_rule"); ok {
		input.SplitChargeRules = expandCostExplorerCostCategorySplitChargeRules(v.(*schema.Set).List())
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
		return diag.Errorf("error creating CostExplorer Cost Category Definition (%s): %s", d.Id(), err)
	}

	d.SetId(aws.StringValue(output.CostCategoryArn))

	return resourceCostExplorerCostCategoryRead(ctx, d, meta)
}

func resourceCostExplorerCostCategoryRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).CostExplorerConn

	resp, err := conn.DescribeCostCategoryDefinitionWithContext(ctx, &costexplorer.DescribeCostCategoryDefinitionInput{CostCategoryArn: aws.String(d.Id())})
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, costexplorer.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] CostExplorer Cost Category Definition (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("error reading CostExplorer Cost Category Definition (%s): %s", d.Id(), err)
	}

	d.Set("arn", resp.CostCategory.CostCategoryArn)
	d.Set("default_value", resp.CostCategory.DefaultValue)
	d.Set("effective_end", resp.CostCategory.EffectiveEnd)
	d.Set("effective_start", resp.CostCategory.EffectiveStart)
	d.Set("name", resp.CostCategory.Name)
	if err = d.Set("rule", flattenCostExplorerCostCategoryRules(resp.CostCategory.Rules)); err != nil {
		return diag.Errorf("error setting `%s` for CostExplorer Cost Category Definition (%s): %s", "rule", d.Id(), err)
	}
	d.Set("rule_version", resp.CostCategory.RuleVersion)
	if err = d.Set("split_charge_rule", flattenCostExplorerCostCategorySplitChargeRules(resp.CostCategory.SplitChargeRules)); err != nil {
		return diag.Errorf("error setting `%s` for CostExplorer Cost Category Definition (%s): %s", "split_charge_rule", d.Id(), err)
	}

	return nil
}

func resourceCostExplorerCostCategoryUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).CostExplorerConn

	input := &costexplorer.UpdateCostCategoryDefinitionInput{
		CostCategoryArn: aws.String(d.Id()),
		Rules:           expandCostExplorerCostCategoryRules(d.Get("rule").(*schema.Set).List()),
		RuleVersion:     aws.String(d.Get("rule_version").(string)),
	}

	if d.HasChange("default_value") {
		input.DefaultValue = aws.String(d.Get("default_value").(string))
	}

	if d.HasChange("split_charge_rule") {
		input.SplitChargeRules = expandCostExplorerCostCategorySplitChargeRules(d.Get("split_charge_rule").(*schema.Set).List())
	}

	_, err := conn.UpdateCostCategoryDefinitionWithContext(ctx, input)

	if err != nil {
		diag.Errorf("error updating CostExplorer Cost Category Definition (%s): %s", d.Id(), err)
	}

	return resourceCostExplorerCostCategoryRead(ctx, d, meta)
}

func resourceCostExplorerCostCategoryDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).CostExplorerConn

	_, err := conn.DeleteCostCategoryDefinitionWithContext(ctx, &costexplorer.DeleteCostCategoryDefinitionInput{
		CostCategoryArn: aws.String(d.Id()),
	})
	if err != nil {
		if tfawserr.ErrCodeEquals(err, costexplorer.ErrCodeResourceNotFoundException) {
			return nil
		}
		return diag.Errorf("error deleting CostExplorer Cost Category Definition (%s): %s", d.Id(), err)
	}

	return nil
}

func expandCostExplorerCostCategoryRule(tfMap map[string]interface{}) *costexplorer.CostCategoryRule {
	if tfMap == nil {
		return nil
	}

	apiObject := &costexplorer.CostCategoryRule{}
	if v, ok := tfMap["inherited_value"]; ok {
		apiObject.InheritedValue = expandCostExplorerCostCategoryInheritedValue(v.([]interface{}))
	}
	if v, ok := tfMap["rule"]; ok {
		apiObject.Rule = expandCostExplorerCostExpressions(v.([]interface{}))[0]
	}
	if v, ok := tfMap["type"]; ok {
		apiObject.Type = aws.String(v.(string))
	}
	if v, ok := tfMap["value"]; ok {
		apiObject.Value = aws.String(v.(string))
	}

	return apiObject
}

func expandCostExplorerCostCategoryInheritedValue(tfList []interface{}) *costexplorer.CostCategoryInheritedValueDimension {
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

func expandCostExplorerCostExpression(tfMap map[string]interface{}) *costexplorer.Expression {
	if tfMap == nil {
		return nil
	}

	apiObject := &costexplorer.Expression{}
	if v, ok := tfMap["and"]; ok {
		apiObject.And = expandCostExplorerCostExpressions(v.(*schema.Set).List())
	}
	if v, ok := tfMap["cost_category"]; ok {
		apiObject.CostCategories = expandCostExplorerCostExpressionCostCategory(v.([]interface{}))
	}
	if v, ok := tfMap["dimension"]; ok {
		apiObject.Dimensions = expandCostExplorerCostExpressionDimension(v.([]interface{}))
	}
	if v, ok := tfMap["not"]; ok && len(v.([]interface{})) > 0 {
		apiObject.Not = expandCostExplorerCostExpressions(v.([]interface{}))[0]
	}
	if v, ok := tfMap["or"]; ok {
		apiObject.Or = expandCostExplorerCostExpressions(v.(*schema.Set).List())
	}
	if v, ok := tfMap["tags"]; ok {
		apiObject.Tags = expandCostExplorerCostExpressionTag(v.([]interface{}))
	}

	return apiObject
}

func expandCostExplorerCostExpressionCostCategory(tfList []interface{}) *costexplorer.CostCategoryValues {
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

func expandCostExplorerCostExpressionDimension(tfList []interface{}) *costexplorer.DimensionValues {
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

func expandCostExplorerCostExpressionTag(tfList []interface{}) *costexplorer.TagValues {
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

func expandCostExplorerCostExpressions(tfList []interface{}) []*costexplorer.Expression {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*costexplorer.Expression

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandCostExplorerCostExpression(tfMap)

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandCostExplorerCostCategoryRules(tfList []interface{}) []*costexplorer.CostCategoryRule {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*costexplorer.CostCategoryRule

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandCostExplorerCostCategoryRule(tfMap)

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandCostExplorerCostCategorySplitChargeRule(tfMap map[string]interface{}) *costexplorer.CostCategorySplitChargeRule {
	if tfMap == nil {
		return nil
	}

	apiObject := &costexplorer.CostCategorySplitChargeRule{
		Method:  aws.String(tfMap["method"].(string)),
		Source:  aws.String(tfMap["source"].(string)),
		Targets: flex.ExpandStringSet(tfMap["targets"].(*schema.Set)),
	}
	if v, ok := tfMap["parameter"]; ok {
		apiObject.Parameters = expandCostExplorerCostCategorySplitChargeRuleParameters(v.(*schema.Set).List())
	}

	return apiObject
}

func expandCostExplorerCostCategorySplitChargeRuleParameter(tfMap map[string]interface{}) *costexplorer.CostCategorySplitChargeRuleParameter {
	if tfMap == nil {
		return nil
	}

	apiObject := &costexplorer.CostCategorySplitChargeRuleParameter{
		Type:   aws.String(tfMap["method"].(string)),
		Values: flex.ExpandStringSet(tfMap["values"].(*schema.Set)),
	}

	return apiObject
}

func expandCostExplorerCostCategorySplitChargeRuleParameters(tfList []interface{}) []*costexplorer.CostCategorySplitChargeRuleParameter {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*costexplorer.CostCategorySplitChargeRuleParameter

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandCostExplorerCostCategorySplitChargeRuleParameter(tfMap)

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandCostExplorerCostCategorySplitChargeRules(tfList []interface{}) []*costexplorer.CostCategorySplitChargeRule {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*costexplorer.CostCategorySplitChargeRule

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandCostExplorerCostCategorySplitChargeRule(tfMap)

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenCostExplorerCostCategoryRule(apiObject *costexplorer.CostCategoryRule) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	var expressions []*costexplorer.Expression
	expressions = append(expressions, apiObject.Rule)

	tfMap["inherited_value"] = flattenCostExplorerCostCategoryRuleInheritedValue(apiObject.InheritedValue)
	tfMap["rule"] = flattenCostExplorerCostCategoryRuleExpressions(expressions)
	tfMap["type"] = aws.StringValue(apiObject.Type)
	tfMap["value"] = aws.StringValue(apiObject.Value)

	return tfMap
}

func flattenCostExplorerCostCategoryRuleInheritedValue(apiObject *costexplorer.CostCategoryInheritedValueDimension) []map[string]interface{} {
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

func flattenCostExplorerCostCategoryRuleExpression(apiObject *costexplorer.Expression) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	tfMap["and"] = flattenCostExplorerCostCategoryRuleOperandExpressions(apiObject.And)
	tfMap["cost_category"] = flattenCostExplorerCostCategoryRuleExpressionCostCategory(apiObject.CostCategories)
	tfMap["dimension"] = flattenCostExplorerCostCategoryRuleExpressionDimension(apiObject.Dimensions)
	tfMap["not"] = flattenCostExplorerCostCategoryRuleOperandExpressions([]*costexplorer.Expression{apiObject.Not})
	tfMap["or"] = flattenCostExplorerCostCategoryRuleOperandExpressions(apiObject.Or)
	tfMap["tags"] = flattenCostExplorerCostCategoryRuleExpressionTag(apiObject.Tags)

	return tfMap
}

func flattenCostExplorerCostCategoryRuleExpressionCostCategory(apiObject *costexplorer.CostCategoryValues) []map[string]interface{} {
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

func flattenCostExplorerCostCategoryRuleExpressionDimension(apiObject *costexplorer.DimensionValues) []map[string]interface{} {
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

func flattenCostExplorerCostCategoryRuleExpressionTag(apiObject *costexplorer.TagValues) []map[string]interface{} {
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

func flattenCostExplorerCostCategoryRuleExpressions(apiObjects []*costexplorer.Expression) []map[string]interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []map[string]interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenCostExplorerCostCategoryRuleExpression(apiObject))
	}

	return tfList
}

func flattenCostExplorerCostCategoryRuleOperandExpression(apiObject *costexplorer.Expression) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	tfMap["cost_category"] = flattenCostExplorerCostCategoryRuleExpressionCostCategory(apiObject.CostCategories)
	tfMap["dimension"] = flattenCostExplorerCostCategoryRuleExpressionDimension(apiObject.Dimensions)
	tfMap["tags"] = flattenCostExplorerCostCategoryRuleExpressionTag(apiObject.Tags)

	return tfMap
}

func flattenCostExplorerCostCategoryRuleOperandExpressions(apiObjects []*costexplorer.Expression) []map[string]interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []map[string]interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenCostExplorerCostCategoryRuleOperandExpression(apiObject))
	}

	return tfList
}

func flattenCostExplorerCostCategoryRules(apiObjects []*costexplorer.CostCategoryRule) []map[string]interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []map[string]interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenCostExplorerCostCategoryRule(apiObject))
	}

	return tfList
}

func flattenCostExplorerCostCategorySplitChargeRule(apiObject *costexplorer.CostCategorySplitChargeRule) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	tfMap["method"] = aws.StringValue(apiObject.Method)
	tfMap["parameter"] = flattenCostExplorerCostCategorySplitChargeRuleParameters(apiObject.Parameters)
	tfMap["source"] = aws.StringValue(apiObject.Source)
	tfMap["targets"] = flex.FlattenStringList(apiObject.Targets)

	return tfMap
}

func flattenCostExplorerCostCategorySplitChargeRuleParameter(apiObject *costexplorer.CostCategorySplitChargeRuleParameter) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	tfMap["type"] = aws.StringValue(apiObject.Type)
	tfMap["values"] = flex.FlattenStringList(apiObject.Values)

	return tfMap
}

func flattenCostExplorerCostCategorySplitChargeRuleParameters(apiObjects []*costexplorer.CostCategorySplitChargeRuleParameter) []map[string]interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []map[string]interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenCostExplorerCostCategorySplitChargeRuleParameter(apiObject))
	}

	return tfList
}

func flattenCostExplorerCostCategorySplitChargeRules(apiObjects []*costexplorer.CostCategorySplitChargeRule) []map[string]interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []map[string]interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenCostExplorerCostCategorySplitChargeRule(apiObject))
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
