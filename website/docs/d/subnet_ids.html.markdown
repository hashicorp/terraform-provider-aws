---
layout: "aws"
page_title: "AWS: aws_subnet_ids"
sidebar_current: "docs-aws-datasource-subnet-ids"
description: |-
    Provides a list of subnet Ids for a VPC
---

# Data Source: aws_subnet_ids

`aws_subnet_ids` provides a list of ids for a vpc_id

This resource can be useful for getting back a list of subnet ids for a vpc.

## Example Usage

The following shows outputing all cidr blocks for every subnet id in a vpc.

```hcl
data "aws_subnet_ids" "example" {
  vpc_id = "${var.vpc_id}"
}

data "aws_subnet" "example" {
  count = "${length(data.aws_subnet_ids.example.ids)}"
  id    = "${data.aws_subnet_ids.example.ids[count.index]}"
}

output "subnet_cidr_blocks" {
  value = ["${data.aws_subnet.example.*.cidr_block}"]
}
```

The following example retrieves a list of all subnets in a VPC with a custom
tag of `Tier` set to a value of "Private" so that the `aws_instance` resource
can loop through the subnets, putting instances across availability zones.

```hcl
data "aws_subnet_ids" "private" {
  vpc_id = "${var.vpc_id}"

  tags = {
    Tier = "Private"
  }
}

resource "aws_instance" "app" {
  count         = "3"
  ami           = "${var.ami}"
  instance_type = "t2.micro"
  subnet_id     = "${element(data.aws_subnet_ids.private.ids, count.index)}"
}
```

## Argument Reference

* `vpc_id` - (Required) The VPC ID that you want to filter from.

* `filter` - (Optional) Custom filter block as described below.

* `tags` - (Optional) A mapping of tags, each pair of which must exactly match
  a pair on the desired subnets.

More complex filters can be expressed using one or more `filter` sub-blocks,
which take the following arguments:

* `name` - (Required) The name of the field to filter by, as defined by
  [the underlying AWS API](http://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeSubnets.html).
  For example, if matching against tag `Name`, use:

```hcl
data "aws_subnet_ids" "selected" {
  filter {
    name   = "tag:Name"
    values = [""]       # insert values here
  }
}
```

* `values` - (Required) Set of values that are accepted for the given field.
  Subnet IDs will be selected if any one of the given values match.

## Attributes Reference

* `ids` - A list of all the subnet ids found. This data source will fail if none are found.
