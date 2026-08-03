---
subcategory: "Backup"
layout: "aws"
page_title: "AWS: aws_backup_global_settings"
description: |-
  Provides an AWS Backup Global Settings resource.
---

# Resource: aws_backup_global_settings

Provides an AWS Backup Global Settings resource.

~> **Note:** This resource will show perpetual differences for any supported settings not explicitly configured in the `global_settings` configuration block. To avoid this, specify all supported options with their default values (typically `"false"`, but check the plan diff for the actual value). See [UpdateGlobalSettings](https://docs.aws.amazon.com/aws-backup/latest/devguide/API_UpdateGlobalSettings.html) in the AWS Backup Developer Guide for available settings.

## Example Usage

```terraform
resource "aws_backup_global_settings" "test" {
  global_settings = {
    "isCrossAccountBackupEnabled"     = "true"
    "isMpaEnabled"                    = "false"
    "isDelegatedAdministratorEnabled" = "false"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `global_settings` - (Required) A list of resources along with the opt-in preferences for the account. For a list of inputs, see [UpdateGlobalSettings](https://docs.aws.amazon.com/aws-backup/latest/devguide/API_UpdateGlobalSettings.html) in the AWS Backup Developer Guide.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The AWS Account ID.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Backup Global Settings using the `id`. For example:

```terraform
import {
  to = aws_backup_global_settings.example
  id = "123456789012"
}
```

Using `terraform import`, import Backup Global Settings using the `id`. For example:

```console
% terraform import aws_backup_global_settings.example 123456789012
```
