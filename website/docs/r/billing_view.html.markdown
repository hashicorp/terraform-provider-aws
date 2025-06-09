---
subcategory: "Billing"
layout: "aws"
page_title: "AWS: aws_billing_view"
description: |-
  Terraform resource for managing an AWS Billing View.
---
# Resource: aws_billing_view

Terraform resource for managing an AWS Billing View.

## Example Usage

### Basic Usage

```terraform
resource "aws_billing_view" "example" {
  name         = "example"
  source_views = ["arn:aws:billing::123456789012:billing-view/primary"]
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the custom billing view to be created.
* `source_views` - (Required) List of ARNs of the source data views for the custom billing view.

The following arguments are optional:

* `client_token` - (Optional) A unique, case-sensitive identifier you specify to ensure idempotency of the request.
* `description` - (Optional) Description of the custom billing view.
* `resource_tags` - (Optional) A list of key value map specifying tags associated to the billing view being created. Refer to the [resource-tags block](#resource-tags) documentation for more details.
* `data_filter_experession` - (Optional) Filter Cost Explorer APIs using the expression. Refer to the [data-filter-expression block](#data-filter-expression) documentation for more details.

### resource-tags

A `resource-tags` block supports the following:

* `key` - (Required) The key of the tag.
* `value` - (Required) The value of the tag.

### data-filter-expression

A `data-filter-expression` block supports the following:

* `dimensions` - (Optional) The specific `dimension` to use for `expression`. Refer to [#dimensions](#dimensions) for more details.
* `tags` - (Optional) The specific `tag` to use for `expression`. Refer to [#tags](#tags) for more details.

#### dimensions

A `dimensions` block supports the following:

* `key` - (Required) The key of the dimension. Possible values are `LINKED_ACCOUNT`.
* `values` - (Required) A list of metadata values that you can use to filter and group your results.

#### tags

A `tags` block supports the following:

* `key` - (Required) The key of the tag.
* `values` - (Required) A list of values for the tag.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the View.
* `created_at` - The timestamp when the billing view was created.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

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
