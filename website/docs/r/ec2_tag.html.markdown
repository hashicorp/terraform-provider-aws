---
subcategory: "EC2"
layout: "aws"
page_title: "AWS: aws_ec2_tag"
description: |-
  Manages a single tag for a given EC2 resource
---

# Resource: aws_ec2_tag

Manages an individual tag for a given EC2 resource. This resource should only be used in cases where EC2 resources are created outside Terraform (e.g. AMIs), being shared via Resource Access Manager (RAM), or implicitly created by other means (e.g. Transit Gateway VPN Attachments).

## Example Usage

```hcl
resource "aws_vpc" "example" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_ec2_tag" "example" {
  resource_id = "${aws_vpc.example.id}"
  key         = "Name"
  value       = "Hello World"
}
```

## Argument Reference

The following arguments are supported:

* `resource_id` - (Required) The ID of the EC2 resource to manage the tag for.
* `key` - (Required) The tag name.
* `value` - (Required) The value of the tag.
