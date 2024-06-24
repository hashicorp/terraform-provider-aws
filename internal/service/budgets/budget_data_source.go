// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package budgets

import (
	"context"
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	awstypes "github.com/aws/aws-sdk-go-v2/service/budgets/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_budgets_budget")
func DataSourceBudget() *schema.Resource {
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

const (
	DSNameBudget = "Budget Data Source"
)

func dataSourceBudgetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BudgetsClient(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	budgetName := create.Name(d.Get(names.AttrName).(string), d.Get(names.AttrNamePrefix).(string))

	accountID := d.Get(names.AttrAccountID).(string)
	if accountID == "" {
		accountID = meta.(*conns.AWSClient).AccountID
	}
	d.Set(names.AttrAccountID, accountID)

	budget, err := FindBudgetByTwoPartKey(ctx, conn, accountID, budgetName)
	if err != nil {
		return create.AppendDiagError(diags, names.Budgets, create.ErrActionReading, DSNameBudget, d.Id(), err)
	}

	d.SetId(fmt.Sprintf("%s:%s", accountID, budgetName))

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "budgets",
		AccountID: accountID,
		Resource:  "budget/" + budgetName,
	}
	d.Set(names.AttrARN, arn.String())

	d.Set("budget_type", budget.BudgetType)

	if err := d.Set("budget_limit", flattenSpend(budget.BudgetLimit)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting budget_spend: %s", err)
	}

	if err := d.Set("calculated_spend", flattenCalculatedSpend(budget.CalculatedSpend)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting calculated_spend: %s", err)
	}

	d.Set("budget_exceeded", false)
	if budget.CalculatedSpend != nil && budget.CalculatedSpend.ActualSpend != nil {
		if aws.ToString(budget.BudgetLimit.Unit) == aws.ToString(budget.CalculatedSpend.ActualSpend.Unit) {
			bLimit, err := strconv.ParseFloat(aws.ToString(budget.BudgetLimit.Amount), 64)
			if err != nil {
				return create.AppendDiagError(diags, names.Budgets, create.ErrActionReading, DSNameBudget, d.Id(), err)
			}
			bSpend, err := strconv.ParseFloat(aws.ToString(budget.CalculatedSpend.ActualSpend.Amount), 64)
			if err != nil {
				return create.AppendDiagError(diags, names.Budgets, create.ErrActionReading, DSNameBudget, d.Id(), err)
			}

			if bLimit < bSpend {
				d.Set("budget_exceeded", true)
			} else {
				d.Set("budget_exceeded", false)
			}
		}
	}

	d.Set(names.AttrName, budget.BudgetName)
	d.Set(names.AttrNamePrefix, create.NamePrefixFromName(aws.ToString(budget.BudgetName)))

	tags, err := listTags(ctx, conn, arn.String())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for Budget (%s): %s", d.Id(), err)
	}

	if err := d.Set(names.AttrTags, tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	return diags
}

func flattenCalculatedSpend(apiObject *awstypes.CalculatedSpend) []interface{} {
	if apiObject == nil {
		return nil
	}

	attrs := map[string]interface{}{
		"actual_spend": flattenSpend(apiObject.ActualSpend),
	}
	return []interface{}{attrs}
}

func flattenSpend(apiObject *awstypes.Spend) []interface{} {
	if apiObject == nil {
		return nil
	}

	attrs := map[string]interface{}{
		"amount":       aws.ToString(apiObject.Amount),
		names.AttrUnit: aws.ToString(apiObject.Unit),
	}

	return []interface{}{attrs}
}
