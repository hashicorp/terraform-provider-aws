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

The following arguments are supported:

* `index_id` - (Required) The identifier of the index that contains the block list.
* `query_suggestions_block_list_id` - (Required) The identifier of the block list.

## Attributes Reference

In addition to all of the arguments above, the following attributes are exported:

* `arn` - The Amazon Resource Name (ARN) of the block list.
* `created_at` - The date-time a block list was created.
* `description` - The description for the block list.
* `error_message` - The error message containing details if there are issues processing the block list.
* `file_size_bytes` - The current size of the block list text file in S3.
* `id` - The unique identifiers of the block list and index separated by a slash (`/`).
* `item_count` - The current number of valid, non-empty words or phrases in the block list text file.
* `name` - The name of the block list.
* `role_arn` - The Amazon Resource Name (ARN) of a role with permission to access the S3 bucket that contains the block list. For more information, see [IAM Roles for Amazon Kendra](https://docs.aws.amazon.com/kendra/latest/dg/iam-roles.html).
* `source_s3_path` - The S3 location of the block list input data. Detailed below.
* `status` - The current status of the block list. When the value is `ACTIVE`, the block list is ready for use.
* `updated_at` - The date and time that the block list was last updated.
* `tags` - Metadata that helps organize the block list you create.

The `source_s3_path` configuration block supports the following attributes:

* `bucket` - The name of the S3 bucket that contains the file.
* `key` - The name of the file.
