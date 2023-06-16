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

The following arguments are supported:

* `cidrBlock` - (Required) The CIDR block for the reservation.
* `reservationType` - (Required) The type of reservation to create. Valid values: `explicit`, `prefix`
* `subnetId` - (Required) The ID of the subnet to create the reservation for.
* `description` - (Optional) A brief description of the reservation.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the CIDR reservation.
* `ownerId` - ID of the AWS account that owns this CIDR reservation.

## Import

Existing CIDR reservations can be imported using `subnetId:reservationId`, e.g.,

```
$ terraform import aws_ec2_subnet_cidr_reservation.example subnet-01llsxvsxabqiymcz:scr-4mnvz6wb7otksjcs9
```

<!-- cache-key: cdktf-0.17.0-pre.15 input-068d78a5a45794a8e8dfb5699785f8dd8fbaf725814c393dab988a122183fd6d -->