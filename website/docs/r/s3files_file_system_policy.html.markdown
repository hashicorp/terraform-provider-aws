---
subcategory: "S3 Files"
layout: "aws"
page_title: "AWS: aws_s3files_file_system_policy"
description: |-
  Terraform resource for managing an Amazon S3 Files File System Policy.
---

# Resource: aws_s3files_file_system_policy

Terraform resource for managing an Amazon S3 Files File System Policy.

## Example Usage

### Basic Usage

```terraform
resource "aws_s3files_file_system_policy" "example" {
  file_system_id = aws_s3files_file_system.example.file_system_id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Principal = { AWS = "*" }
      Action    = ["s3files:ClientMount"]
      Resource  = aws_s3files_file_system.example.file_system_arn
    }]
  })
}
```

## Argument Reference

The following arguments are required:

* `file_system_id` - (Required, Forces new resource) The ID of the file system.
* `policy` - (Required) The IAM resource policy for the file system.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import S3 Files File System Policy using the `file_system_id`. For example:

```terraform
import {
  to = aws_s3files_file_system_policy.example
  id = "fs-0123456789abcdef0"
}
```

Using `terraform import`, import S3 Files File System Policy using the `file_system_id`. For example:

```console
% terraform import aws_s3files_file_system_policy.example fs-0123456789abcdef0
```
