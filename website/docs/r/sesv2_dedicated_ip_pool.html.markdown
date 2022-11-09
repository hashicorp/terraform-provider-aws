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

* `scaling_mode` - (Optional) IP pool scaling mode. Valid values: `STANDARD`, `MANAGED`. If omitted, the AWS API will default to a standard pool.
* `tags` - (Optional) A map of tags to assign to the pool. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - ARN of the Dedicated IP Pool.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

SESv2 (Simple Email V2) Dedicated IP Pool can be imported using the `pool_name`, e.g.,

```
$ terraform import aws_sesv2_dedicated_ip_pool.example my-pool
```
