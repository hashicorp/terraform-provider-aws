---
subcategory: "EFS (Elastic File System)"
layout: "aws"
page_title: "AWS: aws_efs_backup_policy"
description: |-
  Provides an Elastic File System (EFS) Backup Policy resource.
---

# Resource: aws_efs_backup_policy

Provides an Elastic File System (EFS) Backup Policy resource.
Backup policies turn automatic backups on or off for an existing file system.

## Example Usage

```terraform
resource "aws_efs_file_system" "fs" {
  creation_token = "my-product"
}

resource "aws_efs_backup_policy" "policy" {
  file_system_id = aws_efs_file_system.fs.id

  backup_policy {
    status = "ENABLED"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `file_system_id` - (Required) The ID of the EFS file system.
* `backup_policy` - (Required) A backup_policy object (documented below).

### Backup Policy Arguments

`backup_policy` supports the following arguments:

* `status` - (Required) A status of the backup policy. Valid values: `ENABLED`, `DISABLED`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The ID that identifies the file system (e.g., fs-ccfc0d65).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import the EFS backup policies using the `id`. For example:

```terraform
import {
  to = aws_efs_backup_policy.example
  id = "fs-6fa144c6"
}
```

Using `terraform import`, import the EFS backup policies using the `id`. For example:

```console
% terraform import aws_efs_backup_policy.example fs-6fa144c6
```
