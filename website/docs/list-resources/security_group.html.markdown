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

### Multiple Filters

```terraform
list "aws_security_group" "example" {
  provider = aws

  filter {
    name   = "vpc-id"
    values = [aws_vpc.main.id]
  }

  filter {
    name   = "group-name"
    values = ["app-*", "web-*"]
  }

  filter {
    name   = "owner-id"
    values = ["123456789012"]
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `group_ids` - (Optional) List of security group IDs to filter results. If specified, only security groups with the provided IDs will be returned.
* `region` - (Optional) Region to query. Defaults to provider region.

### filter Configuration Block

The `filter` block supports the following:

* `name` - (Required) Name of the filter field. Valid values can be found in the [EC2 DescribeSecurityGroups API documentation](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeSecurityGroups.html).
* `values` - (Required) Set of values for the filter.

#### Common Filter Examples

* `vpc-id` - Filter by VPC ID
* `group-name` - Filter by security group name (supports wildcards)
* `description` - Filter by security group description
* `ip-permission.cidr` - Filter by CIDR range in ingress rules
* `owner-id` - Filter by AWS account ID
