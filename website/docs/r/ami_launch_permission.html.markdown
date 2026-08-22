---
subcategory: "EC2 (Elastic Compute Cloud)"
layout: "aws"
page_title: "AWS: aws_ami_launch_permission"
description: |-
  Adds a launch permission to an Amazon Machine Image (AMI).
---

# Resource: aws_ami_launch_permission

Adds a launch permission to an AMI.

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

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `account_id` - (Optional) AWS account ID for the launch permission.
* `group` - (Optional) Name of the group for the launch permission. Valid values: `"all"`.
* `image_id` - (Required) ID of the AMI.
* `organization_arn` - (Optional) ARN of an organization for the launch permission.
* `organizational_unit_arn` - (Optional) ARN of an organizational unit for the launch permission.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Launch permission ID.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_ami_launch_permission.example
  identity = {
    image_id                     = "ami-12345678"
    launch_permission_account_id = "123456789012"
  }
}

resource "aws_ami_launch_permission" "example" {
  # Configuration omitted for brevity
}
```

### Identity Schema

#### Required

* `image_id` (String) ID of the AMI.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `group` (String) Name of the group for the launch permission.
* `launch_permission_account_id` (String) AWS account ID for the launch permission.
* `organization_arn` (String) ARN of an organization for the launch permission.
* `organizational_unit_arn` (String) ARN of an organizational unit for the launch permission.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import AMI Launch Permissions using `[ACCOUNT-ID|GROUP-NAME|ORGANIZATION-ARN|ORGANIZATIONAL-UNIT-ARN]/IMAGE-ID`. For example:

```terraform
import {
  to = aws_ami_launch_permission.example
  id = "123456789012/ami-12345678"
}
```

Using `terraform import`, import AMI Launch Permissions using `[ACCOUNT-ID|GROUP-NAME|ORGANIZATION-ARN|ORGANIZATIONAL-UNIT-ARN]/IMAGE-ID`. For example:

```console
% terraform import aws_ami_launch_permission.example 123456789012/ami-12345678
```
