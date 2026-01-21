---
subcategory: "Savings Plans"
layout: "aws"
page_title: "AWS: aws_savingsplans_savings_plan"
description: |-
  Get information on an AWS Savings Plan.
---

# Data Source: aws_savingsplans_savings_plan

Use this data source to get information on an existing AWS Savings Plan.

## Example Usage

```terraform
data "aws_savingsplans_savings_plan" "example" {
  savings_plan_id = "sp-12345678901234567"
}

output "savings_plan_state" {
  value = data.aws_savingsplans_savings_plan.example.state
}
```

## Argument Reference

The following arguments are required:

* `savings_plan_id` - (Required) The ID of the Savings Plan.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `savings_plan_arn` - The ARN of the Savings Plan.
* `state` - The current state of the Savings Plan.
* `start` - The start time of the Savings Plan.
* `end` - The end time of the Savings Plan.
* `savings_plan_type` - The type of Savings Plan.
* `payment_option` - The payment option for the Savings Plan.
* `currency` - The currency of the Savings Plan.
* `commitment` - The hourly commitment amount.
* `upfront_payment_amount` - The up-front payment amount.
* `recurring_payment_amount` - The recurring payment amount.
* `term_duration_in_seconds` - The duration of the term, in seconds.
* `ec2_instance_family` - The EC2 instance family for the Savings Plan.
* `region` - The AWS Region.
* `offering_id` - The ID of the offering.
* `tags` - A map of tags assigned to the resource.
