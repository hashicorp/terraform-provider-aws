---
subcategory: "Web Services Budgets"
layout: "aws"
page_title: "AWS: aws_budgets_budget"
description: |-
  Terraform data source for managing an AWS Web Services Budgets Budget.
---

# Data Source: aws_budgets_budget

Terraform data source for managing an AWS Web Services Budgets Budget.

## Example Usage

### Basic Usage

```terraform
data "aws_budgets_budget" "test" {
  name = aws_budgets_budget.test.name
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the budget. Unique within an account.

The following arguments are optional:

* `account_id` - (Optional) ID of the target account for the budget. Defaults to the current account ID.
* `name_prefix` - (Optional) Prefix of the budget name. Unique within an account.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the budget.
* `auto_adjust_data` - Object that determines the budget amount for an auto-adjusting budget. See [`auto_adjust_data` Block](#auto_adjust_data-block) for details.
* `billing_view_arn` - ARN of the billing view.
* `budget_exceeded` - Whether the budget has been exceeded.
* `budget_limit` - Amount of cost, usage, RI utilization, RI coverage, Savings Plans utilization, or Savings Plans coverage tracked by the budget. See [`budget_limit` Block](#budget_limit-block) for details.
* `budget_type` - Whether the budget tracks monetary cost or usage.
* `calculated_spend` - Spend objects associated with the budget. See [`calculated_spend` Block](#calculated_spend-block) for details.
* `cost_filter` - Cost filters applied to the budget. See [`cost_filter` Block](#cost_filter-block) for details.
* `cost_types` - Types of cost included in the budget. See [`cost_types` Block](#cost_types-block) for details.
* `notification` - Notifications associated with the budget. See [`notification` Block](#notification-block) for details.
* `planned_limit` - Budget limits planned for future periods. See [`planned_limit` Block](#planned_limit-block) for details.
* `tags` - Map of tags assigned to the resource.
* `time_period_end` - End of the time period covered by the budget. Format: `2017-01-01_12:00`.
* `time_period_start` - Start of the time period covered by the budget. Format: `2017-01-01_12:00`.
* `time_unit` - Length of time until the budget resets the actual and forecasted spend. Valid values: `MONTHLY`, `QUARTERLY`, `ANNUALLY`, and `DAILY`.

### `auto_adjust_data` Block

* `auto_adjust_type` - String that defines whether the budget auto-adjusts based on historical or forecasted data. Valid values: `FORECAST`, `HISTORICAL`.
* `historical_options` - Historical data that the auto-adjusting budget is based on. See [`historical_options` Block](#historical_options-block) for details.
* `last_auto_adjust_time` - Last time that the budget was auto-adjusted.

### `historical_options` Block

* `budget_adjustment_period` - Number of budget periods included in the moving-average calculation that determines the auto-adjusted budget amount.
* `lookback_available_periods` - Number of budget periods in the `budget_adjustment_period` included in the calculation of the current budget limit.

### `budget_limit` Block

* `amount` - Cost or usage amount associated with the budget.
* `unit` - Unit of measurement used for the budget, such as dollars or GB.

### `calculated_spend` Block

* `actual_spend` - Amount of cost, usage, RI units, or Savings Plans units used. See [`actual_spend` Block](#actual_spend-block) for details.

### `actual_spend` Block

* `amount` - Cost or usage amount associated with the spend.
* `unit` - Unit of measurement used for the spend, such as USD or GBP.

### `cost_filter` Block

* `name` - Name of the cost filter.
* `values` - Values of the cost filter.

### `cost_types` Block

* `include_credit` - Whether to include credits in the cost budget.
* `include_discount` - Whether to include discounts in the cost budget.
* `include_other_subscription` - Whether to include other subscription costs in the cost budget.
* `include_recurring` - Whether to include recurring costs in the cost budget.
* `include_refund` - Whether to include refunds in the cost budget.
* `include_subscription` - Whether to include subscriptions in the cost budget.
* `include_support` - Whether to include support costs in the cost budget.
* `include_tax` - Whether to include tax in the cost budget.
* `include_upfront` - Whether to include upfront costs in the cost budget.
* `use_amortized` - Whether the budget uses the amortized rate.
* `use_blended` - Whether to use blended costs in the cost budget.

### `notification` Block

* `comparison_operator` - Comparison operator used to evaluate the condition. Valid values: `LESS_THAN`, `EQUAL_TO`, `GREATER_THAN`.
* `notification_type` - Type of budget value to notify on. Valid values: `ACTUAL`, `FORECASTED`.
* `subscriber_email_addresses` - Email addresses to notify.
* `subscriber_sns_topic_arns` - SNS topics to notify.
* `threshold` - Threshold at which the notification is sent.
* `threshold_type` - Type of threshold. Valid values: `PERCENTAGE`, `ABSOLUTE_VALUE`.

### `planned_limit` Block

* `amount` - Amount of cost or usage measured for the budget.
* `start_time` - Start time of the budget limit. Format: `2017-01-01_12:00`.
* `unit` - Unit of measurement used for the budget, such as dollars or GB.
