---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_subnets"
description: |-
    Get information about a set of subnets.
---

# Data Source: aws_subnets

This resource can be useful for getting back a set of subnet IDs.

## Example Usage

The following shows outputing all CIDR blocks for every subnet ID in a VPC.

```terraform
data "aws_subnets" "example" {
  filter {
    name   = "vpc-id"
    values = [var.vpc_id]
  }
}

data "aws_subnet" "example" {
  for_each = toset(data.aws_subnets.example.ids)
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
data "aws_subnets" "private" {
  filter {
    name   = "vpc-id"
    values = [var.vpc_id]
  }

  tags = {
    Tier = "Private"
  }
}

resource "aws_instance" "app" {
  for_each      = toset(data.aws_subnets.private.ids)
  ami           = var.ami
  instance_type = "t2.micro"
  subnet_id     = each.value
}
```

## Argument Reference

* `filter` - (Optional) Custom filter block as described below.
* `tags` - (Optional) Map of tags, each pair of which must exactly match
  a pair on the desired subnets.

More complex filters can be expressed using one or more `filter` sub-blocks,
which take the following arguments:

* `name` - (Required) Name of the field to filter by, as defined by
  [the underlying AWS API](http://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeSubnets.html).
  For example, if matching against tag `Name`, use:

```terraform
data "aws_subnets" "selected" {
  filter {
    name   = "tag:Name"
    values = [""] # insert values here
  }
}
```

* `values` - (Required) Set of values that are accepted for the given field.
  Subnet IDs will be selected if any one of the given values match.

## Attributes Reference

* `ids` - List of all the subnet ids found.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `read` - (Default `20m`)
