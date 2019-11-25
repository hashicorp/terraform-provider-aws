---
subcategory: "Workspaces"
layout: "aws"
page_title: "AWS: aws_workspaces_directory"
sidebar_current: "docs-aws-resource-workspaces-directory"
description: |-
  Provides a directory registration in AWS Workspaces Service.
---

# Resource: aws_workspaces_directory

Provides a directory registration in AWS Workspaces Service

## Example Usage

```hcl
resource "aws_vpc" "main" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "private-a" {
  vpc_id = "${aws_vpc.main.id}"
  availability_zone = "us-east-1a"
  cidr_block = "10.0.0.0/24"
}

resource "aws_subnet" "private-b" {
  vpc_id = "${aws_vpc.main.id}"
  availability_zone = "us-east-1b"
  cidr_block = "10.0.1.0/24"
}

resource "aws_directory_service_directory" "main" {
  name = "corp.example.com"
  password = "#S1ncerely"
  size = "Small"
  vpc_settings {
    vpc_id = "${aws_vpc.main.id}"
    subnet_ids = ["${aws_subnet.private-a.id}","${aws_subnet.private-b.id}"]
  }
}

resource "aws_workspaces_directory" "main" {
  directory_id = "${aws_directory_service_directory.main.id}"
}
```

## Argument Reference

The following arguments are supported:

* `directory_id` - (Required) The directory identifier for registration in Workspaces service.
* `subnet_ids` - (Optional) The identifiers of the subnets where the directory resides new workspaces.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The workspaces directory identifier.

## Import

Workspaces directory can be imported using the directory ID, e.g.

```
$ terraform import aws_workspaces_directory.main d-4444444444
```
