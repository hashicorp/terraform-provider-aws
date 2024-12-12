---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_vpc_endpoint_subnet_association"
description: |-
  Provides a resource to create an association between a VPC endpoint and a subnet.
---

# Resource: aws_vpc_endpoint_subnet_association

Provides a resource to create an association between a VPC endpoint and a subnet.

~> **NOTE on VPC Endpoints and VPC Endpoint Subnet Associations:** Terraform provides
both a standalone VPC Endpoint Subnet Association (an association between a VPC endpoint
and a single `subnet_id`) and a [VPC Endpoint](vpc_endpoint.html) resource with a `subnet_ids`
attribute. Do not use the same subnet ID in both a VPC Endpoint resource and a VPC Endpoint Subnet
Association resource. Doing so will cause a conflict of associations and will overwrite the association.

## Example Usage

Basic usage:

```terraform
resource "aws_vpc_endpoint_subnet_association" "sn_ec2" {
  vpc_endpoint_id = aws_vpc_endpoint.ec2.id
  subnet_id       = aws_subnet.sn.id
}
```

## Argument Reference

This resource supports the following arguments:

* `vpc_endpoint_id` - (Required) The ID of the VPC endpoint with which the subnet will be associated.
* `subnet_id` - (Required) The ID of the subnet to be associated with the VPC endpoint.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The ID of the association.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `10m`)
- `delete` - (Default `10m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import VPC Endpoint Subnet Associations using `vpc_endpoint_id` together with `subnet_id`. For example:

```terraform
import {
  to = aws_vpc_endpoint_subnet_association.example
  id = "vpce-aaaaaaaa/subnet-bbbbbbbbbbbbbbbbb"
}
```

Using `terraform import`, import VPC Endpoint Subnet Associations using `vpc_endpoint_id` together with `subnet_id`. For example:

```console
% terraform import aws_vpc_endpoint_subnet_association.example vpce-aaaaaaaa/subnet-bbbbbbbbbbbbbbbbb
```
