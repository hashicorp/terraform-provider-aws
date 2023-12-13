---
subcategory: "Kendra"
layout: "aws"
page_title: "AWS: aws_kendra_query_suggestions_block_list"
description: |-
  Provides details about a specific Amazon Kendra block list used for query suggestions for an index.
---

# Data Source: aws_kendra_query_suggestions_block_list

Provides details about a specific Amazon Kendra block list used for query suggestions for an index.

## Example Usage

```hcl
data "aws_kendra_query_suggestions_block_list" "example" {
  index_id                        = "12345678-1234-1234-1234-123456789123"
  query_suggestions_block_list_id = "87654321-1234-4321-4321-321987654321"
}
```

## Argument Reference

This data source supports the following arguments:

* `index_id` - (Required) Identifier of the index that contains the block list.
* `query_suggestions_block_list_id` - (Required) Identifier of the block list.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the block list.
* `created_at` - Date-time a block list was created.
* `description` - Description for the block list.
* `error_message` - Error message containing details if there are issues processing the block list.
* `file_size_bytes` - Current size of the block list text file in S3.
* `id` - Unique identifiers of the block list and index separated by a slash (`/`).
* `item_count` - Current number of valid, non-empty words or phrases in the block list text file.
* `name` - Name of the block list.
* `role_arn` - ARN of a role with permission to access the S3 bucket that contains the block list. For more information, see [IAM Roles for Amazon Kendra](https://docs.aws.amazon.com/kendra/latest/dg/iam-roles.html).
* `source_s3_path` - S3 location of the block list input data. Detailed below.
* `status` - Current status of the block list. When the value is `ACTIVE`, the block list is ready for use.
* `updated_at` - Date and time that the block list was last updated.
* `tags` - Metadata that helps organize the block list you create.

The `source_s3_path` configuration block supports the following attributes:

* `bucket` - Name of the S3 bucket that contains the file.
* `key` - Name of the file.
