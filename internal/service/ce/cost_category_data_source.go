package ce

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/costexplorer"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func DataSourceCostCategory() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceCostCategoryRead,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"cost_category_arn": {
				Type:     schema.TypeString,
				Required: true,
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
				Type:     schema.TypeString,
				Computed: true,
			},
			"rule": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"inherited_value": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"dimension_key": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"dimension_name": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"rule": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     schemaCostCategoryRuleComputed(),
						},
						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"value": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"rule_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"split_charge_rule": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"method": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"parameter": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"type": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"values": {
										Type:     schema.TypeSet,
										Computed: true,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: validation.StringLenBetween(0, 1024),
										},
										Set: schema.HashString,
									},
								},
							},
							Set: costCategorySplitChargesParameter,
						},
						"source": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"targets": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringLenBetween(0, 1024),
							},
							Set: schema.HashString,
						},
					},
				},
				Set: costCategorySplitCharges,
			},
		},
	}
}

func schemaCostCategoryRuleComputed() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"and": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     schemaCostCategoryRuleExpressionComputed(),
			},
			"cost_category": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"match_options": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Set: schema.HashString,
						},
						"values": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Set: schema.HashString,
						},
					},
				},
			},
			"dimension": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"match_options": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Set: schema.HashString,
						},
						"values": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Set: schema.HashString,
						},
					},
				},
			},
			"not": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     schemaCostCategoryRuleExpressionComputed(),
			},
			"or": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     schemaCostCategoryRuleExpressionComputed(),
			},
			"tags": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"match_options": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Set: schema.HashString,
						},
						"values": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Set: schema.HashString,
						},
					},
				},
			},
		},
	}
}

func schemaCostCategoryRuleExpressionComputed() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"cost_category": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"match_options": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Set: schema.HashString,
						},
						"values": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Set: schema.HashString,
						},
					},
				},
			},
			"dimension": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"match_options": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"values": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Set: schema.HashString,
						},
					},
				},
			},
			"tags": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"match_options": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Set: schema.HashString,
						},
						"values": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Set: schema.HashString,
						},
					},
				},
			},
		},
	}
}

func dataSourceCostCategoryRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).CEConn

	resp, err := conn.DescribeCostCategoryDefinitionWithContext(ctx, &costexplorer.DescribeCostCategoryDefinitionInput{CostCategoryArn: aws.String(d.Get("cost_category_arn").(string))})

	if err != nil {
		return names.DiagError(names.CE, names.ErrActionReading, ResCostCategory, d.Id(), err)
	}

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

	d.SetId(aws.StringValue(resp.CostCategory.CostCategoryArn))

	return nil
}
