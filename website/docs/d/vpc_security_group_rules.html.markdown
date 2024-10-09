---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_vpc_security_group_rules"
description: |-
    Get information about a set of security group rules.
---

# Data Source: aws_vpc_security_group_rules

This resource can be useful for getting back a set of security group rule IDs.

## Example Usage

```terraform
data "aws_vpc_security_group_rules" "example" {
  filter {
    name   = "group-id"
    values = [var.security_group_id]
  }
}
```

## Argument Reference

* `filter` - (Optional) Custom filter block as described below.
* `tags` - (Optional) Map of tags, each pair of which must exactly match
  a pair on the desired security group rule.

More complex filters can be expressed using one or more `filter` sub-blocks,
which take the following arguments:

* `name` - (Required) Name of the field to filter by, as defined by
  [the underlying AWS API](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeSecurityGroupRules.html).
* `values` - (Required) Set of values that are accepted for the given field.
  Security group rule IDs will be selected if any one of the given values match.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `ids` - List of all the security group rule IDs found.
