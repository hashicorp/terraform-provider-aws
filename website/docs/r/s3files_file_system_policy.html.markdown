---
subcategory: "S3 Files"
layout: "aws"
page_title: "AWS: aws_s3files_file_system_policy"
description: |-
  Manages an S3 Files File System Policy.
---

# Resource: aws_s3files_file_system_policy

Manages an S3 Files File System Policy.

## Example Usage

```terraform
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}

resource "aws_s3files_file_system_policy" "example" {
  file_system_id = aws_s3files_file_system.example.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          AWS = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"
        }
        Action   = "s3files:ClientMount"
        Resource = "*"
      }
    ]
  })
}
```

## Argument Reference

The following arguments are required:

* `file_system_id` - (Required) File system ID. Changing this value forces replacement.
* `policy` - (Required) JSON policy document.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_s3files_file_system_policy.example
  identity = {
    file_system_id = "fs-1234567890abcdef0"
  }
}

resource "aws_s3files_file_system_policy" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `file_system_id` - File system ID.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import S3 Files File System Policy using the file system ID. For example:

```terraform
import {
  to = aws_s3files_file_system_policy.example
  id = "fs-1234567890abcdef0"
}
```

Using `terraform import`, import S3 Files File System Policy using `file_system_id`. For example:

```console
% terraform import aws_s3files_file_system_policy.example fs-1234567890abcdef0
```
