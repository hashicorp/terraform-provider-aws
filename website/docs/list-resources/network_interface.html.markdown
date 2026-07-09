---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_network_interface"
description: |-
  Lists VPC Network Interface resources.
---

# List Resource: aws_network_interface

Lists VPC Network Interface (ENI) resources.

## Example Usage

### Basic Usage

```terraform
list "aws_network_interface" "example" {
  provider = aws
}
```

### Filter Usage

This example returns Network Interfaces in a specific subnet.

```terraform
list "aws_network_interface" "example" {
  provider = aws

  config {
    filter {
      name   = "subnet-id"
      values = ["subnet-12345678"]
    }
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `filter` - (Optional) One or more filters to apply to the search. If multiple `filter` blocks are provided, they all must be true. For a full reference of filter names, see [describe-network-interfaces in the AWS CLI reference](http://docs.aws.amazon.com/cli/latest/reference/ec2/describe-network-interfaces.html).
* `network_interface_ids` - (Optional) List of Network Interface IDs to query.
* `region` - (Optional) [Region](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints) to query. Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

### `filter` Block

The `filter` block supports the following arguments:

* `name` - (Required) Name of the filter. For a full reference of filter names, see [describe-network-interfaces in the AWS CLI reference](http://docs.aws.amazon.com/cli/latest/reference/ec2/describe-network-interfaces.html).
* `values` - (Required) One or more values to match.
