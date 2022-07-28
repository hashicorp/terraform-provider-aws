---
subcategory: "RDS (Relational Database)"
layout: "aws"
page_title: "AWS: aws_rds_reserved_instance"
description: |-
  Manages RDS DB Instance Reservations
---

# Resource: aws_rds_reserved_instance

Manages RDS DB Instance Reservations. **Once created, a reservation is valid for the `duration` of the provided `offering_id` and cannot be deleted. Performing a `destroy` or removing this resource from your code will only remove the resource from state.** For more information see the official [RDS Reserved Instances Documentation](https://aws.amazon.com/rds/reserved-instances/)

## Example Usage

```terraform
resource "aws_rds_reserved_instance" "my-reservation" {
  offering_id    = "438012d3-4052-4cc7-b2e3-8d3372e0e706"
  reservation_id = "optionalCustomReservationID"
  instance_count = 3
}
```

## Argument Reference

For more detailed documentation around purchasing an rds reservation, refer to the AWS official documentation [purchase-reserved-db-instances-offering](https://docs.aws.amazon.com/cli/latest/reference/rds/purchase-reserved-db-instances-offering.html)

The following arguments are supported:

* `instance_count` - (Required) The number of instances to reserve.
* `instance_id` - (Required) Customer-specified identifier to track this reservation.
* `offering_id` - (Required) The ID of the Reserved DB instance offering to purchase. To identify the `offering_id` for the preferred instance type, duration, price, etc, use the cli command [describe-reserved-db-instances-offerings](https://docs.aws.amazon.com/cli/latest/reference/rds/describe-reserved-db-instances-offerings.html).
* `tags` - (Optional) A map of tags to assign to the DB reservation. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The Amazon Resource Name (ARN) for the reserved DB instance.
* `id` - The unique identifier for the reservation. same as `instance_id`.
* `currency_code` - The currency code for the reserved DB instance.
* `duration` - The duration of the reservation in seconds.
* `fixed_price` – The fixed price charged for this reserved DB instance.
* `instance_class` - The DB instance class for the reserved DB instance.
* `lease_id` - The unique identifier for the lease associated with the reserved DB instance. Amazon Web Services Support might request the lease ID for an issue related to a reserved DB instance.
* `multi_az` - Indicates if the reservation applies to Multi-AZ deployments.
* `offering_type` - The offering type of this reserved DB instance.
* `product_description` - The description of the reserved DB instance.
* `recurring_charges` - The recurring price charged to run this reserved DB instance.
* `start_time` - The time the reservation started.
* `state` - The state of the reserved DB instance.
* `usage_price` - The hourly price charged for this reserved DB instance.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Import

RDS DB Instance Reservations can be imported using the `instance_id`, e.g.,

```
$ terraform import aws_rds_reserved_instance.reservation_instance CustomReservationID
```
