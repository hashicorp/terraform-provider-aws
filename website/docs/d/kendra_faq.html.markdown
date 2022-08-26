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

The following arguments are supported:

* `faq_id` - (Required) The identifier of the FAQ.
* `index_id` - (Required) The identifier of the index that contains the FAQ.

## Attributes Reference

In addition to all of the arguments above, the following attributes are exported:

* `arn` - The Amazon Resource Name (ARN) of the FAQ.
* `created_at` - The Unix datetime that the faq was created.
* `description` - The description of the FAQ.
* `error_message` - When the `status` field value is `FAILED`, this contains a message that explains why.
* `file_format` - The file format used by the input files for the FAQ. Valid Values are `CSV`, `CSV_WITH_HEADER`, `JSON`.
* `id` - The unique identifiers of the FAQ and index separated by a slash (`/`).
* `language_code` - The code for a language. This shows a supported language for the FAQ document. For more information on supported languages, including their codes, see [Adding documents in languages other than English](https://docs.aws.amazon.com/kendra/latest/dg/in-adding-languages.html).
* `name` - Specifies the name of the FAQ.
* `role_arn` - The Amazon Resource Name (ARN) of a role with permission to access the S3 bucket that contains the FAQs. For more information, see [IAM Roles for Amazon Kendra](https://docs.aws.amazon.com/kendra/latest/dg/iam-roles.html).
* `s3_path` - The S3 location of the FAQ input data. Detailed below.
* `status` - The status of the FAQ. It is ready to use when the status is ACTIVE.
* `updated_at` - The date and time that the FAQ was last updated.
* `tags` - Metadata that helps organize the FAQs you create.

The `s3_path` configuration block supports the following attributes:

* `bucket` - The name of the S3 bucket that contains the file.
* `key` - The name of the file.
