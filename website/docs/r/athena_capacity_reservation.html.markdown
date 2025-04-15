---
subcategory: "Athena"
layout: "aws"
page_title: "AWS: aws_athena_capacity_reservation"
description: |-
  Terraform resource for managing an AWS Athena Capacity Reservation.
---
# Resource: aws_athena_capacity_reservation

Terraform resource for managing an AWS Athena Capacity Reservation.

~> Destruction of this resource will both [cancel](https://docs.aws.amazon.com/athena/latest/ug/capacity-management-cancelling-a-capacity-reservation.html) and [delete](https://docs.aws.amazon.com/athena/latest/ug/capacity-management-deleting-a-capacity-reservation.html) the capacity reservation.

## Example Usage

### Basic Usage

```terraform
resource "aws_athena_capacity_reservation" "example" {
  name        = "example-reservation"
  target_dpus = 24
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the capacity reservation.
* `target_dpus` - (Required) Number of data processing units requested. Must be at least `24` units.

The following arguments are optional:

* `tags` - (Optional) Map of tags assigned to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `allocated_dpus` - Number of data processing units currently allocated.
* `arn` - ARN of the Capacity Reservation.
* `status` - Status of the capacity reservation.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Athena Capacity Reservation using the `name`. For example:

```terraform
import {
  to = aws_athena_capacity_reservation.example
  id = "example-reservation"
}
```

Using `terraform import`, import Athena Capacity Reservation using the `name`. For example:

```console
% terraform import aws_athena_capacity_reservation.example example-reservation
```
