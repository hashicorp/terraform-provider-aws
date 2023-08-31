---
subcategory: "Kendra"
layout: "aws"
page_title: "AWS: aws_kendra_faq"
description: |-
  Provides details about a specific Amazon Kendra Faq.
---

# Data Source: aws_kendra_faq

Provides details about a specific Amazon Kendra Faq.

## Example Usage

```hcl
data "aws_kendra_faq" "test" {
  faq_id   = "87654321-1234-4321-4321-321987654321"
  index_id = "12345678-1234-1234-1234-123456789123"
}
```

## Argument Reference

This data source supports the following arguments:

* `faq_id` - (Required) Identifier of the FAQ.
* `index_id` - (Required) Identifier of the index that contains the FAQ.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the FAQ.
* `created_at` - Unix datetime that the faq was created.
* `description` - Description of the FAQ.
* `error_message` - When the `status` field value is `FAILED`, this contains a message that explains why.
* `file_format` - File format used by the input files for the FAQ. Valid Values are `CSV`, `CSV_WITH_HEADER`, `JSON`.
* `id` - Unique identifiers of the FAQ and index separated by a slash (`/`).
* `language_code` - Code for a language. This shows a supported language for the FAQ document. For more information on supported languages, including their codes, see [Adding documents in languages other than English](https://docs.aws.amazon.com/kendra/latest/dg/in-adding-languages.html).
* `name` - Name of the FAQ.
* `role_arn` - ARN of a role with permission to access the S3 bucket that contains the FAQs. For more information, see [IAM Roles for Amazon Kendra](https://docs.aws.amazon.com/kendra/latest/dg/iam-roles.html).
* `s3_path` - S3 location of the FAQ input data. Detailed below.
* `status` - Status of the FAQ. It is ready to use when the status is ACTIVE.
* `updated_at` - Date and time that the FAQ was last updated.
* `tags` - Metadata that helps organize the FAQs you create.

The `s3_path` configuration block supports the following attributes:

* `bucket` - Name of the S3 bucket that contains the file.
* `key` - Name of the file.
