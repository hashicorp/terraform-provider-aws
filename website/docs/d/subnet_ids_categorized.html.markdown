---
subcategory: "VPC"
layout: "aws"
page_title: "AWS: aws_subnet_ids_categorized"
description: |-
    Provides categorized into public and private sets of subnet Ids for a VPC
---

# Data Source: aws_subnet_ids_categorized

`aws_subnet_ids_categorized` provides two sets of ids (public and private) for a vpc_id

This resource can be useful for determining the public and private subnets for a vpc.

## Example Usage

The following shows outputing all cidr blocks for every public subnet id in a vpc.

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
can loop through the subnets, putting instances across availability zones.

```terraform
data "aws_subnet_ids_categorized" "example" {
  vpc_id = var.vpc_id
}

resource "aws_instance" "app" {
  for_each      = data.aws_subnet_ids_categorized.example.private_subnet_ids
  ami           = var.ami
  instance_type = "t2.micro"
  subnet_id     = each.value
}
```

## Argument Reference

* `vpc_id` - (Required) The VPC ID that you want to filter from.

## Attributes Reference

* `public_subnet_ids` - A set of all the public subnet ids found. This set may be empty.

* `private_subnet_ids` - A set of all the private subnet ids found. This set may be empty.
