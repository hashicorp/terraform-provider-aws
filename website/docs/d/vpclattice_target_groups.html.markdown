---
subcategory: "VPC Lattice"
layout: "aws"
page_title: "AWS: aws_vpclattice_target_groups"
description: |-
  Provides a list of VPC Lattice Target Group Ids in a region.
---

# Data Source: aws_vpclattice_target_groups

This resource can be useful for getting back a list of VPC Lattice Target Group Ids for a region.

## Example Usage

### Basic Usage

```terraform
data "aws_vpclattice_target_groups" "example" {}
```

### Get target group ports using a name prefix

```terraform
data "aws_vpclattice_target_groups" "example" {
  name_prefix = "prefix-"
}

data "aws_vpclattice_target_group" "example" {
  for_each   = toset(data.aws_vpclattice_target_groups.example.ids)
  identifier = each.value
}

output "target_group_ports" {
  value = {for k, v in data.aws_vpclattice_target_group.example : k => v.config[0].port}
}
```

## Argument Reference

The arguments of this data source act as filters for querying the available VPC lattice target groups.

* `name_prefix` - (Optional) Name prefix of the desired target groups.
* `tags` - (Optional) A Map of tags that exist on the desired target groups.
* `type` - (Optional) Type of the desired target groups. Can be one of either `IP`, `LAMBDA`, `INSTANCE` or `ALB`.
* `vpc_id` - (Optional) ID of the VPC the desired target groups are associated with.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `ids` - List of target group ids.