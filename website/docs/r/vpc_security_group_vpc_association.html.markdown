---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_vpc_security_group_vpc_association"
description: |-
  Terraform resource for managing Security Group VPC Associations.
---

# Resource: aws_vpc_security_group_vpc_association

Terraform resource for managing Security Group VPC Associations.

## Example Usage

```terraform
resource "aws_vpc_security_group_vpc_association" "example" {
  security_group_id = "sg-05f1f54ab49bb39a3"
  vpc_id            = "vpc-01df9d105095412ba"
}
```

## Argument Reference

The following arguments are required:

* `security_group_id` - (Required) The ID of the security group.
* `vpc_id` - (Required) The ID of the VPC to make the association with.

## Attribute Reference

This resource exports no additional attributes.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `5m`)
* `delete` - (Default `5m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Security Group VPC Association using the `security_group_id` and `vpc_id`. For example:

```terraform
import {
  to = aws_vpc_security_group_vpc_association.example
  id = "security_group_id-12345678:vpc_id-233323"
}
```

Using `terraform import`, import Security Group VPC Association using the `security_group_id` and `vpc_id`. For example:

```console
% terraform import aws_vpc_security_group_vpc_association.example security_group_id-12345678:vpc_id-233323
```
