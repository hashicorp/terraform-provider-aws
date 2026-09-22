---
subcategory: "EC2 (Elastic Compute Cloud)"
layout: "aws"
page_title: "AWS: aws_default_security_group"
description: |-
  Lists EC2 Default Security Group resources.
---

# List Resource: aws_default_security_group

Lists EC2 Default Security Group resources. Only the default security group of each VPC (or the default group of an EC2-Classic account) is returned.

## Example Usage

### Basic Usage

```terraform
list "aws_default_security_group" "example" {
  provider = aws
}
```

### With Filters

```terraform
list "aws_default_security_group" "example" {
  provider = aws

  filter {
    name   = "vpc-id"
    values = ["vpc-12345678"]
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.

### filter Configuration Block

The `filter` block supports the following:

* `name` - (Required) Name of the filter field. Valid values can be found in the [EC2 DescribeSecurityGroups API documentation](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeSecurityGroups.html).
* `values` - (Required) Set of values for the filter.
