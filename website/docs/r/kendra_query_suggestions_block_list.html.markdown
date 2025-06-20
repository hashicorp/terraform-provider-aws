---
subcategory: "Kendra"
layout: "aws"
page_title: "AWS: aws_kendra_query_suggestions_block_list"
description: |-
  Terraform resource for managing an AWS Kendra block list used for query suggestions for an index
---

# Resource: aws_kendra_query_suggestions_block_list

Use the `aws_kendra_index_block_list` resource to manage an AWS Kendra block list used for query suggestions for an index.

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

* `index_id` - (Required, Forces New Resource) Identifier of the index for a block list.
* `name` - (Required) Name for the block list.
* `role_arn` - (Required) IAM (Identity and Access Management) role used to access the block list text file in S3.
* `source_s3_path` - (Required) S3 path where your block list text file is located. See details below.

The `source_s3_path` configuration block supports the following arguments:

* `bucket` - (Required) Name of the S3 bucket that contains the file.
* `key` - (Required) Name of the file.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `description` - (Optional) Description for a block list.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block), tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the block list.
* `query_suggestions_block_list_id` - Unique identifier of the block list.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider's [default_tags configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

Configuration options for operation timeouts can be found [here](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts).

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import the `aws_kendra_query_suggestions_block_list` resource using the unique identifiers of the block list and index separated by a slash (`/`). For example:

```terraform
import {
  to = aws_kendra_query_suggestions_block_list.example
  id = "blocklist-123456780/idx-8012925589"
}
```

Using `terraform import`, import the `aws_kendra_query_suggestions_block_list` resource using the unique identifiers of the block list and index separated by a slash (`/`). For example:

```console
% terraform import aws_kendra_query_suggestions_block_list.example blocklist-123456780/idx-8012925589
```
