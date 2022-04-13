---
subcategory: "EC2 (Elastic Compute Cloud)"
layout: "aws"
page_title: "AWS: aws_ami_launch_permission"
description: |-
  Adds a launch permission to an Amazon Machine Image (AMI).
---

# Resource: aws_ami_launch_permission

Adds a launch permission to an Amazon Machine Image (AMI).

## Example Usage

### AWS Account ID

```terraform
resource "aws_ami_launch_permission" "example" {
  image_id   = "ami-12345678"
  account_id = "123456789012"
}
```

### Public Access

```terraform
resource "aws_ami_launch_permission" "example" {
  image_id = "ami-12345678"
  group    = "all"
}
```

## Argument Reference

The following arguments are supported:

* `account_id` - (Optional) The AWS account ID for the launch permission.
* `group` - (Optional) The name of the group for the launch permission. Valid values: `"all"`.
* `image_id` - (Required) The ID of the AMI.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Launch permission ID.

## Import

AMI Launch Permissions can be imported using `ACCOUNT-ID/IMAGE-ID`, e.g.,

```sh
$ terraform import aws_ami_launch_permission.example 123456789012/ami-12345678
```

or `group/GROUP-NAME/IMAGE-ID`, e.g.,

```sh
$ terraform import aws_ami_launch_permission.example group/all/ami-12345678
```