---
subcategory: "VPC Lattice"
layout: "aws"
page_title: "AWS: aws_vpclattice_target_group"
description: |-
  Terraform data source for managing an AWS VPC Lattice Target Group.
---

# Data Source: aws_vpclattice_target_group

Terraform data source for managing an AWS VPC Lattice Target Group.

## Example Usage

### Basic Usage

```terraform
data "aws_vpclattice_target_group" "example" {
  name = "example"
}
```

## Argument Reference

The arguments of this data source act as filters for querying the available VPC lattice target groups.
The given filters must match exactly one VPC lattice target group whose data will be exported as attributes.

* `name` - (Optional) Target group name.
    Cannot be used with `target_group_identifier`.
* `target_group_identifier` - (Optional) ID or Amazon Resource Name (ARN) of the target group.
    Cannot be used with `name`.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the target group.
* `id` - Unique identifier for the target group.
* `config` - Configuration of the target group.
* `created_at` - Date and time that the target group was created.
* `last_updated_at` - Date and time that the target group was last updated.
* `service_arns` - List of VPC lattice service arns the target group is associated with.
* `status` - Status of the target group. Either `CREATE_IN_PROGRESS`, `ACTIVE`, `DELETE_IN_PROGRESS`, `CREATE_FAILED` or `DELETE_FAILED`
* `type` - Type of the target group. Either `IP`, `LAMBDA`, `INSTANCE` or `ALB`.