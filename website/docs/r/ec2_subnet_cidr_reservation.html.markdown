---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_ec2_subnet_cidr_reservation"
description: |-
  Provides a subnet CIDR reservation resource.
---

# Resource: aws_ec2_subnet_cidr_reservation

Provides a subnet CIDR reservation resource.

## Example Usage

```terraform
resource "aws_ec2_subnet_cidr_reservation" "example" {
  cidr_block       = "10.0.0.16/28"
  reservation_type = "prefix"
  subnet_id        = aws_subnet.example.id
}
```

## Argument Reference

This resource supports the following arguments:

* `cidr_block` - (Required) The CIDR block for the reservation.
* `reservation_type` - (Required) The type of reservation to create. Valid values: `explicit`, `prefix`
* `subnet_id` - (Required) The ID of the subnet to create the reservation for.
* `description` - (Optional) A brief description of the reservation.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - ID of the CIDR reservation.
* `owner_id` - ID of the AWS account that owns this CIDR reservation.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Existing CIDR reservations using `SUBNET_ID:RESERVATION_ID`. For example:

```terraform
import {
  to = aws_ec2_subnet_cidr_reservation.example
  id = "subnet-01llsxvsxabqiymcz:scr-4mnvz6wb7otksjcs9"
}
```

Using `terraform import`, import Existing CIDR reservations using `SUBNET_ID:RESERVATION_ID`. For example:

```console
% terraform import aws_ec2_subnet_cidr_reservation.example subnet-01llsxvsxabqiymcz:scr-4mnvz6wb7otksjcs9
```
