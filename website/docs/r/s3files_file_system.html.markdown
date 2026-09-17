---
subcategory: "S3 Files"
layout: "aws"
page_title: "AWS: aws_s3files_file_system"
description: |-
  Manages an S3 Files File System.
---

# Resource: aws_s3files_file_system

Manages an S3 Files File System.

## Example Usage

```terraform
resource "aws_s3files_file_system" "example" {
  bucket   = aws_s3_bucket.example.arn
  role_arn = aws_iam_role.example.arn
}
```

## Argument Reference

The following arguments are required:

* `bucket` - (Required) S3 bucket ARN. Changing this value forces replacement.
* `role_arn` - (Required) IAM role ARN for S3 access. Changing this value forces replacement.

The following arguments are optional:

* `accept_bucket_warning` - (Optional) Set to `true` to acknowledge and accept any warnings related to the bucket configuration. If not specified, the operation may fail when such warnings are present. For example, warnings may be raised when creating a file system scoped to a prefix containing a large number of objects (approximately 12 million objects). See [the AWS documentation](https://docs.aws.amazon.com/AmazonS3/latest/userguide/s3-files-synchronization.html#s3-files-sync-rename-move) for more details.
* `kms_key_id` - (Optional) KMS key ID for encryption. Changing this value forces replacement.
* `prefix` - (Optional) S3 bucket prefix. Changing this value forces replacement.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Map of tags assigned to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the file system.
* `creation_time` - Creation time.
* `id` - Identifier of the file system.
* `name` - File system name.
* `owner_id` - AWS account ID of the owner.
* `status` - File system status.
* `status_message` - Status message.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `10m`)
* `delete` - (Default `10m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_s3files_file_system.example
  identity = {
    id = "fs-1234567890abcdef0"
  }
}

resource "aws_s3files_file_system" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `id` - Identifier of the file system.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import S3 Files File System using the resource ID. For example:

```terraform
import {
  to = aws_s3files_file_system.example
  id = "fs-1234567890abcdef0"
}
```

Using `terraform import`, import S3 Files File System using `id`. For example:

```console
% terraform import aws_s3files_file_system.example fs-1234567890abcdef0
```
