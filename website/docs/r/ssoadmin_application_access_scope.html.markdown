---
subcategory: "SSO Admin"
layout: "aws"
page_title: "AWS: aws_ssoadmin_application_access_scope"
description: |-
  Terraform resource for managing an AWS SSO Admin Application Access Scope.
---
# Resource: aws_ssoadmin_application_access_scope

Terraform resource for managing an AWS SSO Admin Application Access Scope.

## Example Usage

### Basic Usage

```terraform
data "aws_ssoadmin_instances" "example" {}

resource "aws_ssoadmin_application" "example" {
  name                     = "example"
  application_provider_arn = "arn:aws:sso::aws:applicationProvider/custom"
  instance_arn             = tolist(data.aws_ssoadmin_instances.example.arns)[0]
}

resource "aws_ssoadmin_application_access_scope" "example" {
  application_arn    = aws_ssoadmin_application.example.application_arn
  authorized_targets = ["arn:aws:sso::123456789012:application/ssoins-123456789012/apl-123456789012"]
  scope              = "sso:account:access"
}
```

## Argument Reference

The following arguments are required:

* `application_arn` - (Required) Specifies the ARN of the application with the access scope with the targets to add or update.
* `scope` - (Required) Specifies the name of the access scope to be associated with the specified targets.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `authorized_targets` - (Optional) Specifies an array list of ARNs that represent the authorized targets for this access scope.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - A comma-delimited string concatenating `application_arn` and `scope`.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import SSO Admin Application Access Scope using the `id`. For example:

```terraform
import {
  to = aws_ssoadmin_application_access_scope.example
  id = "arn:aws:sso::123456789012:application/ssoins-123456789012/apl-123456789012,sso:account:access"
}
```

Using `terraform import`, import SSO Admin Application Access Scope using the `id`. For example:

```console
% terraform import aws_ssoadmin_application_access_scope.example arn:aws:sso::123456789012:application/ssoins-123456789012/apl-123456789012,sso:account:access
```
