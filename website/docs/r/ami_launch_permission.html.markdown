---
subcategory: "EC2"
layout: "aws"
page_title: "AWS: aws_ami_launch_permission"
description: |-
  Adds launch permission to Amazon Machine Image (AMI).
---

# Resource: aws_ami_launch_permission

Adds launch permission to Amazon Machine Image (AMI) from another AWS account.

## Example Usage

```terraform
resource "aws_ami_launch_permission" "example" {
  image_id   = "ami-12345678"
  account_id = "123456789012"
}
```

```terraform
resource "aws_ami_launch_permission" "example" {
  image_id   = "ami-12345678"
  group_name = "all"
}
```

## Argument Reference

The following arguments are supported:

* `image_id` - (required) A region-unique name for the AMI.

One of the following launch permission arguments must be supplied:

* `account_id` - (Optional) An AWS Account ID to grant launch permissions.
* `group_name` - (Optional) The name of the group to grant launch permissions.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - A combination of "`image_id`-account-`account_id`" or "`image_id`-group-`group_name`".

## Import

AWS AMI Launch Permission can be imported using the `ACCOUNT-ID/IMAGE-ID` or `group/GROUP-NAME/IMAGE-ID`, , e.g.

```sh
$ terraform import aws_ami_launch_permission.example 123456789012/ami-12345678
```

```sh
$ terraform import aws_ami_launch_permission.example group/all/ami-12345678
```
