---
subcategory: "EC2 (Elastic Compute Cloud)"
layout: "aws"
page_title: "AWS: aws_ec2_secondary_subnet"
description: |-
  Lists EC2 (Elastic Compute Cloud) Secondary Subnet resources.
---

# List Resource: aws_ec2_secondary_subnet

Lists EC2 (Elastic Compute Cloud) Secondary Subnet resources.

## Example Usage

### Basic Usage

```terraform
list "aws_ec2_secondary_subnet" "example" {
  provider = aws
}
```

### Filter Usage

This example will return only Secondary Subnets within the specified Secondary Network.

```terraform
list "aws_subnet" "example" {
  provider = aws

  config {
    filter {
      name   = "secondary-network-id"
      values = ["sn-0123456789abcdef0"]
    }
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `filter` - (Optional) One or more filters to apply to the search.
  If multiple `filter` blocks are provided, they all must be true.
  For a full reference of filter names, see [describe-secondary-subnets in the AWS CLI reference][describe-secondary-subnets].
  See [`filter` Block](#filter-block) below.
* `region` - (Optional) Region to query. Defaults to provider region.

### `filter` Block

The `filter` block supports the following arguments:

* `name` - (Required) Name of the filter.
  For a full reference of filter names, see [describe-secondary-subnets in the AWS CLI reference][describe-secondary-subnets].
  `default-for-az` is not supported.
* `values` - (Required) One or more values to match.

[describe-secondary-subnets]: https://docs.aws.amazon.com/cli/latest/reference/ec2/describe-secondary-subnets.html
