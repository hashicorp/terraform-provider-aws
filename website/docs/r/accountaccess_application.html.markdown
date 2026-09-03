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
  identity_source {
    identity_center {
      instance_arn = tolist(data.aws_ssoadmin_instances.example.arns)[0]
    }
  }
}
```

### With Tags

```terraform
resource "aws_accountaccess_application" "example" {
  identity_source {
    identity_center {
      instance_arn = tolist(data.aws_ssoadmin_instances.example.arns)[0]
    }
  }

  tags = {
    Environment = "production"
    ManagedBy   = "terraform"
  }
}
```

## Argument Reference

The following arguments are required:

* `identity_source` - (Required) Identity source for the application. Forces replacement when changed. See [`identity_source` Block](#identity_source-block) below.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Map of tags to assign to the Application. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block), tags with matching keys will overwrite those defined at the provider-level.

### `identity_source` Block

Exactly one argument must be configured.
The `identity_source` block supports:

* `identity_center` - (Optional) IAM Identity Center instance to use as the identity source. See [`identity_center` Block](#identity_center-block) below.

### `identity_center` Block

The `identity_center` block supports:

* `instance_arn` - (Required) ARN of the IAM Identity Center instance.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Application. Used as the resource ID.
* `tags_all` - Map of tags assigned to the Application, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `tenant_id` - Internal tenant identifier returned by the service.

### `identity_source.identity_center` Block

The `identity_source.identity_center` block exports the following attributes in addition to the arguments above:

* `application_arn` - ARN of the IAM Identity Center application created for this account access manager application.

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
