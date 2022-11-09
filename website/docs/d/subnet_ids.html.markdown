---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_subnet_ids"
description: |-
    Provides a set of subnet Ids for a VPC
---

# Data Source: aws_subnet_ids

`aws_subnet_ids` provides a set of ids for a vpc_id

This resource can be useful for getting back a set of subnet ids for a vpc.

~> **NOTE:** The `aws_subnet_ids` data source has been deprecated and will be removed in a future version. Use the [`aws_subnets`](subnets.html) data source instead.

## Example Usage

The following shows outputing all cidr blocks for every subnet id in a vpc.

```terraform
data "aws_subnet_ids" "example" {
  vpc_id = var.vpc_id
}

data "aws_subnet" "example" {
  for_each = data.aws_subnet_ids.example.ids
  id       = each.value
}

output "subnet_cidr_blocks" {
  value = [for s in data.aws_subnet.example : s.cidr_block]
}
```

The following example retrieves a set of all subnets in a VPC with a custom
tag of `Tier` set to a value of "Private" so that the `aws_instance` resource
can loop through the subnets, putting instances across availability zones.

```terraform
data "aws_subnet_ids" "private" {
  vpc_id = var.vpc_id

  tags = {
    Tier = "Private"
  }
}

resource "aws_instance" "app" {
  for_each      = data.aws_subnet_ids.private.ids
  ami           = var.ami
  instance_type = "t2.micro"
  subnet_id     = each.value
}
```

## Argument Reference

* `vpc_id` - (Required) VPC ID that you want to filter from.

* `filter` - (Optional) Custom filter block as described below.

* `tags` - (Optional) Map of tags, each pair of which must exactly match
  a pair on the desired subnets.

More complex filters can be expressed using one or more `filter` sub-blocks,
which take the following arguments:

* `name` - (Required) Name of the field to filter by, as defined by
  [the underlying AWS API](http://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeSubnets.html).
  For example, if matching against tag `Name`, use:

```terraform
data "aws_subnet_ids" "selected" {
  filter {
    name   = "tag:Name"
    values = [""] # insert values here
  }
}
```

* `values` - (Required) Set of values that are accepted for the given field.
  Subnet IDs will be selected if any one of the given values match.

## Attributes Reference

* `ids` - Set of all the subnet ids found. This data source will fail if none are found.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `read` - (Default `20m`)
