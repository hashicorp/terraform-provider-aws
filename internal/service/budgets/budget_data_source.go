// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package budgets

import (
	"cmp"
	"context"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/budgets/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_budgets_budget", name="Budget")
// @Tags(identifierAttribute="arn")
func dataSourceBudget() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceBudgetRead,

		Schema: map[string]*schema.Schema{
			names.AttrAccountID: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"auto_adjust_data": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"auto_adjust_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"historical_options": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"budget_adjustment_period": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"lookback_available_periods": {
										Type:     schema.TypeInt,
										Computed: true,
									},
								},
							},
						},
						"last_auto_adjust_time": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"billing_view_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"budget_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"budget_limit": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"amount": {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrUnit: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"calculated_spend": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"actual_spend": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"amount": {
										Type:     schema.TypeString,
										Computed: true,
									},
									names.AttrUnit: {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
			"cost_filter": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrName: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrValues: {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
			"cost_types": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"include_credit": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"include_discount": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"include_other_subscription": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"include_recurring": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"include_refund": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"include_subscription": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"include_support": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"include_tax": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"include_upfront": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"use_amortized": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"use_blended": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrNamePrefix: {
				Type:     schema.TypeString,
				Optional: true,
			},
			"notification": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"comparison_operator": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"notification_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"subscriber_email_addresses": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"subscriber_sns_topic_arns": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"threshold": {
							Type:     schema.TypeFloat,
							Computed: true,
						},
						"threshold_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"planned_limit": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"amount": {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrStartTime: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrUnit: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
			"time_period_end": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"time_period_start": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"time_unit": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"budget_exceeded": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func dataSourceBudgetRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	c := meta.(*conns.AWSClient)
	conn := c.BudgetsClient(ctx)

	accountID := cmp.Or(d.Get(names.AttrAccountID).(string), c.AccountID(ctx))
	budgetName := create.Name(ctx, d.Get(names.AttrName).(string), d.Get(names.AttrNamePrefix).(string))

	budget, err := findBudgetByTwoPartKey(ctx, conn, accountID, budgetName)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Budget (%s): %s", d.Id(), err)
	}

	d.SetId(budgetCreateResourceID(accountID, budgetName))
	d.Set(names.AttrAccountID, accountID)
	d.Set(names.AttrARN, budgetARN(ctx, c, accountID, budgetName))
	d.Set("billing_view_arn", budget.BillingViewArn)
	d.Set("budget_exceeded", false)
	if budget.CalculatedSpend != nil && budget.CalculatedSpend.ActualSpend != nil {
		if aws.ToString(budget.BudgetLimit.Unit) == aws.ToString(budget.CalculatedSpend.ActualSpend.Unit) {
			bLimit, err := strconv.ParseFloat(aws.ToString(budget.BudgetLimit.Amount), 64)
			if err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}
			bSpend, err := strconv.ParseFloat(aws.ToString(budget.CalculatedSpend.ActualSpend.Amount), 64)
			if err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}

			if bLimit < bSpend {
				d.Set("budget_exceeded", true)
			} else {
				d.Set("budget_exceeded", false)
			}
		}
	}
	if err := d.Set("budget_limit", flattenSpend(budget.BudgetLimit)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting budget_spend: %s", err)
	}
	d.Set("budget_type", budget.BudgetType)
	if err := d.Set("calculated_spend", flattenCalculatedSpend(budget.CalculatedSpend)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting calculated_spend: %s", err)
	}

	d.Set(names.AttrName, budget.BudgetName)
	d.Set(names.AttrNamePrefix, create.NamePrefixFromName(aws.ToString(budget.BudgetName)))

	return diags
}

func flattenCalculatedSpend(apiObject *awstypes.CalculatedSpend) []any {
	if apiObject == nil {
		return nil
	}

	attrs := map[string]any{
		"actual_spend": flattenSpend(apiObject.ActualSpend),
	}
	return []any{attrs}
}

func flattenSpend(apiObject *awstypes.Spend) []any {
	if apiObject == nil {
		return nil
	}

	attrs := map[string]any{
		"amount":       aws.ToString(apiObject.Amount),
		names.AttrUnit: aws.ToString(apiObject.Unit),
	}

	return []any{attrs}
}
