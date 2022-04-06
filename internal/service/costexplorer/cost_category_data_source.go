package costexplorer

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/costexplorer"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func DataSourceCostExplorerCostCategory() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceCostExplorerCostCategoryRead,
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
							Elem:     schemaCostExplorerCostCategoryRuleComputed(),
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
							Set: costExplorerCostCategorySplitChargesParameter,
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
				Set: costExplorerCostCategorySplitCharges,
			},
		},
	}
}

func schemaCostExplorerCostCategoryRuleComputed() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"and": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     schemaCostExplorerCostCategoryRuleExpressionComputed(),
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
				Elem:     schemaCostExplorerCostCategoryRuleExpressionComputed(),
			},
			"or": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     schemaCostExplorerCostCategoryRuleExpressionComputed(),
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

func schemaCostExplorerCostCategoryRuleExpressionComputed() *schema.Resource {
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

func dataSourceCostExplorerCostCategoryRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).CostExplorerConn

	resp, err := conn.DescribeCostCategoryDefinitionWithContext(ctx, &costexplorer.DescribeCostCategoryDefinitionInput{CostCategoryArn: aws.String(d.Get("cost_category_arn").(string))})

	if err != nil {
		return diag.Errorf("error reading CostExplorer Cost Category Definition (%s): %w", d.Id(), err)
	}

	d.Set("default_value", resp.CostCategory.DefaultValue)
	d.Set("effective_end", resp.CostCategory.EffectiveEnd)
	d.Set("effective_start", resp.CostCategory.EffectiveStart)
	d.Set("name", resp.CostCategory.Name)
	if err = d.Set("rule", flattenCostExplorerCostCategoryRules(resp.CostCategory.Rules)); err != nil {
		return diag.Errorf("error setting `%s` for CostExplorer Cost Category Definition (%s): %w", "rule", d.Id(), err)
	}
	d.Set("rule_version", resp.CostCategory.RuleVersion)
	if err = d.Set("split_charge_rule", flattenCostExplorerCostCategorySplitChargeRules(resp.CostCategory.SplitChargeRules)); err != nil {
		return diag.Errorf("error setting `%s` for CostExplorer Cost Category Definition (%s): %w", "split_charge_rule", d.Id(), err)
	}

	d.SetId(aws.StringValue(resp.CostCategory.CostCategoryArn))

	return nil
}
