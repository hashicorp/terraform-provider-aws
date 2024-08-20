// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ce

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_ce_cost_category", name="Cost Category")
func dataSourceCostCategory() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceCostCategoryRead,

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				"cost_category_arn": {
					Type:     schema.TypeString,
					Required: true,
				},
				names.AttrDefaultValue: {
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
				names.AttrName: {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrRule: {
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
							names.AttrRule: {
								Type:     schema.TypeList,
								Computed: true,
								Elem:     sdkv2.DataSourceElemFromResourceElem(expressionElem(costCategoryRuleRootElementSchemaLevel)),
							},
							names.AttrType: {
								Type:     schema.TypeString,
								Computed: true,
							},
							names.AttrValue: {
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
							names.AttrParameter: {
								Type:     schema.TypeSet,
								Computed: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										names.AttrType: {
											Type:     schema.TypeString,
											Computed: true,
										},
										names.AttrValues: {
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
							names.AttrSource: {
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
				names.AttrTags: tftags.TagsSchemaComputed(),
			}
		},
	}
}

func dataSourceCostCategoryRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CEClient(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	arn := d.Get("cost_category_arn").(string)
	costCategory, err := findCostCategoryByARN(ctx, conn, arn)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Cost Explorer Cost Category (%s): %s", arn, err)
	}

	d.SetId(aws.ToString(costCategory.CostCategoryArn))
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

	tags, err := listTags(ctx, conn, d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing Cost Explorer Cost Category (%s) tags: %s", d.Id(), err)
	}

	if err := d.Set(names.AttrTags, tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting split_charge_rule: %s", err)
	}

	return diags
}
