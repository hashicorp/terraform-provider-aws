---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_internet_gateway_attachment"
description: |-
  Provides a resource to create a VPC Internet Gateway Attachment.
---

# Resource: aws_internet_gateway_attachment

Provides a resource to create a VPC Internet Gateway Attachment.

## Example Usage

```terraform
resource "aws_internet_gateway_attachment" "example" {
  internet_gateway_id = aws_internet_gateway.example.id
  vpc_id              = aws_vpc.example.id
}

resource "aws_vpc" "example" {
  cidr_block = "10.1.0.0/16"
}

resource "aws_internet_gateway" "example" {}
```

## Argument Reference

The following arguments are supported:

* `internet_gateway_id` - (Required) The ID of the internet gateway.
* `vpc_id` - (Required) The ID of the VPC.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the VPC and Internet Gateway separated by a colon.


## Import

Internet Gateway Attachments can be imported using the `id`, e.g.

```
$ terraform import aws_internet_gateway_attachment.example igw-c0a643a9:vpc-123456
```
