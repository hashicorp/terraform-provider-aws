---
layout: "aws"
page_title: "AWS: aws_budget"
sidebar_current: "docs-aws-resource-budget"
description: |-
  Provides a budget resource.
---

# aws_budget

Provides a budget resource. Budgets use the cost visualisation provided by Cost Explorer to show you the status of your budgets, to provide forecasts of your estimated costs, and to track your AWS usage, including your free tier usage.

## Example Usage

```hcl
resource "aws_budget" "ec2" {
  budget_name           = "budget-ec2-monthly"
  budget_type           = "spend"
  limit_amount          = "1200"
  limit_unit            = "dollars"
  time_period_end       = "2087-06-15_00:00"
  time_period_start     = "2017-07-01"
  time_unit             = "MONTHLY"

  cost_filters {
    Service = "ec2"
  }
}
```

## Argument Reference

For more detailed documentation about each argument, refer to the [AWS official
documentation](http://docs.aws.amazon.com/awsaccountbilling/latest/aboutv2/data-type-budget.html).

The following arguments are supported:

* `budget_name` - The name of a budget. Unique within accounts.
* `budget_type` - Whether this budget tracks monetary cost or usage.
* `cost_filters` - Map of [CostFilters](#CostFilters) key/value pairs to apply to the budget.
* `include_blended` - Boolean value indicating whether the budget uses blended rate. Defaults to `false`.
* `include_subscriptions` - Boolean value indicating whether the budget includes subscriptions. Defaults to `false`.
* `include_tax` - Boolean value indicating whether to include tax in budget. Defaults to `true`.
* `limit_amount` - The amount of cost or usage being measured for a budget.
* `limit_unit` - The unit of measurement used for the budget forecast, actual spend, or budget threshold, such as dollars or GB. See [Spend ](http://docs.aws.amazon.com/awsaccountbilling/latest/aboutv2/data-type-spend.html) documentation.
* `time_period_end` - The end of the time period covered by the budget. There are no restrictions on the end date. Format: `2017-01-01_12:00`.
* `time_period_start` - The start of the time period covered by the budget. The start date must come before the end date. Format: `2017-01-01_12:00`.
* `time_unit` - The length of time until a budget resets the actual and forecasted spend. Valid values: `MONTHLY`, `QUARTERLY`, `ANNUALLY`.

## Attributes Reference

The following attributes are exported:

* `budget_name` - The name of the budget.
* `budget_type` - The type of the budget.
* `cost_filters` - Map of cost filters applied to the budget. See [CostFilters](#CostFilters) for possible options.
* `include_blended` - whether the budget uses blended rate
* `include_subscriptions` - whether the budget includes subscriptions
* `include_tax` - Whether tax is included in the budget.
* `limit_amount` - The amount of cost or usage being measured
* `limit_unit` - The unit of measurement used, such as dollars or GB.
* `time_period_end` - The end of the time period covered by the budget
* `time_period_start` - The start of the time period covered by the budget
* `time_unit` - The length of time until the budget resets the actual and forecasted spend.

### CostFilters

Valid keys for `cost_filters` parameter vary depending on the `budget_type` value.

* `cost`
  * `AZ`
  * `LinkedAccount`
  * `Operation`
  * `PurchaseType`
  * `Service`
  * `TagKeyValue`
* `usage`
  * `AZ`
  * `LinkedAccount`
  * `Operation`
  * `PurchaseType`
  * `UsageType:<service name>`
  * `TagKeyValue`

Refer to [AWS CostFilter documentation](http://docs.aws.amazon.com/awsaccountbilling/latest/aboutv2/data-type-filter.html) for further detail.

## Examples

Create a budget for *$100*.

```hcl
resource "aws_budget" "cost" {
  ...
  budget_type  = "spend"
  limit_amount = "100"
  limit_unit   = "dollars"
}
```

Create a budget for s3 with a limit of *3 GB* of storage.

```hcl
resource "aws_budget" "s3" {
  ...
  budget_type  = "usage"
  limit_amount = "3"
  limit_unit   = "GB"
}
```

## Import

Budgets can be imported using the budget `name`, e.g.

``` $ terraform import aws_budget.default my-budget ```
