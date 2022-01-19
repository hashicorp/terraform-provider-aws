---
subcategory: "VPC"
layout: "aws"
page_title: "AWS: aws_subnet_ids_categorized"
description: |-
    Provides categorized into public and private sets of subnet Ids for a VPC
---

# Data Source: aws_subnet_ids_categorized

`aws_subnet_ids_categorized` provides four sets of subnet ids for a given VPC. The subnets are categorized as public, private (*all* subnets not routed to the internet gateway), private-routed (subnets routed to a NAT gateway), and private-isolated (subnets not routed to a NAT gateway).

This resource can be useful for determining the public and private subnets for a VPC when you are adding resources to a VPC *not* built by the current configuration.

The data source works by examining the VPC's route tables and route table associations to determine which route tables contain a route to an internet gateway and/or routes to NAT gateways. Thus if you do use this data source in the same configuration that is creating the network, you should add a `depends_on` to the data source that will ensure that all AWS resources that will be queried by the data source (IGW, NAT, Subnets, Route Tables) are ready.

## Example Usage

The following shows outputting all cidr blocks for every public subnet id in a vpc.

```terraform
data "aws_subnet_ids_categorized" "example" {
  vpc_id = var.vpc_id
}

data "aws_subnet" "example" {
  for_each = data.aws_subnet_ids_categorized.example.public_subnet_ids
  id       = each.value
}

output "subnet_cidr_blocks" {
  value = [for s in data.aws_subnet.example : s.cidr_block]
}
```

The following example retrieves a set of all private subnets in a VPC so that the `aws_instance` resource
can loop through the routed private subnets, putting instances across availability zones and giving them the ability to talk to the Internet.

```terraform
data "aws_subnet_ids_categorized" "example" {
  vpc_id = var.vpc_id
}

resource "aws_instance" "app" {
  for_each      = data.aws_subnet_ids_categorized.example.private_subnet_routed_ids
  ami           = var.ami
  instance_type = "t2.micro"
  subnet_id     = each.value
}
```

## Argument Reference

* `vpc_id` - (Required) The VPC ID that you want to filter from.

## Attributes Reference

* `public_subnet_ids` - A set of IDs of all the public subnets found. This set may be empty.

* `private_subnet_ids` - A set of IDs of all the private subnets found. This is the complete set of all private subnets, whether or not routed to a NAT gateway. This set may be empty.

* `private_subnet_routed_ids` - A set of IDs of all the private subnets that have a route to a NAT gateway. This set may be empty.

* `private_subnet_isolated_ids` - A set of IDs of all the private subnets that have no route to a NAT gateway. This set is logically `private_subnet-ids` minus `private_subnet_routed_ids`. This set may be empty.
