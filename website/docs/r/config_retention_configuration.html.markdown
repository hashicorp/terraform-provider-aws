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

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `retention_period_in_days` - (Required) The number of days AWS Config stores historical information.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `name` - The name of the retention configuration object. The object is always named **default**.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_config_retention_configuration.example
  identity = {
    name = "default"
  }
}

resource "aws_config_retention_configuration" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `name` (String) Name of the rule.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Config Retention Configurations using the `name`. For example:

```terraform
import {
  to = aws_config_retention_configuration.example
  id = "default"
}
```

Using `terraform import`, import Config Retention Configurations using the `name`. For example:

```console
% terraform import aws_config_retention_configuration.example default
```
