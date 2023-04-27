---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_ec2_managed_prefix_lists"
description: |-
    Get information on managed prefix lists
---

# Data Source: aws_ec2_managed_prefix_lists

This resource can be useful for getting back a list of managed prefix list ids to be referenced elsewhere.

## Example Usage

The following returns all managed prefix lists filtered by tags

```terraform
data "aws_ec2_managed_prefix_lists" "test_env" {
  tags = {
    Env = "test"
  }
}

data "aws_ec2_managed_prefix_list" "test_env" {
  count = length(data.aws_ec2_managed_prefix_lists.test_env.ids)
  id    = tolist(data.aws_ec2_managed_prefix_lists.test_env.ids)[count.index]
}
```

## Argument Reference

* `filter` - (Optional) Custom filter block as described below.
* `tags` - (Optional) Map of tags, each pair of which must exactly match
  a pair on the desired .

More complex filters can be expressed using one or more `filter` sub-blocks,
which take the following arguments:

* `name` - (Required) Name of the field to filter by, as defined by
  [the underlying AWS API](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeManagedPrefixLists.html).
* `values` - (Required) Set of values that are accepted for the given field.
  A managed prefix list will be selected if any one of the given values matches.

## Attributes Reference

* `id` - AWS Region.
* `ids` - List of all the managed prefix list ids found.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `read` - (Default `20m`)
