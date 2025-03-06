---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_vpc_endpoint_security_group_association"
description: |-
  Provides a resource to create an association between a VPC endpoint and a security group.
---

# Resource: aws_vpc_endpoint_security_group_association

Provides a resource to create an association between a VPC endpoint and a security group.

~> **NOTE on VPC Endpoints and VPC Endpoint Security Group Associations:** Terraform provides
both a standalone VPC Endpoint Security Group Association (an association between a VPC endpoint
and a single `security_group_id`) and a [VPC Endpoint](vpc_endpoint.html) resource with a `security_group_ids`
attribute. Do not use the same security group ID in both a VPC Endpoint resource and a VPC Endpoint Security
Group Association resource. Doing so will cause a conflict of associations and will overwrite the association.

## Example Usage

Basic usage:

```terraform
resource "aws_vpc_endpoint_security_group_association" "sg_ec2" {
  vpc_endpoint_id   = aws_vpc_endpoint.ec2.id
  security_group_id = aws_security_group.sg.id
}
```

## Argument Reference

This resource supports the following arguments:

* `security_group_id` - (Required) The ID of the security group to be associated with the VPC endpoint.
* `vpc_endpoint_id` - (Required) The ID of the VPC endpoint with which the security group will be associated.
* `replace_default_association` - (Optional) Whether this association should replace the association with the VPC's default security group that is created when no security groups are specified during VPC endpoint creation. At most 1 association per-VPC endpoint should be configured with `replace_default_association = true`. `false` should be used when importing resources.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The ID of the association.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import VPC Endpoint Security Group Associations using `vpc_endpoint_id` together with `security_group_id`. For example:

```terraform
import {
  to = aws_vpc_endpoint_security_group_association.example
  id = "vpce-aaaaaaaa/sg-bbbbbbbbbbbbbbbbb"
}
```

Using `terraform import`, import VPC Endpoint Security Group Associations using `vpc_endpoint_id` together with `security_group_id`. For example:

```console
% terraform import aws_vpc_endpoint_security_group_association.example vpce-aaaaaaaa/sg-bbbbbbbbbbbbbbbbb
```
