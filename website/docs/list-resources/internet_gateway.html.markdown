---
subcategory: "EC2 (Elastic Compute Cloud)"
layout: "aws"
page_title: "AWS: aws_internet_gateway"
description: |-
  Lists EC2 (Elastic Compute Cloud) Internet Gateway resources.
---

# List Resource: aws_internet_gateway

Lists EC2 (Elastic Compute Cloud) Internet Gateway resources.

## Example Usage

### Basic Usage

```terraform
list "aws_internet_gateway" "example" {
  provider = aws
}
```

### Filter Usage

This example returns Internet Gateways attached to a specific VPC.

```terraform
list "aws_internet_gateway" "example" {
  provider = aws

  config {
    filter {
      name   = "attachment.vpc-id"
      values = [aws_vpc.example.id]
    }
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `filter` - (Optional) One or more filters to apply to the search.
  If multiple `filter` blocks are provided, they all must be true.
  For a full reference of filter names, see [describe-internet-gateways in the AWS CLI reference][describe-internet-gateways].
  See [`filter` Block](#filter-block) below.
* `internet_gateway_ids` - (Optional) List of Internet Gateway IDs to query.
* `region` - (Optional) [Region](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints) to query.
  Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

### `filter` Block

The `filter` block supports the following arguments:

* `name` - (Required) Name of the filter.
  For a full reference of filter names, see [describe-internet-gateways in the AWS CLI reference][describe-internet-gateways].
* `values` - (Required) One or more values to match.

[describe-internet-gateways]: https://docs.aws.amazon.com/cli/latest/reference/ec2/describe-internet-gateways.html
