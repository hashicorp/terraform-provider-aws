---
subcategory: "Outposts (EC2)"
layout: "aws"
page_title: "AWS: aws_ec2_coip_pool"
description: |-
    Provides details about a specific EC2 Customer-Owned IP Pool
---

# Data Source: aws_ec2_coip_pool

Provides details about a specific EC2 Customer-Owned IP Pool.

This data source can prove useful when a module accepts a CoIP pool ID as
an input variable and needs to, for example, determine the CIDR block of that
CoIP Pool.

## Example Usage

The following example returns a specific CoIP pool ID

```terraform
variable "coip_pool_id" {}

data "aws_ec2_coip_pool" "selected" {
  id = var.coip_pool_id
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `local_gateway_route_table_id` - (Optional) Local Gateway Route Table ID assigned to desired CoIP Pool
* `pool_id` - (Optional) ID of the specific CoIP Pool to retrieve.
* `tags` - (Optional) Mapping of tags, each pair of which must exactly match
  a pair on the desired CoIP Pool.

More complex filters can be expressed using one or more `filter` sub-blocks,
which take the following arguments:

* `name` - (Required) Name of the field to filter by, as defined by
  [the underlying AWS API](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeCoipPools.html).
* `values` - (Required) Set of values that are accepted for the given field.
  A CoIP Pool will be selected if any one of the given values matches.

## Attribute Reference

All of the argument attributes except `filter` blocks are also exported as
result attributes. This data source will complete the data by populating
any fields that are not included in the configuration with the data for
the selected CoIP Pool.

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the CoIP pool
* `pool_cidrs` - Set of CIDR blocks in pool

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `read` - (Default `20m`)
