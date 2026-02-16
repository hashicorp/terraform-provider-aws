---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_subnet"
description: |-
  Lists VPC Subnet resources.
---

# List Resource: aws_subnet

Lists VPC Subnet resources.

Note: Default VPCs are not included.

## Example Usage

### Basic Usage

```terraform
list "aws_subnet" "example" {
  provider = aws
}
```

### Filter Usage

This example will return VPC Subnets with the tag `Project` with the value `example`.

```terraform
list "aws_subnet" "example" {
  provider = aws

  config {
    filter {
      name   = "tag:Project"
      values = ["example"]
    }
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `filter` - (Optional) One or more filters to apply to the search.
  If multiple `filter` blocks are provided, they all must be true.
  For a full reference of filter names, see [describe-subnets in the AWS CLI reference][describe-subnets].
  See [`filter` Block](#filter-block) below.
* `region` - (Optional) [Region](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints) to query.
  Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `subnet_ids` - (Optional) List of VPC Subnets IDs to query.

### `filter` Block

The `filter` block supports the following arguments:

* `name` - (Required) Name of the filter.
  For a full reference of filter names, see [describe-subnets in the AWS CLI reference][describe-subnets].
  `default-for-az` is not supported.
* `values` - (Required) One or more values to match.

[describe-subnets]: http://docs.aws.amazon.com/cli/latest/reference/ec2/describe-subnets.html
