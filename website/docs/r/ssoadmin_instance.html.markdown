---
subcategory: "SSO Admin"
layout: "aws"
page_title: "AWS: aws_ssoadmin_instance"
description: |-
  Manages an IAM Identity Center instance.
---

# Resource: aws_ssoadmin_instance

Manages an IAM Identity Center instance (formerly AWS Single Sign-On). IAM Identity Center supports one organization instance or one account instance per AWS account.

~> **NOTE:** Creating an instance fails when the AWS account already has an IAM Identity Center instance. Import an existing instance instead.

~> **WARNING:** Destroying this resource deletes the IAM Identity Center instance and can remove access to assigned AWS accounts and applications.

## Example Usage

### Basic Usage

```terraform
resource "aws_ssoadmin_instance" "example" {
  name = "example"

  tags = {
    Environment = "production"
  }
}
```

### Customer Managed KMS Key

```terraform
resource "aws_ssoadmin_instance" "example" {
  name = "example"

  encryption_configuration {
    key_type   = "CUSTOMER_MANAGED_KEY"
    kms_key_arn = aws_kms_key.example.arn
  }
}
```

## Argument Reference

The following arguments are optional:

* `client_token` - (Optional, Forces new resource) Unique, case-sensitive identifier to ensure idempotency of the request. Must be between 1 and 64 characters.
* `encryption_configuration` - (Optional) Encryption configuration for data at rest. See [Encryption Configuration](#encryption-configuration).
* `name` - (Optional) Name of the IAM Identity Center instance. Must be between 1 and 32 characters.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block), tags with matching keys will overwrite those defined at the provider-level.

### `encryption_configuration` Block

* `key_type` - (Required) Type of KMS key. Valid values are `AWS_OWNED_KMS_KEY` and `CUSTOMER_MANAGED_KEY`.
* `kms_key_arn` - (Optional) ARN of the KMS key. Required when `key_type` is `CUSTOMER_MANAGED_KEY` and must not be specified when `key_type` is `AWS_OWNED_KMS_KEY`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the IAM Identity Center instance.
* `created_date` - Date and time the instance was created.
* `encryption_configuration.encryption_status` - Current encryption configuration status.
* `encryption_configuration.encryption_status_reason` - Additional context for the encryption configuration status.
* `identity_store_id` - Identifier of the connected identity store.
* `owner_account_id` - AWS account ID of the instance owner.
* `status` - Current instance status.
* `status_reason` - Additional context for the instance status.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_ssoadmin_instance.example
  identity = {
    arn = "arn:aws:sso:::instance/ssoins-0123456789abcdef"
  }
}
```

### Identity Schema

#### Required

* `arn` - ARN of the IAM Identity Center instance.

#### Optional

* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import an IAM Identity Center instance using its ARN. For example:

```terraform
import {
  to = aws_ssoadmin_instance.example
  id = "arn:aws:sso:::instance/ssoins-0123456789abcdef"
}
```

Using `terraform import`, import an IAM Identity Center instance using its ARN. For example:

```console
% terraform import aws_ssoadmin_instance.example arn:aws:sso:::instance/ssoins-0123456789abcdef
```
