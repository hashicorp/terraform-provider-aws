---
subcategory: "Billing"
layout: "aws"
page_title: "AWS: aws_billing_views"
description: |-
  Retrieve a list of AWS Billing Views.
---

# Data Source: aws_billing_views

Provides details about an AWS Billing Views.

## Example Usage

### Basic Usage

```terraform
data "aws_billing_views" "example" {
  billing_view_types = ["PRIMARY"]
}

output "primary_view_arn_by_types" {
  value = data.aws_billing_views.example.billing_view[0].arn
}
```

```terraform
data "aws_billing_views" "example" {}

output "view_arns" {
  value = [for view in data.aws_billing_views.example.billing_view : view.arn]
}

output "primary_view_arn_by_name" {
  value = [for view in data.aws_billing_views.example.billing_view : view.arn if view.name == "Primary View"][0]
}
```

## Argument Reference

The following arguments are optional:

* `billing_view_types` - (Optional) List of billing view types to retrieve. Valid values are `PRIMARY`, `BILLING_GROUP`, `CUSTOM`.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `billing_view` - List of billing view objects with the following attributes:
    * `arn` - ARN of the billing view.
    * `description` - Description of the billing view.
    * `name` - Name of the billing view.
    * `owner_account_id` - Account ID of the billing view owner.
