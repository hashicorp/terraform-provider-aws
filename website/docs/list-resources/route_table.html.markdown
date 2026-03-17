---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_route_table"
description: |-
  Lists Route Table resources.
---

# List Resource: aws_route_table

Lists Route Table resources.

## Example Usage

### Basic Usage

```terraform
list "aws_route_table" "example" {
  provider = aws
}
```

### Filter Usage

This example will return route tables associated with a specific VPC.

```terraform
list "aws_route_table" "example" {
  provider = aws

  config {
    filter {
      name   = "vpc-id"
      values = ["vpc-12345678"]
    }
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `filter` - (Optional) One or more filters to apply to the search.
  If multiple `filter` blocks are provided, they all must be true.
  For a full reference of filter names, see [describe-route-tables in the AWS CLI reference][describe-route-tables].
  See [`filter` Block](#filter-block) below.
* `region` - (Optional) [Region](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints) to query.
  Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `route_table_ids` - (Optional) List of Route Table IDs to query.

### `filter` Block

The `filter` block supports the following arguments:

* `name` - (Required) Name of the filter.
  For a full reference of filter names, see [describe-route-tables in the AWS CLI reference][describe-route-tables].
* `values` - (Required) One or more values to match.

[describe-route-tables]: http://docs.aws.amazon.com/cli/latest/reference/ec2/describe-route-tables.html
