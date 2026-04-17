---
subcategory: "EC2 (Elastic Compute Cloud)"
layout: "aws"
page_title: "AWS: aws_nat_gateway"
description: |-
  Lists EC2 (Elastic Compute Cloud) NAT Gateway resources.
---

# List Resource: aws_nat_gateway

Lists EC2 (Elastic Compute Cloud) NAT Gateway resources.

By default, NAT Gateways in terminal states (`deleting` and `deleted`) are excluded.

## Example Usage

### Basic Usage

```terraform
list "aws_nat_gateway" "example" {
  provider = aws
}
```

### Filter Usage

This example returns NAT Gateways in the `deleted` state.

```terraform
list "aws_nat_gateway" "example" {
  provider = aws

  config {
    filter {
      name   = "state"
      values = ["deleted"]
    }
  }
}
```

This example returns NAT Gateways in a specific subnet.

```terraform
list "aws_nat_gateway" "example" {
  provider = aws

  config {
    filter {
      name   = "subnet-id"
      values = [aws_subnet.example.id]
    }
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `filter` - (Optional) One or more filters to apply to the search.
  If multiple `filter` blocks are provided, they all must be true.
  For a full reference of filter names, see [describe-nat-gateways in the AWS CLI reference][describe-nat-gateways].
  See [`filter` Block](#filter-block) below.
* `nat_gateway_ids` - (Optional) List of NAT Gateway IDs to query.
* `region` - (Optional) [Region](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints) to query.
  Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

### `filter` Block

The `filter` block supports the following arguments:

* `name` - (Required) Name of the filter.
  For a full reference of filter names, see [describe-nat-gateways in the AWS CLI reference][describe-nat-gateways].
* `values` - (Required) One or more values to match.

[describe-nat-gateways]: http://docs.aws.amazon.com/cli/latest/reference/ec2/describe-nat-gateways.html
