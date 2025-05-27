---
subcategory: "SSO Admin"
layout: "aws"
page_title: "AWS: aws_ssoadmin_application_assignment_configuration"
description: |-
  Terraform resource for managing an AWS SSO Admin Application Assignment Configuration.
---
# Resource: aws_ssoadmin_application_assignment_configuration

Terraform resource for managing an AWS SSO Admin Application Assignment Configuration.

By default, applications will require users to have an explicit assignment in order to access an application.
This resource can be used to adjust this default behavior if necessary.

~> Deleting this resource will return the assignment configuration for the application to the default AWS behavior (ie. `assignment_required = true`).

## Example Usage

### Basic Usage

```terraform
resource "aws_ssoadmin_application_assignment_configuration" "example" {
  application_arn     = aws_ssoadmin_application.example.application_arn
  assignment_required = true
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `application_arn` - (Required) ARN of the application.
* `assignment_required` - (Required) Indicates whether users must have an explicit assignment to access the application. If `false`, all users have access to the application.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - ARN of the application.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import SSO Admin Application Assignment Configuration using the `id`. For example:

```terraform
import {
  to = aws_ssoadmin_application_assignment_configuration.example
  id = "arn:aws:sso::123456789012:application/id-12345678"
}
```

Using `terraform import`, import SSO Admin Application Assignment Configuration using the `id`. For example:

```console
% terraform import aws_ssoadmin_application_assignment_configuration.example arn:aws:sso::123456789012:application/id-12345678
```
