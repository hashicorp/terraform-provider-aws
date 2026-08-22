---
subcategory: "Account Access"
layout: "aws"
page_title: "AWS: aws_accountaccess_application"
description: |-
  Manages an AWS Account Access Application bound to an IAM Identity Center instance.
---

# Resource: aws_accountaccess_application

Manages an AWS Account Access Application bound to an IAM Identity Center instance.

~> **NOTE:** AWS Account Access allows only one Application per Identity Center instance. Attempting to create a second Application for the same instance will fail. Use `terraform import` to bring an existing Application under management.

## Example Usage

### Basic Usage

```terraform
data "aws_ssoadmin_instances" "example" {}

resource "aws_accountaccess_application" "example" {
  identity_center_instance_arn = tolist(data.aws_ssoadmin_instances.example.arns)[0]
}
```

### With Tags

```terraform
resource "aws_accountaccess_application" "example" {
  identity_center_instance_arn = tolist(data.aws_ssoadmin_instances.example.arns)[0]

  tags = {
    Environment = "production"
    ManagedBy   = "terraform"
  }
}
```

## Argument Reference

The following arguments are required:

* `identity_center_instance_arn` - (Required) ARN of the IAM Identity Center instance to bind this Application to. Forces replacement when changed.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Map of tags to assign to the Application. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block), tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Application. Used as the resource ID.
* `created_at` - Date and time, in [RFC3339 format](https://datatracker.ietf.org/doc/html/rfc3339), when the Application was created.
* `id` - ARN of the Application. Same as `arn`.
* `identity_center_application_arn` - ARN of the IAM Identity Center Application that Account Access provisioned for this resource.
* `status` - Current lifecycle status. One of `CREATE_IN_PROGRESS`, `ACTIVE`, `DELETE_IN_PROGRESS`, `CREATE_FAILED`, `DELETE_FAILED`.
* `tags_all` - Map of tags assigned to the Application, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `tenant_id` - Internal tenant identifier returned by the service.
* `updated_at` - Date and time, in [RFC3339 format](https://datatracker.ietf.org/doc/html/rfc3339), when the Application was last updated.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `15m`)
* `delete` - (Default `15m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_accountaccess_application.example
  identity = {
    arn = "arn:aws:account-access:us-east-1:123456789012:application/aam-0123456789abcdef"
  }
}
```

### Identity Schema

#### Required

* `arn` (String) ARN of the Account Access Application.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Account Access Applications using the Application ARN. For example:

```terraform
import {
  to = aws_accountaccess_application.example
  id = "arn:aws:account-access:us-east-1:123456789012:application/aam-0123456789abcdef"
}
```

Using `terraform import`, import Account Access Applications using the Application ARN. For example:

```console
% terraform import aws_accountaccess_application.example arn:aws:account-access:us-east-1:123456789012:application/aam-0123456789abcdef
```
