---
subcategory: "S3 Files"
layout: "aws"
page_title: "AWS: aws_s3files_file_system"
description: |-
  Get information on an S3 Files File System.
---

# Data Source: aws_s3files_file_system

Get information on an S3 Files File System.

## Example Usage

```terraform
data "aws_s3files_file_system" "example" {
  id = "fs-1234567890abcdef0"
}
```

## Argument Reference

The following arguments are required:

* `id` - (Required) Identifier of the file system.

The following arguments are optional:

* `region` - (Optional) Region where this data source will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the file system.
* `bucket` - S3 bucket ARN.
* `creation_time` - Creation time.
* `kms_key_id` - KMS key ID for encryption.
* `name` - File system name.
* `owner_id` - AWS account ID of the owner.
* `prefix` - S3 bucket prefix.
* `role_arn` - IAM role ARN for S3 access.
* `status` - File system status.
* `status_message` - Status message.
* `tags` - Map of tags assigned to the resource.
