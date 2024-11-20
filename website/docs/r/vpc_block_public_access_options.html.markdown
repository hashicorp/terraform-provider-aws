---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_vpc_block_public_access_options"
description: |-
  Terraform resource for managing AWS VPC Block Public Access Options in a region.
---

# Resource: aws_vpc_block_public_access_options

Terraform resource for managing an AWS VPC Block Public Access Options.

## Example Usage

### Basic Usage

```terraform
resource "aws_vpc_block_public_access_options" "example" {
  internet_gateway_block_mode = "block-bidirectional"
}
```

## Argument Reference

The following arguments are required:

* `internet_gateway_block_mode` - (Required) Block mode. Needs to be one of `block-bidirectional`, `block-ingress`, `off`. If this resource is deleted, then this value will be set to `off` in the AWS account and region.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `aws_account_id` - The AWS account id to which these options apply.
* `aws_region_id` - The AWS region to which these options apply.
* `last_update_timestamp` - Last Update Timestamp
* `reason` - Reason for update.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `60m`)
* `update` - (Default `180m`)
* `delete` - (Default `90m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import VPC Block Public Access Options using the `example_id_arg`. For example:

```terraform
import {
  to = aws_vpc_block_public_access_options.example
  id = "111222333444:us-east-1"
}
```

Using `terraform import`, import VPC Block Public Access Options using the `example_id_arg`. For example:

```console
% terraform import aws_vpc_block_public_access_options.example 111222333444:us-east-1
```
