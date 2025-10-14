---
subcategory: "EC2 (Elastic Compute Cloud)"
layout: "aws"
page_title: "AWS: aws_ec2_public_ipv4_pools"
description: |-
  Terraform data source for getting information about AWS EC2 Public IPv4 Pools.
---

# Data Source: aws_ec2_public_ipv4_pools

Terraform data source for getting information about AWS EC2 Public IPv4 Pools.

## Example Usage

### Basic Usage

```terraform
# Returns all public IPv4 pools.
data "aws_ec2_public_ipv4_pools" "example" {}
```

### Usage with Filter

```terraform
data "aws_ec2_public_ipv4_pools" "example" {
  filter {
    name   = "tag-key"
    values = ["ExampleTagKey"]
  }
}
```

## Argument Reference

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `filter` - (Optional) Custom filter block as described below.
* `tags` - (Optional) Map of tags, each pair of which must exactly match a pair on the desired pools.

More complex filters can be expressed using one or more `filter` sub-blocks,
which take the following arguments:

* `name` - (Required) Name of the field to filter by, as defined by [the underlying AWS API](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribePublicIpv4Pools.html).
* `values` - (Required) Set of values that are accepted for the given field. Pool IDs will be selected if any one of the given values match.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `pool_ids` - List of all the pool IDs found.
