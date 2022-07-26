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

The following arguments are supported:

* `index_id` - (Required) The identifier of the index that contains the Thesaurus.
* `thesaurus_id` - (Required) The identifier of the Thesaurus.

## Attributes Reference

In addition to all of the arguments above, the following attributes are exported:

* `arn` - The Amazon Resource Name (ARN) of the Thesaurus.
* `created_at` - The Unix datetime that the Thesaurus was created.
* `description` - The description of the Thesaurus.
* `error_message` - When the `status` field value is `FAILED`, this contains a message that explains why.
* `file_size_bytes` - The size of the Thesaurus file in bytes.
* `id` - The unique identifiers of the Thesaurus and index separated by a slash (`/`).
* `name` - Specifies the name of the Thesaurus.
* `role_arn` - The Amazon Resource Name (ARN) of a role with permission to access the S3 bucket that contains the Thesaurus. For more information, see [IAM Roles for Amazon Kendra](https://docs.aws.amazon.com/kendra/latest/dg/iam-roles.html).
* `source_s3_path` - The S3 location of the Thesaurus input data. Detailed below.
* `status` - The status of the Thesaurus. It is ready to use when the status is `ACTIVE`.
* `synonym_rule_count` - The number of synonym rules in the Thesaurus file.
* `term_count` - The number of unique terms in the Thesaurus file. For example, the synonyms `a,b,c` and `a=>d`, the term count would be 4.
* `updated_at` - The date and time that the Thesaurus was last updated.
* `tags` - Metadata that helps organize the Thesaurus you create.

The `source_s3_path` configuration block supports the following attributes:

* `bucket` - The name of the S3 bucket that contains the file.
* `key` - The name of the file.
