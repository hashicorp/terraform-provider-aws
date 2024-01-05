---
subcategory: "Web Services Budgets"
layout: "aws"
page_title: "AWS: aws_budgets_budget"
description: |-
  Provides a budgets budget resource.
---

# Resource: aws_budgets_budget

Provides a budgets budget resource. Budgets use the cost visualisation provided by Cost Explorer to show you the status of your budgets, to provide forecasts of your estimated costs, and to track your AWS usage, including your free tier usage.

## Example Usage

```terraform
resource "aws_budgets_budget" "ec2" {
  name              = "budget-ec2-monthly"
  budget_type       = "COST"
  limit_amount      = "1200"
  limit_unit        = "USD"
  time_period_end   = "2087-06-15_00:00"
  time_period_start = "2017-07-01_00:00"
  time_unit         = "MONTHLY"

  cost_filter {
    name = "Service"
    values = [
      "Amazon Elastic Compute Cloud - Compute",
    ]
  }

  notification {
    comparison_operator        = "GREATER_THAN"
    threshold                  = 100
    threshold_type             = "PERCENTAGE"
    notification_type          = "FORECASTED"
    subscriber_email_addresses = ["test@example.com"]
  }
}
```

Create a budget for *$100*.

```terraform
resource "aws_budgets_budget" "cost" {
  # ...
  budget_type  = "COST"
  limit_amount = "100"
  limit_unit   = "USD"
}
```

Create a budget with planned budget limits.

```terraform
resource "aws_budgets_budget" "cost" {
  # ...

  planned_limit {
    start_time = "2017-07-01_00:00"
    amount     = "100"
    unit       = "USD"
  }

  planned_limit {
    start_time = "2017-08-01_00:00"
    amount     = "200"
    unit       = "USD"
  }
}
```

Create a budget for s3 with a limit of *3 GB* of storage.

```terraform
resource "aws_budgets_budget" "s3" {
  # ...
  budget_type  = "USAGE"
  limit_amount = "3"
  limit_unit   = "GB"
}
```

Create a Savings Plan Utilization Budget

```terraform
resource "aws_budgets_budget" "savings_plan_utilization" {
  # ...
  budget_type  = "SAVINGS_PLANS_UTILIZATION"
  limit_amount = "100.0"
  limit_unit   = "PERCENTAGE"

  cost_types {
    include_credit             = false
    include_discount           = false
    include_other_subscription = false
    include_recurring          = false
    include_refund             = false
    include_subscription       = true
    include_support            = false
    include_tax                = false
    include_upfront            = false
    use_blended                = false
  }
}
```

Create a RI Utilization Budget

```terraform
resource "aws_budgets_budget" "ri_utilization" {
  # ...
  budget_type  = "RI_UTILIZATION"
  limit_amount = "100.0" # RI utilization must be 100
  limit_unit   = "PERCENTAGE"

  #Cost types must be defined for RI budgets because the settings conflict with the defaults
  cost_types {
    include_credit             = false
    include_discount           = false
    include_other_subscription = false
    include_recurring          = false
    include_refund             = false
    include_subscription       = true
    include_support            = false
    include_tax                = false
    include_upfront            = false
    use_blended                = false
  }

  # RI Utilization plans require a service cost filter to be set
  cost_filter {
    name = "Service"
    values = [
      "Amazon Relational Database Service",
    ]
  }
}
```

Create a Cost Filter using Resource Tags

```terraform
resource "aws_budgets_budget" "cost" {
  # ...
  cost_filter {
    name = "TagKeyValue"
    values = [
      "TagKey$TagValue",
    ]
  }
}
```

Create a cost_filter using resource tags, obtaining the tag value from a terraform variable

```terraform
resource "aws_budgets_budget" "cost" {
  # ...
  cost_filter {
    name = "TagKeyValue"
    values = [
      "TagKey${"$"}${var.TagValue}"
    ]
  }
}
```

## Argument Reference

