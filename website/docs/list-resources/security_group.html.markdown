---
subcategory: "EC2 (Elastic Compute Cloud)"
layout: "aws"
page_title: "AWS: aws_security_group"
description: |-
  Lists EC2 (Elastic Compute Cloud) Security Group resources.
---

# List Resource: aws_security_group

Lists EC2 (Elastic Compute Cloud) Security Group resources.

## Example Usage

### Basic Usage

```terraform
list "aws_security_group" "example" {
  provider = aws
}
```

### With Filters

```terraform
list "aws_security_group" "example" {
  provider = aws

  filter {
    name   = "vpc-id"
    values = ["vpc-12345678"]
  }

  filter {
    name   = "group-name"
    values = ["my-security-group*"]
  }
}
```

### Filter by Security Group IDs

```terraform
list "aws_security_group" "example" {
  provider = aws

  group_ids = ["sg-12345678", "sg-87654321"]
}
```

## Argument Reference

This list resource supports the following arguments:

* `group_ids` - (Optional) List of security group IDs to filter results.
* `filter` - (Optional) One or more name/value pairs to filter off of. See the [EC2 API documentation](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeSecurityGroups.html) for supported filters.
* `region` - (Optional) Region to query. Defaults to provider region.
