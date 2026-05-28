---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_vpc_security_group_ingress_rule"
description: |-
  Lists VPC Security Group Ingress Rule resources.
---

# List Resource: aws_vpc_security_group_ingress_rule

Lists VPC Security Group Ingress Rule resources.

## Example Usage

### Basic Usage

```terraform
list "aws_vpc_security_group_ingress_rule" "example" {
  provider = aws
}
```

### Filter by Security Group

```terraform
list "aws_vpc_security_group_ingress_rule" "example" {
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

* `filter` - (Optional) One or more filters to apply to the search. If multiple `filter` blocks are provided, they all must be true. See [`filter` Block](#filter-block) below.
* `region` - (Optional) Region to query. Defaults to the Region set in the provider configuration.
* `security_group_rule_ids` - (Optional) Security group rule IDs to query.

### `filter` Block

The `filter` block supports the following arguments:

* `name` - (Required) Name of the filter. For a full reference of filter names, see [describe-security-group-rules in the AWS CLI reference](https://docs.aws.amazon.com/cli/latest/reference/ec2/describe-security-group-rules.html).
* `values` - (Required) One or more values to match.
