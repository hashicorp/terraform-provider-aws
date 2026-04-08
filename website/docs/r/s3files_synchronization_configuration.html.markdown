---
subcategory: "S3 Files"
layout: "aws"
page_title: "AWS: aws_s3files_synchronization_configuration"
description: |-
  Terraform resource for managing an Amazon S3 Files Synchronization Configuration.
---

# Resource: aws_s3files_synchronization_configuration

Terraform resource for managing an Amazon S3 Files Synchronization Configuration.

## Example Usage

### Basic Usage

```terraform
resource "aws_s3files_synchronization_configuration" "example" {
  file_system_id                  = aws_s3files_file_system.example.file_system_id
  expiration_days_after_last_access = 30
  import_size_less_than           = 131072
}
```

## Argument Reference

The following arguments are required:

* `file_system_id` - (Required, Forces new resource) The ID of the file system.
* `expiration_days_after_last_access` - (Required) The number of days after last access before cached data expires from the file system (1-365).
* `import_size_less_than` - (Required) The upper size limit in bytes for imported files. Only objects smaller than this value will have data imported into the file system.

The following arguments are optional:

* `import_prefix` - (Optional) The S3 key prefix that scopes the import rule.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import S3 Files Synchronization Configuration using the `file_system_id`. For example:

```terraform
import {
  to = aws_s3files_synchronization_configuration.example
  id = "fs-0123456789abcdef0"
}
```

Using `terraform import`, import S3 Files Synchronization Configuration using the `file_system_id`. For example:

```console
% terraform import aws_s3files_synchronization_configuration.example fs-0123456789abcdef0
```
