---
subcategory: "Kendra"
layout: "aws"
page_title: "AWS: aws_kendra_thesaurus"
description: |-
  Provides details about a specific Amazon Kendra Thesaurus.
---

# Data Source: aws_kendra_thesaurus

Provides details about a specific Amazon Kendra Thesaurus.

## Example Usage

```hcl
data "aws_kendra_thesaurus" "example" {
  index_id     = "12345678-1234-1234-1234-123456789123"
  thesaurus_id = "87654321-1234-4321-4321-321987654321"
}
```

## Argument Reference

This data source supports the following arguments:

* `index_id` - (Required) Identifier of the index that contains the Thesaurus.
* `thesaurus_id` - (Required) Identifier of the Thesaurus.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Thesaurus.
* `created_at` - Unix datetime that the Thesaurus was created.
* `description` - Description of the Thesaurus.
* `error_message` - When the `status` field value is `FAILED`, this contains a message that explains why.
* `file_size_bytes` - Size of the Thesaurus file in bytes.
* `id` - Unique identifiers of the Thesaurus and index separated by a slash (`/`).
* `name` - Name of the Thesaurus.
* `role_arn` - ARN of a role with permission to access the S3 bucket that contains the Thesaurus. For more information, see [IAM Roles for Amazon Kendra](https://docs.aws.amazon.com/kendra/latest/dg/iam-roles.html).
* `source_s3_path` - S3 location of the Thesaurus input data. Detailed below.
* `status` - Status of the Thesaurus. It is ready to use when the status is `ACTIVE`.
* `synonym_rule_count` - Number of synonym rules in the Thesaurus file.
* `term_count` - Number of unique terms in the Thesaurus file. For example, the synonyms `a,b,c` and `a=>d`, the term count would be 4.
* `updated_at` - Date and time that the Thesaurus was last updated.
* `tags` - Metadata that helps organize the Thesaurus you create.

The `source_s3_path` configuration block supports the following attributes:

* `bucket` - Name of the S3 bucket that contains the file.
* `key` - Name of the file.
