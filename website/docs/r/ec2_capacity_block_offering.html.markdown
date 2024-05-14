---
subcategory: "EC2 (Elastic Compute Cloud)"
layout: "aws"
page_title: "AWS: aws_ec2_capacity_block_offering"
description: |-
  Information about a single EC2 Capacity Block Reservation.
---

# Data Source: aws_ec2_capacity_block_offering

Information about a single EC2 Capacity Block Reservation.

## Example Usage

```terraform


data "aws_ec2_capacity_block_offering" "test" {
    capacity_duration = 24
    end_date          = "2024-05-30T15:04:05Z"
    instance_count    = 1
    instance_platform = "Linux/UNIX"
    instance_type     = "p4d.24xlarge"
    start_date        = "2024-04-28T15:04:05Z"
}

```

## Argument Reference

This resource supports the following arguments:

* `capacity_duration` - (Required) The amount of time of the Capacity Block reservation in hours.
* `end_date` - (Optional) The date and time at which the Capacity Block Reservation expires. When a Capacity Reservation expires, the reserved capacity is released and you can no longer launch instances into it. Valid values: [RFC3339 time string](https://tools.ietf.org/html/rfc3339#section-5.8) (`YYYY-MM-DDTHH:MM:SSZ`)
* `instance_count` - (Required) The number of instances for which to reserve capacity.
* `instance_type` - (Required) The instance type for which to reserve capacity.
* `start_date` - (Optional) The date and time at which the Capacity Block Reservation starts. Valid values: [RFC3339 time string](https://tools.ietf.org/html/rfc3339#section-5.8) (`YYYY-MM-DDTHH:MM:SSZ`)

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `availability_zone` - The Availability Zone in which to create the Capacity Reservation.
* `currency_code` - The currency of the payment for the Capacity Block.
* `id` - The Capacity Block Reservation ID.
* `instance_platform` - The type of operating system for which to reserve capacity. Valid options are `Linux/UNIX`, `Red Hat Enterprise Linux`, `SUSE Linux`, `Windows`, `Windows with SQL Server`, `Windows with SQL Server Enterprise`, `Windows with SQL Server Standard` or `Windows with SQL Server Web`.
* `upfront_fee` - The total price to be paid up front.
* `tenancy` - Indicates the tenancy of the Capacity Reservation. Specify either `default` or `dedicated`.
