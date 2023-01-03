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

### Organization Access

```terraform
data "aws_organizations_organization" "current" {}

resource "aws_ami_launch_permission" "example" {
  image_id         = "ami-12345678"
  organization_arn = data.aws_organizations_organization.current.arn
}
```

## Argument Reference

The following arguments are supported:

* `account_id` - (Optional) AWS account ID for the launch permission.
* `group` - (Optional) Name of the group for the launch permission. Valid values: `"all"`.
* `image_id` - (Required) ID of the AMI.
* `organization_arn` - (Optional) ARN of an organization for the launch permission.
* `organizational_unit_arn` - (Optional) ARN of an organizational unit for the launch permission.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Launch permission ID.

## Import

AMI Launch Permissions can be imported using `[ACCOUNT-ID|GROUP-NAME|ORGANIZATION-ARN|ORGANIZATIONAL-UNIT-ARN]/IMAGE-ID`, e.g.,

```sh
$ terraform import aws_ami_launch_permission.example 123456789012/ami-12345678
```
