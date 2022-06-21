---
subcategory: "Kendra"
layout: "aws"
page_title: "AWS: aws_kendra_query_suggestions_block_list"
description: |-
  Terraform resource for managing an AWS Kendra block list used for query suggestions for an index
---

# Resource: aws_kendra_query_suggestions_block_list

Terraform resource for managing an AWS Kendra block list used for query suggestions for an index.

## Example Usage

### Basic Usage

```terraform
resource "aws_kendra_query_suggestions_block_list" "example" {
  index_id = aws_kendra_index.example.id
  name     = "Example"
  role_arn = aws_iam_role.example.arn

  source_s3_path {
    bucket = aws_s3_bucket.example.id
    key    = "example/suggestions.txt"
  }

  tags = {
    Name = "Example Kendra Index"
  }
}
```

## Argument Reference

The following arguments are required:

* `index_id`- (Required, Forces new resource) The identifier of the index for a block list.
* `name` - (Required) The name for the block list.
* `role_arn` - (Required) The IAM (Identity and Access Management) role used to access the block list text file in S3.
* `source_s3_path` - (Required) The S3 path where your block list text file sits in S3. Detailed below.

The `source_s3_path` configuration block supports the following arguments:

* `bucket` - (Required) The name of the S3 bucket that contains the file.
* `key` - (Required) The name of the file.

The following arguments are optional:

* `description` - (Optional) The description for a block list.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://www.terraform.io/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - ARN of the block list.
* `query_suggestions_block_list_id` - The unique indentifier of the block list.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Timeouts

`aws_kendra_query_suggestions_block_list` provides the following [Timeouts](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts) configuration options:

* `create` - (Optional, Default: `30m`)
* `update` - (Optional, Default: `30m`)
* `delete` - (Optional, Default: `30m`)

## Import

`aws_kendra_query_suggestions_block_list` can be imported using the unique identifiers of the block list and index separated by a slash (`/`), e.g.,

```
$ terraform import aws_kendra_query_suggestions_block_list.example blocklist-123456780/idx-8012925589
```
