---
subcategory: "Config"
layout: "aws"
page_title: "AWS: aws_config_retention_configuration"
description: |-
  Provides a resource to manage the AWS Config retention configuration.
---

# Resource: aws_config_retention_configuration

Provides a resource to manage the AWS Config retention configuration.
The retention configuration defines the number of days that AWS Config stores historical information.

## Example Usage

```terraform
resource "aws_config_retention_configuration" "example" {
  retention_period_in_days = 90
}
```

## Argument Reference

This resource supports the following arguments:

* `retention_period_in_days` - (Required) The number of days AWS Config stores historical information.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `name` - The name of the retention configuration object. The object is always named **default**.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import the AWS Config retention configuration using the `name`. For example:

```terraform
import {
  to = aws_config_retention_configuration.example
  id = "default"
}
```

Using `terraform import`, import the AWS Config retention configuration using the `name`. For example:

```console
% terraform import aws_config_retention_configuration.example default
```
