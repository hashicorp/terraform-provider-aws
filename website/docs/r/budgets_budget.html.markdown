---
subcategory: "Web Services Budgets"
layout: "aws"
page_title: "AWS: aws_budgets_budget"
description: |-
  Provides a budgets budget resource.
---

# Resource: aws_budgets_budget

Provides a budgets budget resource. Budgets use the cost visualization provided by Cost Explorer to show you the status of your budgets, to provide forecasts of your estimated costs, and to track your AWS usage, including your free tier usage. For more detailed documentation about each argument, refer to the [AWS official documentation](http://docs.aws.amazon.com/awsaccountbilling/latest/aboutv2/data-type-budget.html).

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

  tags = {
    Tag1 = "Value1"
    Tag2 = "Value2"
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

Create a cost filter using resource tags

```terraform
resource "aws_budgets_budget" "cost" {
  # ...
  cost_filter {
    name = "TagKeyValue"
    values = [
      # Format "TagKey$TagValue",
      "aws:createdBy$Terraform",
      "user:business-unit$human_resources",
    ]
  }
}
```

Create a cost filter using resource tags, obtaining the tag value from a Terraform variable

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

Create a budget with a simple dimension filter for unblended costs

```terraform
resource "aws_budgets_budget" "simple" {
  name         = "budget-ec2-filter"
  budget_type  = "COST"
  limit_amount = "500"
  limit_unit   = "USD"
  time_unit    = "MONTHLY"

  metrics = ["UnblendedCost"]
  filter_expression {
    dimensions {
      key    = "SERVICE"
      values = ["Amazon Elastic Compute Cloud - Compute"]
    }
  }
}
```

Create a budget with AND filter for blended costs

```terraform
resource "aws_budgets_budget" "and_example" {
  name         = "budget-and-filter"
  budget_type  = "COST"
  limit_amount = "1200"
  limit_unit   = "USD"
  time_unit    = "MONTHLY"

  metrics = ["BlendedCost"]
  # Each `and` block is one operand. AND requires at least 2 operands.
  filter_expression {
    and {
      dimensions {
        key    = "SERVICE"
        values = ["Amazon Elastic Compute Cloud - Compute"]
      }
    }
    and {
      tags {
        key    = "Environment"
        values = ["Production"]
      }
    }
  }
}
```

Create a budget with OR filter for amortized costs

```terraform
resource "aws_budgets_budget" "or_example" {
  name         = "budget-or-filter"
  budget_type  = "COST"
  limit_amount = "2000"
  limit_unit   = "USD"
  time_unit    = "MONTHLY"

  metrics = ["AmortizedCost"]
  # Each `or` block is one operand. OR requires at least 2 operands.
  filter_expression {
    or {
      dimensions {
        key    = "SERVICE"
        values = ["Amazon Elastic Compute Cloud - Compute"]
      }
    }
    or {
      dimensions {
        key    = "SERVICE"
        values = ["Amazon Relational Database Service"]
      }
    }
  }
}
```

Create a budget with NOT filter for net unblended costs

```terraform
resource "aws_budgets_budget" "not_example" {
  name         = "budget-not-filter"
  budget_type  = "COST"
  limit_amount = "1000"
  limit_unit   = "USD"
  time_unit    = "MONTHLY"

  metrics = ["NetUnblendedCost"]
  filter_expression {
    not {
      dimensions {
        key    = "REGION"
        values = ["us-west-2"]
      }
    }
  }
}
```

Create a budget with a compound filter for net amortized costs

```terraform
resource "aws_budgets_budget" "compound_example" {
  name         = "budget-compound-filter"
  budget_type  = "COST"
  limit_amount = "1500"
  limit_unit   = "USD"
  time_unit    = "MONTHLY"

  metrics = ["NetAmortizedCost"]
  filter_expression {
    # First OR operand: an AND expression
    or {
      and {
        dimensions {
          key    = "SERVICE"
          values = ["Amazon Elastic Compute Cloud - Compute"]
        }
      }
      and {
        tags {
          key    = "Environment"
          values = ["production"]
        }
      }
      and {
        cost_categories {
          key    = "Environment"
          values = ["production"]
        }
      }
    }
    # Second OR operand: a NOT expression
    or {
      not {
        dimensions {
          key    = "REGION"
          values = ["us-west-2"]
        }
      }
    }
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

## Argument Reference

The following arguments are required:

* `budget_type` - (Required) Whether this budget tracks monetary cost or usage.
* `time_unit` - (Required) Length of time until a budget resets the actual and forecasted spend. Valid values: `MONTHLY`, `QUARTERLY`, `ANNUALLY`, and `DAILY`.

The following arguments are optional:

* `account_id` - (Optional) ID of the target account for budget. Uses the current user's account ID by default if omitted.
* `auto_adjust_data` - (Optional) Object containing [AutoAdjustData](#auto_adjust_data-block) which determines the budget amount for an auto-adjusting budget.
* `billing_view_arn` - (Optional) ARN of the billing view.
* `cost_filter` - (Optional) List of [CostFilter](#cost_filter-block) name/values pair to apply to budget. Conflicts with `filter_expression`.
* `cost_types` - (Optional) Object containing [CostTypes](#cost_types-block) that defines the types of cost included in a budget, such as tax and subscriptions.
* `filter_expression` - (Optional) Object containing [Filter Expression](#filter_expression-block) to apply to budget. Conflicts with `cost_filter` and requires `metrics`.
* `limit_amount` - (Optional) Amount of cost or usage being measured for a budget.
* `limit_unit` - (Optional) Unit of measurement used for the budget forecast, actual spend, or budget threshold, such as dollars or GB. See [Spend](http://docs.aws.amazon.com/awsaccountbilling/latest/aboutv2/data-type-spend.html) documentation.
* `metrics` - (Optional) List containing definition for how the budget data is aggregated. Conflicts with `cost_types` and requires `filter_expression`.
* `name` - (Optional) Name of a budget. Unique within accounts.
* `name_prefix` - (Optional) Prefix of the name of a budget. Unique within accounts.
* `notification` - (Optional) Object containing [Budget Notifications](#notification-block). Can be used multiple times to define more than one budget notification.
* `planned_limit` - (Optional) Object containing [Planned Budget Limits](#planned_limit-block). Can be used multiple times to plan more than one budget limit. See [PlannedBudgetLimits](https://docs.aws.amazon.com/aws-cost-management/latest/APIReference/API_budgets_Budget.html#awscostmanagement-Type-budgets_Budget-PlannedBudgetLimits) documentation.
* `tags` - (Optional) Map of tags assigned to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `time_period_end` - (Optional) End of the time period covered by the budget. There are no restrictions on the end date. Format: `2017-01-01_12:00`.
* `time_period_start` - (Optional) Start of the time period covered by the budget. If you don't specify a start date, AWS defaults to the start of your chosen time period. The start date must come before the end date. Format: `2017-01-01_12:00`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the budget.
* `id` - id of resource.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

### `auto_adjust_data` Block

The parameters that determine the budget amount for an auto-adjusting budget.

* `auto_adjust_type` - Whether your budget auto-adjusts based on historical or forecasted data. Valid values: `FORECAST`, `HISTORICAL`.
* `historical_options` - Configuration block of [Historical Options](#historical_options-block). Required for `auto_adjust_type` of `HISTORICAL`. Defines the historical data that your auto-adjusting budget is based on.
* `last_auto_adjust_time` - Last time that your budget was auto-adjusted.

### `historical_options` Block

* `budget_adjustment_period` - Number of budget periods included in the moving-average calculation that determines your auto-adjusted budget amount.
* `lookback_available_periods` - Integer that describes how many budget periods in your BudgetAdjustmentPeriod are included in the calculation of your current budget limit. If the first budget period in your BudgetAdjustmentPeriod has no cost data, then that budget period isn’t included in the average that determines your budget limit. You can’t set your own LookBackAvailablePeriods. The value is automatically calculated from the `budget_adjustment_period` and your historical cost data.

### `cost_types` Block

Valid keys for `cost_types` parameter. Refer to [AWS CostTypes documentation](https://docs.aws.amazon.com/aws-cost-management/latest/APIReference/API_budgets_CostTypes.html) for further detail.

* `include_credit` - Whether to include credits in the cost budget. Defaults to `true`.
* `include_discount` - Whether a budget includes discounts. Defaults to `true`.
* `include_other_subscription` - Whether to include other subscription costs in the cost budget. Defaults to `true`.
* `include_recurring` - Whether to include recurring costs in the cost budget. Defaults to `true`.
* `include_refund` - Whether to include refunds in the cost budget. Defaults to `true`.
* `include_subscription` - Whether to include subscriptions in the cost budget. Defaults to `true`.
* `include_support` - Whether to include support costs in the cost budget. Defaults to `true`.
* `include_tax` - Whether to include tax in the cost budget. Defaults to `true`.
* `include_upfront` - Whether to include upfront costs in the cost budget. Defaults to `true`.
* `use_amortized` - Whether a budget uses the amortized rate. Defaults to `false`.
* `use_blended` - Whether to use blended costs in the cost budget. Defaults to `false`.

### `cost_filter` Block

Valid keys for `cost_filter` parameter. Refer to [AWS CostFilter documentation](https://docs.aws.amazon.com/cost-management/latest/userguide/budgets-create-filters.html) for further detail.

* `name` - Name of the cost filter. Valid values are `AZ`, `BillingEntity`, `CostCategory`, `InstanceType`, `InvoicingEntity`, `LegalEntityName`, `LinkedAccount`, `Operation`, `PurchaseType`, `Region`, `Service`, `TagKeyValue`, `UsageType`, and `UsageTypeGroup`.
* `values` - List of values used for filtering.

### `notification` Block

Valid keys for `notification` parameter.

* `comparison_operator` - Comparison operator to use to evaluate the condition. Can be `LESS_THAN`, `EQUAL_TO` or `GREATER_THAN`.
* `notification_type` - What kind of budget value to notify on. Can be `ACTUAL` or `FORECASTED`.
* `subscriber_email_addresses` - E-Mail addresses to notify. Either this or `subscriber_sns_topic_arns` is required.
* `subscriber_sns_topic_arns` - SNS topics to notify. Either this or `subscriber_email_addresses` is required.
* `threshold` - Threshold when the notification should be sent.
* `threshold_type` - What kind of threshold is defined. Can be `PERCENTAGE` OR `ABSOLUTE_VALUE`.

### `planned_limit` Block

Valid keys for `planned_limit` parameter.

* `amount` - Amount of cost or usage being measured for a budget.
* `start_time` - Start time of the budget limit. Format: `2017-01-01_12:00`. See [PlannedBudgetLimits](https://docs.aws.amazon.com/aws-cost-management/latest/APIReference/API_budgets_Budget.html#awscostmanagement-Type-budgets_Budget-PlannedBudgetLimits) documentation.
* `unit` - Unit of measurement used for the budget forecast, actual spend, or budget threshold, such as dollars or GB. See [Spend](http://docs.aws.amazon.com/awsaccountbilling/latest/aboutv2/data-type-spend.html) documentation.

### `filter_expression` Block

The `filter_expression` block maps directly to the [AWS Expression](https://docs.aws.amazon.com/aws-cost-management/latest/APIReference/API_budgets_Expression.html) object. Each expression block must have **exactly one root** — you can set one of `and`, `or`, `not`, `dimensions`, `tags`, or `cost_categories`, but not multiple at the same level. Refer to [AWS Expression documentation](https://docs.aws.amazon.com/aws-cost-management/latest/APIReference/API_budgets_Expression.html) for further detail.

~> **Important:** `and` and `or` require **at least 2 operands**. Each operand is a separate `and` or `or` block within the parent. Do not place multiple leaf filters (e.g., `dimensions` and `tags`) inside a single `and`/`or` block — instead, use one block per leaf. The maximum expression nesting depth allowed by the AWS API is 2.

* `and` - List of filter expressions to combine with AND logic. Each `and` block is one operand and must itself contain exactly one root.
* `cost_categories` - [Cost Categories](#cost_categories-block) block.
* `dimensions` - [Dimensions](#dimensions-block) block.
* `not` - Single filter expression to negate. Must contain exactly one root.
* `or` - List of filter expressions to combine with OR logic. Each `or` block is one operand and must itself contain exactly one root.
* `tags` - [Tags](#tags-block) block.

### `dimensions` Block

* `key` - Dimension to filter on. Valid values include `AZ`, `INSTANCE_TYPE`, `LINKED_ACCOUNT`, `OPERATION`, `PURCHASE_TYPE`, `REGION`, `SERVICE`, `USAGE_TYPE`, `USAGE_TYPE_GROUP`, `RECORD_TYPE`, `OPERATING_SYSTEM`, `TENANCY`, `SCOPE`, `PLATFORM`, `SUBSCRIPTION_ID`, `LEGAL_ENTITY_NAME`, `DEPLOYMENT_OPTION`, `DATABASE_ENGINE`, `CACHE_ENGINE`, `INSTANCE_TYPE_FAMILY`, `BILLING_ENTITY`, `RESERVATION_ID`, `RESOURCE_ID`, `RIGHTSIZING_TYPE`, `SAVINGS_PLANS_TYPE`, `SAVINGS_PLAN_ARN`, `PAYMENT_OPTION`, and `AGREEMENT_END_DATE_TIME_AFTER`, `AGREEMENT_END_DATE_TIME_BEFORE`.
* `match_options` - Match options for the dimension filter. Valid values are `EQUALS`, `STARTS_WITH`, `ENDS_WITH`, `CONTAINS`, `GREATER_THAN_OR_EQUAL`, `CASE_SENSITIVE`, `CASE_INSENSITIVE`. Note: `ABSENT` is not supported due to AWS API contradictions (it requires values to be absent but also cannot have values set).
* `values` - List of values to match against the dimension. At least one value is required.

### `tags` Block

* `key` - Tag key to filter on.
* `match_options` - Match options for the tag filter. Valid values are `EQUALS`, `STARTS_WITH`, `ENDS_WITH`, `CONTAINS`, `GREATER_THAN_OR_EQUAL`, `CASE_SENSITIVE`, `CASE_INSENSITIVE`. Note: `ABSENT` is not supported due to AWS API contradictions (it requires values to be absent but also cannot have values set).
* `values` - List of tag values to match. At least one value is required.

### `cost_categories` Block

* `key` - Cost category key to filter on.
* `match_options` - Match options for the cost category filter. Valid values are `EQUALS`, `STARTS_WITH`, `ENDS_WITH`, `CONTAINS`, `GREATER_THAN_OR_EQUAL`, `CASE_SENSITIVE`, `CASE_INSENSITIVE`. Note: `ABSENT` is not supported due to AWS API contradictions (it requires values to be absent but also cannot have values set).
* `values` - List of cost category values to match. At least one value is required.

### Metrics

`metrics` defines how AWS calculates the budget amount. Set `metrics` when using `filter_expression`. This argument is a list with a single value. Valid values are `UnblendedCost`, `BlendedCost`, `AmortizedCost`, `NetUnblendedCost`, `NetAmortizedCost`, `UsageQuantity`, `NormalizedUsageAmount`, and `Hours`.

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
