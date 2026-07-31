---
subcategory: "Pricing Plan Manager"
layout: "aws"
page_title: "AWS: aws_pricingplanmanager_subscription"
description: |-
  Terraform resource for managing an AWS Pricing Plan Manager Subscription.
---

# Resource: aws_pricingplanmanager_subscription

Terraform resource for managing an AWS Pricing Plan Manager Subscription. A subscription applies a flat-rate pricing plan, such as an [Amazon CloudFront flat-rate plan](https://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/flat-rate-plans.html), to a set of AWS resources: you pay a fixed recurring fee for the covered resources instead of usage-based pricing.

~> Destroying this resource cancels the subscription. For active paid subscriptions, AWS schedules the cancellation to take effect at the end of the current billing period, so the subscription remains visible (and billed) in AWS until that date. Subscriptions in `PENDING_APPROVAL` status (and free-tier subscriptions) are deleted immediately. If a scheduled change (such as a downgrade) is pending, it is reverted automatically before the cancellation is requested.

~> While a distribution is covered by a subscription, CloudFront blocks removing or replacing its web ACL and blocks deleting the distribution — including after cancellation, until the cancellation takes effect at the end of the billing cycle.

~> Resources must meet the plan family's eligibility requirements or `CreateSubscription` fails with a `ValidationException`. For the `CloudFront` plan family, distributions using legacy cache settings (`forwarded_values`) are rejected as not eligible — use [cache policies](https://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/working-with-policies.html) instead.

## Example Usage

### Basic Usage

```terraform
resource "aws_pricingplanmanager_subscription" "example" {
  plan_family = "CloudFront"
  plan_tier   = "PRO"

  resource_arns = [
    aws_cloudfront_distribution.example.arn,
    aws_wafv2_web_acl.example.arn,
  ]
}
```

### Manual Approval

Paid-tier subscriptions created with `approval_mode = "MANUAL"` remain in `PENDING_APPROVAL` status, and billing does not start, until approved with a separate `ApprovePaidSubscription` API call.

```terraform
resource "aws_pricingplanmanager_subscription" "example" {
  approval_mode = "MANUAL"
  plan_family   = "CloudFront"
  plan_tier     = "BUSINESS"

  resource_arns = [
    aws_cloudfront_distribution.example.arn,
    aws_wafv2_web_acl.example.arn,
  ]
}
```

## Argument Reference

The following arguments are required:

* `plan_family` - (Required, Forces new resource) Pricing plan family to subscribe to, such as `CloudFront`.
* `plan_tier` - (Required) Tier level for the subscription, such as `FREE`, `PRO`, `BUSINESS`, or `PREMIUM`. Upgrades take effect immediately. Downgrades are scheduled by AWS to take effect at the end of the current billing period: until then AWS keeps billing the old tier and this argument reflects the desired (scheduled) tier, with the pending change exposed in `scheduled_change`. Raising the tier again before the downgrade takes effect reverts the pending change.
* `resource_arns` - (Required) Set of ARNs of the AWS resources to include in the subscription, between 1 and 10 entries. For subscriptions in the `CloudFront` plan family, the resources must include exactly one CloudFront distribution and exactly one WAF web ACL, and can also include other supported resources such as Route 53 hosted zones and CloudFront KeyValueStores.

The following arguments are optional:

* `approval_mode` - (Optional, Forces new resource) Whether the subscription requires explicit approval before billing starts. Valid values: `MANUAL`, `IMMEDIATE`. Defaults to `IMMEDIATE`. With `MANUAL`, paid-tier subscriptions remain in `PENDING_APPROVAL` (unbilled, and unmodifiable) until approved with a separate `ApprovePaidSubscription` API call. This value is used only at creation time and is not returned by the AWS API.
* `usage_level` - (Optional) Usage level within the plan tier. Specify `DEFAULT` for the base configuration, or a higher level if the plan tier supports it. If omitted on an update, the usage level is reset to the plan tier's default.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the subscription.
* `etag` - Entity tag of the subscription, used by AWS for optimistic concurrency control.
* `scheduled_change` - Pending change that takes effect at the end of the current billing period, if any. See [`scheduled_change`](#scheduled_change) below.
* `status` - Current status of the subscription. Valid values: `PENDING_APPROVAL`, `ACTIVE`, `SYNC_IN_PROGRESS`, `FAILED`.
* `status_reason` - Human-readable explanation of the current status, when available.

### scheduled_change

* `change_type` - Type of pending change. Valid values: `DOWNGRADE`, `CANCELLATION`.
* `effective_date` - Date and time when the change takes effect.
* `plan_tier` - For downgrades, the tier level that the subscription changes to.
* `usage_level` - For downgrades, the target usage level after the change takes effect.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `60m`) Paid-tier subscriptions created with `IMMEDIATE` approval mode pass through `PENDING_APPROVAL` while AWS auto-approves them, which can take from minutes to over half an hour.
* `delete` - (Default `30m`) Reverting a pending scheduled change before cancellation waits out the intermediate `SYNC_IN_PROGRESS` status.
* `update` - (Default `30m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_pricingplanmanager_subscription.example
  identity = {
    arn = "arn:aws:pricingplanmanager::123456789012:subscription:sub_1a2B3c4D5e6F7g8H9i0JkLmNoPq"
  }
}

resource "aws_pricingplanmanager_subscription" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `arn` (String) ARN of the Pricing Plan Manager Subscription.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Pricing Plan Manager Subscription using the `arn`. For example:

```terraform
import {
  to = aws_pricingplanmanager_subscription.example
  id = "arn:aws:pricingplanmanager::123456789012:subscription:sub_1a2B3c4D5e6F7g8H9i0JkLmNoPq"
}
```

Using `terraform import`, import Pricing Plan Manager Subscription using the `arn`. For example:

```console
% terraform import aws_pricingplanmanager_subscription.example arn:aws:pricingplanmanager::123456789012:subscription:sub_1a2B3c4D5e6F7g8H9i0JkLmNoPq
```
