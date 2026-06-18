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

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `internet_gateway_block_mode` - (Required) Block mode. Needs to be one of `block-bidirectional`, `block-ingress`, `off`. If this resource is deleted, then this value will be set to `off` in the AWS account and region.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `aws_account_id` - The AWS account id to which these options apply.
* `aws_region` - The AWS region to which these options apply.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import VPC Block Public Access Options using the `aws_region`. For example:

```terraform
import {
  to = aws_vpc_block_public_access_options.example
  id = "us-east-1"
}
```

Using `terraform import`, import VPC Block Public Access Options using the `aws_region`. For example:

```console
% terraform import aws_vpc_block_public_access_options.example us-east-1
```
