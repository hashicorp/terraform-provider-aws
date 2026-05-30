---
subcategory: "S3 Files"
layout: "aws"
page_title: "AWS: aws_s3files_file_systems"
description: |-
  Provides details about S3 Files File Systems.
---

# Data Source: aws_s3files_file_systems

Provides details about S3 Files File Systems.

## Example Usage

```terraform
data "aws_s3files_file_systems" "example" {}
```

## Argument Reference

The following arguments are optional:

* `region` - (Optional) Region where this data source will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `file_systems` - List of file systems. See [`file_systems`](#file_systems-attribute-reference) below.

### `file_systems` Attribute Reference

* `arn` - ARN of the file system.
* `bucket` - S3 bucket ARN.
* `creation_time` - Creation time.
* `id` - Identifier of the file system.
* `kms_key_id` - KMS key ID for encryption.
* `name` - File system name.
* `owner_id` - AWS account ID of the owner.
* `prefix` - S3 bucket prefix.
* `role_arn` - IAM role ARN for S3 access.
* `status` - File system status.
* `status_message` - Status message.
* `tags` - Map of tags assigned to the resource.
