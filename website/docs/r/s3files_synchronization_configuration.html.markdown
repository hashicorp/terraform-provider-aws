---
subcategory: "S3 Files"
layout: "aws"
page_title: "AWS: aws_s3files_synchronization_configuration"
description: |-
  Manages an S3 Files Synchronization configuration.
---

# Resource: aws_s3files_synchronization_configuration

Manages an S3 Files Synchronization configuration.

## Example Usage

```terraform
resource "aws_s3files_synchronization_configuration" "example" {
  file_system_id = aws_s3files_file_system.example.id

  import_data_rule {
    prefix         = ""
    size_less_than = 52673613135872
    trigger        = "ON_FILE_ACCESS"
  }

  expiration_data_rule {
    days_after_last_access = 30
  }
}
```

## Argument Reference

The following arguments are required:

* `file_system_id` - (Required) File system ID. Changing this value forces replacement.
* `import_data_rule` - (Required) One or more import data rules. See [`import_data_rule`](#import_data_rule) below.

The following arguments are optional:

* `expiration_data_rule` - (Optional) Expiration data rule configuration. See [`expiration_data_rule`](#expiration_data_rule) below.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

### import_data_rule

* `prefix` - (Required) S3 key prefix to apply this rule to. Use `""` for all objects.
* `size_less_than` - (Required) Maximum object size in bytes to import.
* `trigger` - (Required) Import trigger. Valid values: `ON_FILE_ACCESS`.

### expiration_data_rule

* `days_after_last_access` - (Required) Number of days after last access before expiring data.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `latest_version_number` - Latest synchronization configuration version number.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_s3files_synchronization_configuration.example
  identity = {
    file_system_id = "fs-1234567890abcdef0"
  }
}

resource "aws_s3files_synchronization_configuration" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `file_system_id` - File system ID.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import S3 Files Synchronization using the file system ID. For example:

```terraform
import {
  to = aws_s3files_synchronization_configuration.example
  id = "fs-1234567890abcdef0"
}
```

Using `terraform import`, import S3 Files Synchronization using `file_system_id`. For example:

```console
% terraform import aws_s3files_synchronization_configuration.example fs-1234567890abcdef0
```
