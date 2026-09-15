---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_vpc_security_group_egress_rule"
description: |-
  Lists VPC Security Group Egress Rule resources.
---

# List Resource: aws_vpc_security_group_egress_rule

Lists VPC Security Group Egress Rule resources.

## Example Usage

### List All Egress Rules

```terraform
list "aws_vpc_security_group_egress_rule" "example" {
  provider = aws
}
```

### Filter by Security Group

```terraform
list "aws_vpc_security_group_egress_rule" "example" {
  provider = aws

  config {
    filter {
      name   = "group-id"
      values = [aws_security_group.example.id]
    }
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `filter` - (Optional) Custom filter block as described below.
* `region` - (Optional) Region to query. Defaults to provider region.
* `security_group_rule_ids` - (Optional) List of security group rule IDs to retrieve.

### filter

* `name` - (Required) Name of the filter field. Valid values can be found in the [EC2 DescribeSecurityGroupRules API Reference](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeSecurityGroupRules.html).
* `values` - (Required) Set of values for the filter.
