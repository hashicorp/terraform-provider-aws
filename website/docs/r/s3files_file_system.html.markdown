---
subcategory: "S3 Files"
layout: "aws"
page_title: "AWS: aws_s3files_file_system"
description: |-
  Terraform resource for managing an Amazon S3 Files File System.
---

# Resource: aws_s3files_file_system

Terraform resource for managing an Amazon S3 Files File System.

## Example Usage

### Basic Usage

```terraform
resource "aws_s3files_file_system" "example" {
  bucket   = aws_s3_bucket.example.arn
  role_arn = aws_iam_role.example.arn
}
```

## Argument Reference

The following arguments are required:

* `bucket` - (Required, Forces new resource) The ARN of the S3 bucket that will be accessible through the file system.
* `role_arn` - (Required, Forces new resource) The ARN of the IAM role that grants the S3 Files service permission to access the bucket.

The following arguments are optional:

* `accept_bucket_warning` - (Optional, Forces new resource) Set to true to acknowledge and accept any warnings about the bucket configuration.
* `kms_key_id` - (Optional, Forces new resource) The ARN, key ID, or alias of the AWS KMS key to use for encryption.
* `prefix` - (Optional, Forces new resource) An optional prefix within the S3 bucket to scope the file system access.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `file_system_arn` - The ARN of the file system.
* `file_system_id` - The ID of the file system.
* `owner_id` - The AWS account ID of the file system owner.
* `status` - The lifecycle state of the file system.
* `status_message` - Additional information about the file system status.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import S3 Files File System using the `file_system_id`. For example:

```terraform
import {
  to = aws_s3files_file_system.example
  id = "fs-0123456789abcdef0"
}
```

Using `terraform import`, import S3 Files File System using the `file_system_id`. For example:

```console
% terraform import aws_s3files_file_system.example fs-0123456789abcdef0
```
