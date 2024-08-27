---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_internet_gateway_attachment"
description: |-
  Provides a resource to create a VPC Internet Gateway Attachment.
---

# Resource: aws_internet_gateway_attachment

Provides a resource to create a VPC Internet Gateway Attachment.

## Example Usage

```terraform
resource "aws_internet_gateway_attachment" "example" {
  internet_gateway_id = aws_internet_gateway.example.id
  vpc_id              = aws_vpc.example.id
}

resource "aws_vpc" "example" {
  cidr_block = "10.1.0.0/16"
}

resource "aws_internet_gateway" "example" {}
```

## Argument Reference

This resource supports the following arguments:

* `internet_gateway_id` - (Required) The ID of the internet gateway.
* `vpc_id` - (Required) The ID of the VPC.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The ID of the VPC and Internet Gateway separated by a colon.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `20m`)
- `delete` - (Default `20m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Internet Gateway Attachments using the `id`. For example:

```terraform
import {
  to = aws_internet_gateway_attachment.example
  id = "igw-c0a643a9:vpc-123456"
}
```

Using `terraform import`, import Internet Gateway Attachments using the `id`. For example:

```console
% terraform import aws_internet_gateway_attachment.example igw-c0a643a9:vpc-123456
```
