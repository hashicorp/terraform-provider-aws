---
layout: "aws"
page_title: "AWS: aws_dax_subnet_group"
sidebar_current: "docs-aws-resource-dax-subnet-group"
description: |-
  Provides an DAX Subnet Group resource.
---

# aws\_dax\_subnet\_group

Provides a DynamoDB Accelerator (DAX) Subnet Group resource.

## Example Usage

```hcl
resource "aws_vpc" "foo" {
  cidr_block = "10.0.0.0/16"

  tags {
    Name = "tf-test"
  }
}

resource "aws_subnet" "foo" {
  vpc_id            = "${aws_vpc.foo.id}"
  cidr_block        = "10.0.0.0/24"
  availability_zone = "us-west-2a"

  tags {
    Name = "tf-test"
  }
}

resource "aws_dax_subnet_group" "bar" {
  name       = "tf-test-subnet"
  subnet_ids = ["${aws_subnet.foo.id}"]
}
```

## Argument Reference

The following arguments are supported:

* `name` – (Required) Name for the subnet group. DAX converts this name to lowercase.
* `description` – (Optional) Description for the subnet group. Defaults to "Managed by Terraform".
* `subnet_ids` – (Required) List of VPC Subnet IDs for the subnet group

## Attributes Reference

The following attributes are exported:

* `description`
* `name`
* `subnet_ids`

## Import

DAX Subnet Groups can be imported using the `name`, e.g.

```
$ terraform import aws_dax_subnet_group.bar tf-test-subnet
```