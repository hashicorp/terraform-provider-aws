---
subcategory: "AppStream 2.0"
layout: "aws"
page_title: "AWS: aws_appstream_usage_report_subscription"
description: |-
  Manages an AppStream 2.0 Usage Report Subscription.
---

# Resource: aws_appstream_usage_report_subscription

Manages an AppStream 2.0 Usage Report Subscription. Enabling this generates daily usage reports and stores them in an Amazon S3 bucket in your account.

## Example Usage

```terraform
resource "aws_appstream_usage_report_subscription" "example" {}
```

## Argument Reference

This resource does not accept any arguments.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `s3_bucket_name` - The Amazon S3 bucket where generated usage reports are stored.
* `schedule` - The schedule for generating usage reports (`DAILY`).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import AppStream Usage Report Subscriptions using the account ID. For example:

```terraform
import {
  to = aws_appstream_usage_report_subscription.example
  id = "123456789012"
}
```

Using `terraform import`, import AppStream Usage Report Subscriptions using the account ID. For example:

```console
% terraform import aws_appstream_usage_report_subscription.example 123456789012
```
