---
subcategory: "Billing"
layout: "aws"
page_title: "AWS: aws_billing_view"
description: |-
  Manages an AWS Billing View.
---

# Resource: aws_billing_view

Manages an AWS Billing View.

## Example Usage

### Basic Usage

```terraform
resource "aws_billing_view" "example" {
  name         = "example"
  description  = "example description"
  source_views = ["arn:aws:billing::123456789012:billingview/example"]
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the custom billing view to be created.
* `source_views` - (Required) List of ARNs of the source data views for the custom billing view.

The following arguments are optional:

* `description` - (Optional) Description of the custom billing view.
* `data_filter_expression` - (Optional) Filter Cost Explorer APIs using the expression. Refer to the [data-filter-expression block](#data-filter-expression) documentation for more details.
* `tags` - (Optional) List of key value map specifying tags associated to the billing view being created.

### data-filter-expression

A `data-filter-expression` block supports the following:

* `dimensions` - (Optional) Dimension to use for `expression`. Refer to [#dimensions](#dimensions) for more details.
* `tags` - (Optional) Tags to use for `expression`. Refer to [#tags](#tags) for more details.
* `time_range` - (Optional) Time range to use for `expression`. Refer to [#time-range](#time-range) for more details.

#### dimensions

A `dimensions` block supports the following:

* `key` - (Required) Key of the dimension. Possible values are `LINKED_ACCOUNT`.
* `values` - (Required) List of metadata values that you can use to filter and group your results.

#### tags

A `tags` block supports the following:

* `key` - (Required) Key of the tag.
* `values` - (Required) List of values for the tag.

#### time-range

A `time_range` block supports the following:

* `begin_date_inclusive` - (Required) Inclusive start date of the time range.
* `begin_date_inclusive` - (Required) Inclusive end date of the time range.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the View.
* `billing_view_type` - Type of billing group. Valid values are PRIMARY|BILLING_GROUP|CUSTOM.
* `created_at` - Timestamp when the billing view was created.
* `derived_view_count` - Number of billing views that use this billing view as a source.
* `owner_account_id` - Account owner of the billing view.
* `source_account_id` - AWS account ID that owns the source billing view, if this is a derived billing view.
* `source_view_count` - Number of source views associated with this billing view.
* `tags_all` - List of key value map specifying tags associated to the billing view.
* `updated_at` - Time when the billing view was last updated.
* `view_definition_last_updated_at` - Timestamp of when the billing view definition was last updated.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `5m`)
* `update` - (Default `5m`)
* `delete` - (Default `5m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Billing View using the `arn`. For example:

```terraform
import {
  to = aws_billing_view.example
  id = "arn:aws:billing::123456789012:billing-view/example"
}
```

Using `terraform import`, import Billing View using the `arn`. For example:

```console
% terraform import aws_billing_view.example arn:aws:billing::123456789012:billing-view/example
```
