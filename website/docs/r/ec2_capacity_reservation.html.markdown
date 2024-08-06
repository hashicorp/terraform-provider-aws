---
subcategory: "EC2 (Elastic Compute Cloud)"
layout: "aws"
page_title: "AWS: aws_ec2_capacity_reservation"
description: |-
  Provides an EC2 Capacity Reservation. This allows you to reserve capacity for your Amazon EC2 instances in a specific Availability Zone for any duration.
---

# Resource: aws_ec2_capacity_reservation

Provides an EC2 Capacity Reservation. This allows you to reserve capacity for your Amazon EC2 instances in a specific Availability Zone for any duration.

## Example Usage

```terraform
resource "aws_ec2_capacity_reservation" "default" {
  instance_type     = "t2.micro"
  instance_platform = "Linux/UNIX"
  availability_zone = "eu-west-1a"
  instance_count    = 1
}
```

## Argument Reference

This resource supports the following arguments:

* `availability_zone` - (Required) The Availability Zone in which to create the Capacity Reservation.
* `ebs_optimized` - (Optional) Indicates whether the Capacity Reservation supports EBS-optimized instances.
* `end_date` - (Optional) The date and time at which the Capacity Reservation expires. When a Capacity Reservation expires, the reserved capacity is released and you can no longer launch instances into it. Valid values: [RFC3339 time string](https://tools.ietf.org/html/rfc3339#section-5.8) (`YYYY-MM-DDTHH:MM:SSZ`)
* `end_date_type` - (Optional) Indicates the way in which the Capacity Reservation ends. Specify either `unlimited` or `limited`.
* `ephemeral_storage` - (Optional) Indicates whether the Capacity Reservation supports instances with temporary, block-level storage.
* `instance_count` - (Required) The number of instances for which to reserve capacity.
* `instance_match_criteria` - (Optional) Indicates the type of instance launches that the Capacity Reservation accepts. Specify either `open` or `targeted`.
* `instance_platform` - (Required) The type of operating system for which to reserve capacity. Valid options are `Linux/UNIX`, `Red Hat Enterprise Linux`, `SUSE Linux`, `Windows`, `Windows with SQL Server`, `Windows with SQL Server Enterprise`, `Windows with SQL Server Standard` or `Windows with SQL Server Web`.
* `instance_type` - (Required) The instance type for which to reserve capacity.
* `outpost_arn` - (Optional) The Amazon Resource Name (ARN) of the Outpost on which to create the Capacity Reservation.
* `placement_group_arn` - (Optional) The Amazon Resource Name (ARN) of the cluster placement group in which to create the Capacity Reservation.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `tenancy` - (Optional) Indicates the tenancy of the Capacity Reservation. Specify either `default` or `dedicated`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The Capacity Reservation ID.
* `owner_id` - The ID of the AWS account that owns the Capacity Reservation.
* `arn` - The ARN of the Capacity Reservation.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block)

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `10m`)
- `update` - (Default `10m`)
- `delete` - (Default `10m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Capacity Reservations using the `id`. For example:

```terraform
import {
  to = aws_ec2_capacity_reservation.web
  id = "cr-0123456789abcdef0"
}
```

Using `terraform import`, import Capacity Reservations using the `id`. For example:

```console
% terraform import aws_ec2_capacity_reservation.web cr-0123456789abcdef0
```
