---
subcategory: "Savings Plans"
layout: "aws"
page_title: "AWS: aws_savingsplans_savings_plan"
description: |-
  Provides an AWS Savings Plan resource.
---

# Resource: aws_savingsplans_savings_plan

Provides an AWS Savings Plan resource.

~> **WARNING:** Savings Plans represent a financial commitment. Once a Savings Plan becomes active, it **cannot be cancelled or deleted**. Only Savings Plans in the `queued` state (scheduled for future purchase) can be deleted. Use this resource with caution.

~> **Note:** Importing an active Savings Plan will add it to your Terraform state, but destroying it will only remove it from state - the actual Savings Plan will continue until its term ends.

## Example Usage

### Basic Usage

```terraform
resource "aws_savingsplans_savings_plan" "example" {
  savings_plan_offering_id = "00000000-0000-0000-0000-000000000000"
  commitment               = "1.0"

  tags = {
    Environment = "production"
  }
}
```

### Scheduled Purchase

```terraform
resource "aws_savingsplans_savings_plan" "scheduled" {
  savings_plan_offering_id = "00000000-0000-0000-0000-000000000000"
  commitment               = "5.0"
  purchase_time            = "2026-12-01T00:00:00Z"

  tags = {
    Environment = "production"
  }
}
```

## Argument Reference

The following arguments are required:

* `savings_plan_offering_id` - (Required) The unique ID of a Savings Plan offering. You can find available offerings using the `aws savingsplans describe-savings-plans-offerings` CLI command.
* `commitment` - (Required) The hourly commitment, in USD. This is the amount you commit to pay per hour, regardless of actual usage.

The following arguments are optional:

* `purchase_time` - (Optional) The time at which to purchase the Savings Plan, in UTC format (YYYY-MM-DDTHH:MM:SSZ). If not specified, the plan is purchased immediately. Plans with a future purchase time are placed in `queued` state and can be deleted before they become active.
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `savings_plan_arn` - The ARN of the Savings Plan.
* `savings_plan_id` - The ID of the Savings Plan.
* `state` - The current state of the Savings Plan (e.g., `active`, `queued`, `retired`).
* `start` - The start time of the Savings Plan in RFC3339 format.
* `end` - The end time of the Savings Plan in RFC3339 format.
* `savings_plan_type` - The type of Savings Plan (e.g., `Compute`, `EC2Instance`).
* `payment_option` - The payment option for the Savings Plan (e.g., `All Upfront`, `Partial Upfront`, `No Upfront`).
* `currency` - The currency of the Savings Plan (e.g., `USD`).
* `upfront_payment_amount` - The up-front payment amount.
* `recurring_payment_amount` - The recurring payment amount.
* `term_duration_in_seconds` - The duration of the term, in seconds.
* `ec2_instance_family` - The EC2 instance family for the Savings Plan (only applicable to EC2 Instance Savings Plans).
* `region` - The AWS Region.
* `offering_id` - The ID of the offering.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

Using `terraform import`, import Savings Plans using the `id`. For example:

```terraform
import {
  to = aws_savingsplans_savings_plan.example
  id = "sp-12345678901234567"
}
```

Using `terraform state mv`, import Savings Plans using the `id`. For example:

```console
% terraform import aws_savingsplans_savings_plan.example sp-12345678901234567
```
