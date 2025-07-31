---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_nat_gateway_eip_association"
description: |-
  Terraform resource for managing an AWS VPC NAT Gateway EIP Association.
---
# Resource: aws_nat_gateway_eip_association

Terraform resource for managing an AWS VPC NAT Gateway EIP Association.

!> **WARNING:** You should not use the `aws_nat_gateway_eip_association` resource in conjunction with an [`aws_nat_gateway`](aws_nat_gateway.html) resource that has `secondary_allocation_ids` configured. Doing so may cause perpetual differences, and result in associations being overwritten.

## Example Usage

### Basic Usage

```terraform
resource "aws_nat_gateway_eip_association" "example" {
  allocation_id  = aws_eip.example.id
  nat_gateway_id = aws_nat_gateway.example.id
}
```

## Argument Reference

The following arguments are required:

* `allocation_id` - (Required) The ID of the Elastic IP Allocation to associate with the NAT Gateway.
* `nat_gateway_id` - (Required) The ID of the NAT Gateway to associate the Elastic IP Allocation to.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This resource exports no additional attributes.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `10m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import VPC NAT Gateway EIP Association using the `nat_gateway_id,allocation_id`. For example:

```terraform
import {
  to = aws_nat_gateway_eip_association.example
  id = "nat-1234567890abcdef1,eipalloc-1234567890abcdef1"
}
```

Using `terraform import`, import VPC NAT Gateway EIP Association using the `nat_gateway_id,allocation_id`. For example:

```console
% terraform import aws_nat_gateway_eip_association.example nat-1234567890abcdef1,eipalloc-1234567890abcdef1
```
