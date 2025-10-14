---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_vpcs"
description: |-
    Provides a list of VPC Ids in a region
---

# Data Source: aws_vpcs

This resource can be useful for getting back a list of VPC Ids for a region.

The following example retrieves a list of VPC Ids with a custom tag of `service` set to a value of "production".

## Example Usage

The following shows outputting all VPC Ids.

```terraform
data "aws_vpcs" "foo" {
  tags = {
    service = "production"
  }
}

output "foo" {
  value = data.aws_vpcs.foo.ids
}
```

An example use case would be interpolate the `aws_vpcs` output into `count` of an aws_flow_log resource.

```terraform
data "aws_vpcs" "foo" {}

data "aws_vpc" "foo" {
  count = length(data.aws_vpcs.foo.ids)
  id    = tolist(data.aws_vpcs.foo.ids)[count.index]
}

resource "aws_flow_log" "test_flow_log" {
  count = length(data.aws_vpcs.foo.ids)

  # ...
  vpc_id = data.aws_vpc.foo[count.index].id

  # ...
}

output "foo" {
  value = data.aws_vpcs.foo.ids
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Map of tags, each pair of which must exactly match
  a pair on the desired vpcs.
* `filter` - (Optional) Custom filter block as described below.

### `filter`

More complex filters can be expressed using one or more `filter` sub-blocks, which take the following arguments:

* `name` - (Required) Name of the field to filter by, as defined by
  [the underlying AWS API](http://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeVpcs.html).
* `values` - (Required) Set of values that are accepted for the given field.
  A VPC will be selected if any one of the given values matches.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - AWS Region.
* `ids` - List of all the VPC Ids found.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `read` - (Default `20m`)
