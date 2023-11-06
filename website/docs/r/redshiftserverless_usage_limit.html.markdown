---
subcategory: "Redshift Serverless"
layout: "aws"
page_title: "AWS: aws_redshiftserverless_usage_limit"
description: |-
  Provides a Redshift Serverless Usage Limit resource.
---

# Resource: aws_redshiftserverless_usage_limit

Creates a new Amazon Redshift Serverless Usage Limit.

## Example Usage

```terraform
resource "aws_redshiftserverless_workgroup" "example" {
  namespace_name = aws_redshiftserverless_namespace.example.namespace_name
  workgroup_name = "example"
}

resource "aws_redshiftserverless_usage_limit" "example" {
  resource_arn = aws_redshiftserverless_workgroup.example.arn
  usage_type   = "serverless-compute"
  amount       = 60
}
```

## Argument Reference

This resource supports the following arguments:

* `amount` - (Required) The limit amount. If time-based, this amount is in Redshift Processing Units (RPU) consumed per hour. If data-based, this amount is in terabytes (TB) of data transferred between Regions in cross-account sharing. The value must be a positive number.
* `breach_action` - (Optional) The action that Amazon Redshift Serverless takes when the limit is reached. Valid values are `log`, `emit-metric`, and `deactivate`. The default is `log`.
* `period` - (Optional) The time period that the amount applies to. A weekly period begins on Sunday. Valid values are `daily`, `weekly`, and `monthly`. The default is `monthly`.
* `resource_arn` - (Required) The Amazon Resource Name (ARN) of the Amazon Redshift Serverless resource to create the usage limit for.
* `usage_type` - (Required) The type of Amazon Redshift Serverless usage to create a usage limit for. Valid values are `serverless-compute` or `cross-region-datasharing`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - Amazon Resource Name (ARN) of the Redshift Serverless Usage Limit.
* `id` - The Redshift Usage Limit id.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Redshift Serverless Usage Limits using the `id`. For example:

```terraform
import {
  to = aws_redshiftserverless_usage_limit.example
  id = "example-id"
}
```

Using `terraform import`, import Redshift Serverless Usage Limits using the `id`. For example:

```console
% terraform import aws_redshiftserverless_usage_limit.example example-id
```
