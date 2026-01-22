---
subcategory: "SESv2 (Simple Email V2)"
layout: "aws"
page_title: "AWS: aws_sesv2_dedicated_ip_pool"
description: |-
  Terraform resource for managing an AWS SESv2 (Simple Email V2) Dedicated IP Pool.
---

# Resource: aws_sesv2_dedicated_ip_pool

Terraform resource for managing an AWS SESv2 (Simple Email V2) Dedicated IP Pool.

## Example Usage

### Basic Usage

```terraform
resource "aws_sesv2_dedicated_ip_pool" "example" {
  pool_name = "my-pool"
}
```

### Managed Pool

```terraform
resource "aws_sesv2_dedicated_ip_pool" "example" {
  pool_name    = "my-managed-pool"
  scaling_mode = "MANAGED"
}
```

## Argument Reference

The following arguments are required:

* `pool_name` - (Required) Name of the dedicated IP pool.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `scaling_mode` - (Optional) IP pool scaling mode. Valid values: `STANDARD`, `MANAGED`. If omitted, the AWS API will default to a standard pool.
* `tags` - (Optional) A map of tags to assign to the pool. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Dedicated IP Pool.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import SESv2 (Simple Email V2) Dedicated IP Pool using the `pool_name`. For example:

```terraform
import {
  to = aws_sesv2_dedicated_ip_pool.example
  id = "my-pool"
}
```

Using `terraform import`, import SESv2 (Simple Email V2) Dedicated IP Pool using the `pool_name`. For example:

```console
% terraform import aws_sesv2_dedicated_ip_pool.example my-pool
```
