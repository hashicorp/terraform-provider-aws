// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ce

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_ce_cost_category")
func DataSourceCostCategory() *schema.Resource {
	schemaCostCategoryRuleExpressionComputed := func() *schema.Resource {
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
							},
							"values": {
								Type:     schema.TypeSet,
								Computed: true,
								Elem: &schema.Schema{
									Type: schema.TypeString,
								},
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
							},
							"values": {
								Type:     schema.TypeSet,
								Computed: true,
								Elem: &schema.Schema{
									Type: schema.TypeString,
								},
							},
						},
					},
				},
			},
		}
	}
	schemaCostCategoryRuleComputed := func() *schema.Resource {
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
							},
							"values": {
								Type:     schema.TypeSet,
								Computed: true,
								Elem: &schema.Schema{
									Type: schema.TypeString,
								},
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
							},
							"values": {
								Type:     schema.TypeSet,
								Computed: true,
								Elem: &schema.Schema{
									Type: schema.TypeString,
								},
							},
						},
					},
				},
			},
		}
	}

	return &schema.Resource{
		ReadWithoutTimeout: dataSourceCostCategoryRead,

		Schema: map[string]*schema.Schema{
			"cost_category_arn": {
				Type:     schema.TypeString,
				Required: true,
			},
			"default_value": {
				Type:     schema.TypeString,
				Computed: true,
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
									},
								},
							},
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
						},
					},
				},
			},
			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceCostCategoryRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).CEConn(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	costCategory, err := FindCostCategoryByARN(ctx, conn, d.Get("cost_category_arn").(string))

	if err != nil {
		return create.AppendDiagError(diags, names.CE, create.ErrActionReading, ResNameCostCategory, d.Id(), err)
	}

	d.Set("default_value", costCategory.DefaultValue)
	d.Set("effective_end", costCategory.EffectiveEnd)
	d.Set("effective_start", costCategory.EffectiveStart)
	d.Set("name", costCategory.Name)
	if err = d.Set("rule", flattenCostCategoryRules(costCategory.Rules)); err != nil {
		return create.AppendDiagError(diags, names.CE, "setting rule", ResNameCostCategory, d.Id(), err)
	}
	d.Set("rule_version", costCategory.RuleVersion)
	if err = d.Set("split_charge_rule", flattenCostCategorySplitChargeRules(costCategory.SplitChargeRules)); err != nil {
		return create.AppendDiagError(diags, names.CE, "setting split_charge_rule", ResNameCostCategory, d.Id(), err)
	}

	d.SetId(aws.StringValue(costCategory.CostCategoryArn))

	tags, err := listTags(ctx, conn, d.Id())

	if err != nil {
		return create.AppendDiagError(diags, names.CE, "listing tags", ResNameCostCategory, d.Id(), err)
	}

	if err := d.Set("tags", tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return create.AppendDiagError(diags, names.CE, "setting tags", ResNameCostCategory, d.Id(), err)
	}

	return diags
}
