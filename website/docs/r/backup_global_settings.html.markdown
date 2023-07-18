---
subcategory: "Backup"
layout: "aws"
page_title: "AWS: aws_backup_global_settings"
description: |-
  Provides an AWS Backup Global Settings resource.
---

# Resource: aws_backup_global_settings

Provides an AWS Backup Global Settings resource.

## Example Usage

```terraform
resource "aws_backup_global_settings" "test" {
  global_settings = {
    "isCrossAccountBackupEnabled" = "true"
  }
}
```

## Argument Reference

The following arguments are supported:

* `global_settings` - (Required) A list of resources along with the opt-in preferences for the account.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The AWS Account ID.

## Import

Backup Global Settings can be imported using the `id`, e.g.,

```
$ terraform import aws_backup_global_settings.example 123456789012
```
