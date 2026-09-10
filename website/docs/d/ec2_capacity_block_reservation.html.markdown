---
subcategory: "EC2 (Elastic Compute Cloud)"
layout: "aws"
page_title: "AWS: aws_ec2_capacity_block_reservation"
description: |-
  Information about an existing EC2 Capacity Block reservation.
---

# Data Source: aws_ec2_capacity_block_reservation

Information about an existing EC2 Capacity Block reservation.

This data source returns only Capacity Reservations whose `reservation_type` is `capacity-block`. Use the [`aws_ec2_capacity_reservation`](./ec2_capacity_reservation.html.markdown) data source to look up On-Demand Capacity Reservations (ODCR).

At least one of `id` or `filter` must be specified. Filter combinations that match multiple Capacity Block reservations will return an error.

## Example Usage

### Lookup by ID

```terraform
data "aws_ec2_capacity_block_reservation" "example" {
  id = "cr-0123456789abcdef0"
}
```

### Lookup by filter

```terraform
data "aws_ec2_capacity_block_reservation" "example" {
  filter {
    name   = "instance-type"
    values = ["p4d.24xlarge"]
  }

  filter {
    name   = "state"
    values = ["active"]
  }
}
```

### Lookup by tag

```terraform
data "aws_ec2_capacity_block_reservation" "example" {
  filter {
    name   = "tag:Project"
    values = ["ml-training"]
  }
}
```

## Argument Reference

This data source supports the following arguments:

* `filter` - (Optional) Configuration block(s) for filtering. Detailed below.
* `id` - (Optional) ID of the Capacity Block reservation to retrieve.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

### `filter` Block

The `filter` configuration block supports the following arguments:

* `name` - (Required) Name of the filter field. See the [DescribeCapacityReservations API Reference](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeCapacityReservations.html) for valid values. Common filters include `instance-type`, `availability-zone`, `state`, `instance-platform`, `tenancy`, `outpost-arn`, `placement-group-arn`, `instance-match-criteria`, and `tag:<KEY>`.
* `values` - (Required) Set of values that are accepted for the given filter field. A Capacity Block reservation will be selected if any one of the given values matches.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Capacity Block reservation.
* `availability_zone` - Availability Zone in which the capacity is reserved.
* `availability_zone_id` - ID of the Availability Zone in which the capacity is reserved.
* `available_instance_count` - Remaining capacity, indicating the number of instances that can still be launched into the Capacity Block reservation.
* `capacity_block_id` - ID of the underlying Capacity Block.
* `commitment_info` - Information about your commitment for a future-dated Capacity Block reservation. See [`commitment_info` Attribute Reference](#commitment_info-attribute-reference) below.
* `created_date` - Date and time the Capacity Block reservation was created in [RFC3339 format](https://tools.ietf.org/html/rfc3339#section-5.8).
* `delivery_preference` - Delivery method for a future-dated Capacity Block reservation. Either `fixed` or `incremental`.
* `ebs_optimized` - Whether the Capacity Block reservation supports EBS-optimized instances.
* `end_date` - Date and time the Capacity Block reservation expires in [RFC3339 format](https://tools.ietf.org/html/rfc3339#section-5.8).
* `end_date_type` - End type of the Capacity Block reservation. Either `limited` or `unlimited`.
* `instance_count` - Total number of instances for which the Capacity Block reserves capacity.
* `instance_match_criteria` - Type of instance launches that the Capacity Block accepts. Either `open` or `targeted`.
* `instance_platform` - Operating system platform for which the Capacity Block reserves capacity.
* `instance_type` - Instance type for which the Capacity Block reserves capacity.
* `interruptible_capacity_allocation` - Information about the interruptible capacity allocation, if applicable. See [`interruptible_capacity_allocation` Attribute Reference](#interruptible_capacity_allocation-attribute-reference) below.
* `interruption_info` - Information about an interrupted Capacity Block reservation, if applicable. See [`interruption_info` Attribute Reference](#interruption_info-attribute-reference) below.
* `outpost_arn` - ARN of the Outpost on which the Capacity Block was created, if applicable.
* `owner_id` - ID of the AWS account that owns the Capacity Block reservation.
* `placement_group_arn` - ARN of the cluster placement group in which the Capacity Block was created, if applicable.
* `reservation_type` - Type of Capacity Reservation. Always `capacity-block` for this data source.
* `start_date` - Date and time the Capacity Block reservation was started in [RFC3339 format](https://tools.ietf.org/html/rfc3339#section-5.8).
* `state` - Current state of the Capacity Block reservation. One of `active`, `expired`, `cancelled`, `pending`, `failed`, `scheduled`, `payment-pending`, `payment-failed`, or `assessing`.
* `tags` - Map of tags assigned to the Capacity Block reservation.
* `tenancy` - Tenancy of the Capacity Block. Either `default` or `dedicated`.

### `commitment_info` Attribute Reference

* `commitment_end_date` - Date and time the commitment duration ends in [RFC3339 format](https://tools.ietf.org/html/rfc3339#section-5.8).
* `committed_instance_count` - Number of instances committed to the Capacity Block reservation.

### `interruptible_capacity_allocation` Attribute Reference

* `instance_count` - Number of instances allocated as interruptible capacity within the Capacity Block reservation.
* `interruptible_capacity_reservation_id` - ID of the interruptible Capacity Reservation associated with this allocation.
* `interruption_type` - Type of interruption that may occur. Either `spot-interruption` or `capacity-block-interruption`.
* `status` - Status of the interruptible capacity allocation. One of `pending`, `confirmed`, or `cancelled`.
* `target_instance_count` - Target number of interruptible instances for the allocation.

### `interruption_info` Attribute Reference

* `interruption_type` - Type of interruption that occurred. Either `spot-interruption` or `capacity-block-interruption`.
* `source_capacity_reservation_id` - ID of the source Capacity Reservation that originally held the capacity, if the reservation was created as a result of an interruption.
