---
subcategory: "Outposts (EC2)"
layout: "aws"
page_title: "AWS: aws_ec2_coip_pools"
description: |-
    Provides information for multiple EC2 Customer-Owned IP Pools
---

# Data Source: aws_ec2_coip_pools

Provides information for multiple EC2 Customer-Owned IP Pools, such as their identifiers.

## Example Usage

The following shows outputting all COIP Pool Ids.

```terraform
data "aws_ec2_coip_pools" "foo" {}

output "foo" {
  value = data.aws_ec2_coip_pools.foo.ids
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Mapping of tags, each pair of which must exactly match
  a pair on the desired aws_ec2_coip_pools.
* `filter` - (Optional) Custom filter block as described below.

More complex filters can be expressed using one or more `filter` sub-blocks,
which take the following arguments:

* `name` - (Required) Name of the field to filter by, as defined by
  [the underlying AWS API](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeCoipPools.html).
* `values` - (Required) Set of values that are accepted for the given field.
  A COIP Pool will be selected if any one of the given values matches.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - AWS Region.
* `pool_ids` - Set of COIP Pool Identifiers

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `read` - (Default `20m`)
