---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_vpc"
description: |-
  Lists VPC resources.
---

# List Resource: aws_vpc

Lists VPC resources.

Note: The default VPC is not included.

## Example Usage

### Basic Usage

```terraform
list "aws_vpc" "example" {
  provider = aws
}
```

### Filter Usage

This example will return VPCs with the tag `Project` with the value `example`.

```terraform
list "aws_vpc" "example" {
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
  For a full reference of filter names, see [describe-vpcs in the AWS CLI reference][describe-vpcs].
  See [`filter` Block](#filter-block) below.
* `region` - (Optional) [Region](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints) to query.
  Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `vpc_ids` - (Optional) List of VPC IDs to query.

### `filter` Block

The `filter` block supports the following arguments:

* `name` - (Required) Name of the filter.
  For a full reference of filter names, see [describe-vpcs in the AWS CLI reference][describe-vpcs].
  `is-default` is not supported.
* `values` - (Required) One or more values to match.

[describe-vpcs]: http://docs.aws.amazon.com/cli/latest/reference/ec2/describe-vpcs.html