For more detailed documentation about each argument, refer to the [AWS official
documentation](http://docs.aws.amazon.com/awsaccountbilling/latest/aboutv2/data-type-budget.html).

This argument supports the following arguments:

* `account_id` - (Optional) The ID of the target account for budget. Will use current user's account_id by default if omitted.
* `auto_adjust_data` - (Optional) Object containing [AutoAdjustData] which determines the budget amount for an auto-adjusting budget.
* `name` - (Optional) The name of a budget. Unique within accounts.
* `name_prefix` - (Optional) The prefix of the name of a budget. Unique within accounts.
* `budget_type` - (Required) Whether this budget tracks monetary cost or usage.
* `cost_filter` - (Optional) A list of [CostFilter](#cost-filter) name/values pair to apply to budget.
* `cost_types` - (Optional) Object containing [CostTypes](#cost-types) The types of cost included in a budget, such as tax and subscriptions.
* `limit_amount` - (Required) The amount of cost or usage being measured for a budget.
* `limit_unit` - (Required) The unit of measurement used for the budget forecast, actual spend, or budget threshold, such as dollars or GB. See [Spend](http://docs.aws.amazon.com/awsaccountbilling/latest/aboutv2/data-type-spend.html) documentation.
* `time_period_end` - (Optional) The end of the time period covered by the budget. There are no restrictions on the end date. Format: `2017-01-01_12:00`.
* `time_period_start` - (Optional) The start of the time period covered by the budget. If you don't specify a start date, AWS defaults to the start of your chosen time period. The start date must come before the end date. Format: `2017-01-01_12:00`.
* `time_unit` - (Required) The length of time until a budget resets the actual and forecasted spend. Valid values: `MONTHLY`, `QUARTERLY`, `ANNUALLY`, and `DAILY`.
* `notification` - (Optional) Object containing [Budget Notifications](#budget-notification). Can be used multiple times to define more than one budget notification.
* `planned_limit` - (Optional) Object containing [Planned Budget Limits](#planned-budget-limits). Can be used multiple times to plan more than one budget limit. See [PlannedBudgetLimits](https://docs.aws.amazon.com/aws-cost-management/latest/APIReference/API_budgets_Budget.html#awscostmanagement-Type-budgets_Budget-PlannedBudgetLimits) documentation.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - id of resource.
* `arn` - The ARN of the budget.

### Auto Adjust Data

The parameters that determine the budget amount for an auto-adjusting budget.

`auto_adjust_type` (Required) - The string that defines whether your budget auto-adjusts based on historical or forecasted data. Valid values: `FORECAST`,`HISTORICAL`
`historical_options` (Optional) - Configuration block of [Historical Options](#historical-options). Required for `auto_adjust_type` of `HISTORICAL` Configuration block that defines the historical data that your auto-adjusting budget is based on.
`last_auto_adjust_time` (Optional) - The last time that your budget was auto-adjusted.

### Historical Options

`budget_adjustment_period` (Required) - The number of budget periods included in the moving-average calculation that determines your auto-adjusted budget amount.
`lookback_available_periods` (Optional) - The integer that describes how many budget periods in your BudgetAdjustmentPeriod are included in the calculation of your current budget limit. If the first budget period in your BudgetAdjustmentPeriod has no cost data, then that budget period isn’t included in the average that determines your budget limit. You can’t set your own LookBackAvailablePeriods. The value is automatically calculated from the `budget_adjustment_period` and your historical cost data.

### Cost Types

Valid keys for `cost_types` parameter.

* `include_credit` - A boolean value whether to include credits in the cost budget. Defaults to `true`
* `include_discount` - Whether a budget includes discounts. Defaults to `true`
* `include_other_subscription` - A boolean value whether to include other subscription costs in the cost budget. Defaults to `true`
* `include_recurring` - A boolean value whether to include recurring costs in the cost budget. Defaults to `true`
* `include_refund` - A boolean value whether to include refunds in the cost budget. Defaults to `true`
* `include_subscription` - A boolean value whether to include subscriptions in the cost budget. Defaults to `true`
* `include_support` - A boolean value whether to include support costs in the cost budget. Defaults to `true`
* `include_tax` - A boolean value whether to include tax in the cost budget. Defaults to `true`
* `include_upfront` - A boolean value whether to include upfront costs in the cost budget. Defaults to `true`
* `use_amortized` - Whether a budget uses the amortized rate. Defaults to `false`
* `use_blended` - A boolean value whether to use blended costs in the cost budget. Defaults to `false`

Refer to [AWS CostTypes documentation](https://docs.aws.amazon.com/aws-cost-management/latest/APIReference/API_budgets_CostTypes.html) for further detail.

### Cost Filter

Based on your choice of budget type, you can choose one or more of the available budget filters.

* `PurchaseType`
* `UsageTypeGroup`
* `Service`
* `Operation`
* `UsageType`
* `BillingEntity`
* `CostCategory`
* `LinkedAccount`
* `TagKeyValue`
* `LegalEntityName`
* `InvoicingEntity`
* `AZ`
* `Region`
* `InstanceType`

Refer to [AWS CostFilter documentation](https://docs.aws.amazon.com/cost-management/latest/userguide/budgets-create-filters.html) for further detail.

### Budget Notification

Valid keys for `notification` parameter.

* `comparison_operator` - (Required) Comparison operator to use to evaluate the condition. Can be `LESS_THAN`, `EQUAL_TO` or `GREATER_THAN`.
* `threshold` - (Required) Threshold when the notification should be sent.
* `threshold_type` - (Required) What kind of threshold is defined. Can be `PERCENTAGE` OR `ABSOLUTE_VALUE`.
* `notification_type` - (Required) What kind of budget value to notify on. Can be `ACTUAL` or `FORECASTED`
* `subscriber_email_addresses` - (Optional) E-Mail addresses to notify. Either this or `subscriber_sns_topic_arns` is required.
* `subscriber_sns_topic_arns` - (Optional) SNS topics to notify. Either this or `subscriber_email_addresses` is required.

### Planned Budget Limits

Valid keys for `planned_limit` parameter.

* `start_time` - (Required) The start time of the budget limit. Format: `2017-01-01_12:00`. See [PlannedBudgetLimits](https://docs.aws.amazon.com/aws-cost-management/latest/APIReference/API_budgets_Budget.html#awscostmanagement-Type-budgets_Budget-PlannedBudgetLimits) documentation.
* `amount` - (Required) The amount of cost or usage being measured for a budget.
* `unit` - (Required) The unit of measurement used for the budget forecast, actual spend, or budget threshold, such as dollars or GB. See [Spend](http://docs.aws.amazon.com/awsaccountbilling/latest/aboutv2/data-type-spend.html) documentation.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import budgets using `AccountID:BudgetName`. For example:

```terraform
import {
  to = aws_budgets_budget.myBudget
  id = "123456789012:myBudget"
}
```

Using `terraform import`, import budgets using `AccountID:BudgetName`. For example:

```console
% terraform import aws_budgets_budget.myBudget 123456789012:myBudget
```
