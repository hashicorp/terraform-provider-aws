// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package budgets

import (
	"context"
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/budgets"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_budgets_budget")
func DataSourceBudget() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceBudgetRead,

		Schema: map[string]*schema.Schema{
			"account_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"arn": {
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
						"unit": {
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
									"unit": {
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
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"values": {
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
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name_prefix": {
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
						"start_time": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"unit": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
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

	conn := meta.(*conns.AWSClient).BudgetsConn(ctx)

	budgetName := create.Name(d.Get("name").(string), d.Get("name_prefix").(string))

	accountID := d.Get("account_id").(string)
	if accountID == "" {
		accountID = meta.(*conns.AWSClient).AccountID
	}
	d.Set("account_id", accountID)

	budget, err := FindBudgetByTwoPartKey(ctx, conn, accountID, budgetName)
	if err != nil {
		return create.AppendDiagError(diags, names.Budgets, create.ErrActionReading, DSNameBudget, d.Id(), err)
	}

	d.SetId(fmt.Sprintf("%s:%s", accountID, budgetName))

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "budgets",
		AccountID: accountID,
		Resource:  fmt.Sprintf("budget/%s", budgetName),
	}
	d.Set("arn", arn.String())

	d.Set("budget_type", budget.BudgetType)

	if err := d.Set("budget_limit", flattenSpend(budget.BudgetLimit)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting budget_spend: %s", err)
	}

	if err := d.Set("calculated_spend", flattenCalculatedSpend(budget.CalculatedSpend)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting calculated_spend: %s", err)
	}

	d.Set("budget_exceeded", false)
	if budget.CalculatedSpend != nil && budget.CalculatedSpend.ActualSpend != nil {
		if aws.StringValue(budget.BudgetLimit.Unit) == aws.StringValue(budget.CalculatedSpend.ActualSpend.Unit) {
			bLimit, err := strconv.ParseFloat(aws.StringValue(budget.BudgetLimit.Amount), 64)
			if err != nil {
				return create.AppendDiagError(diags, names.Budgets, create.ErrActionReading, DSNameBudget, d.Id(), err)
			}
			bSpend, err := strconv.ParseFloat(aws.StringValue(budget.CalculatedSpend.ActualSpend.Amount), 64)
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

	d.Set("name", budget.BudgetName)
	d.Set("name_prefix", create.NamePrefixFromName(aws.StringValue(budget.BudgetName)))

	return diags
}

func flattenCalculatedSpend(apiObject *budgets.CalculatedSpend) []interface{} {
	if apiObject == nil {
		return nil
	}

	attrs := map[string]interface{}{
		"actual_spend": flattenSpend(apiObject.ActualSpend),
	}
	return []interface{}{attrs}
}

func flattenSpend(apiObject *budgets.Spend) []interface{} {
	if apiObject == nil {
		return nil
	}

	attrs := map[string]interface{}{
		"amount": aws.StringValue(apiObject.Amount),
		"unit":   aws.StringValue(apiObject.Unit),
	}

	return []interface{}{attrs}
}
