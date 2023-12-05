---
subcategory: "SSO Admin"
layout: "aws"
page_title: "AWS: aws_ssoadmin_application_assignment"
description: |-
  Terraform resource for managing an AWS SSO Admin Application Assignment.
---
# Resource: aws_ssoadmin_application_assignment

Terraform resource for managing an AWS SSO Admin Application Assignment.

## Example Usage

### Basic Usage

```terraform
resource "aws_ssoadmin_application_assignment" "example" {
  application_arn = aws_ssoadmin_application.example.application_arn
  principal_id    = aws_identitystore_user.example.user_id
  principal_type  = "USER"
}
```

### Group Type

```terraform
resource "aws_ssoadmin_application_assignment" "example" {
  application_arn = aws_ssoadmin_application.example.application_arn
  principal_id    = aws_identitystore_group.example.group_id
  principal_type  = "GROUP"
}
```

## Argument Reference

The following arguments are required:

* `application_arn` - (Required) ARN of the application provider.
* `principal_id` - (Required) An identifier for an object in IAM Identity Center, such as a user or group.
* `principal_type` - (Required) Entity type for which the assignment will be created. Valid values are `USER` or `GROUP`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - A comma-delimited string concatenating `application_arn`, `principal_id`, and `principal_type`.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import SSO Admin Application Assignment using the `id`. For example:

```terraform
import {
  to = aws_ssoadmin_application_assignment.example
  id = "arn:aws:sso::012345678901:application/id-12345678,abcd1234,USER"
}
```

Using `terraform import`, import SSO Admin Application Assignment using the `id`. For example:

```console
% terraform import aws_ssoadmin_application_assignment.example arn:aws:sso::012345678901:application/id-12345678,abcd1234,USER
```
