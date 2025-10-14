---
subcategory: "Kendra"
layout: "aws"
page_title: "AWS: aws_kendra_faq"
description: |-
  Terraform resource for managing an AWS Kendra FAQ.
---

# Resource: aws_kendra_faq

Terraform resource for managing an AWS Kendra FAQ.

## Example Usage

### Basic

```terraform
resource "aws_kendra_faq" "example" {
  index_id = aws_kendra_index.example.id
  name     = "Example"
  role_arn = aws_iam_role.example.arn

  s3_path {
    bucket = aws_s3_bucket.example.id
    key    = aws_s3_object.example.key
  }

  tags = {
    Name = "Example Kendra Faq"
  }
}
```

### With File Format

```terraform
resource "aws_kendra_faq" "example" {
  index_id    = aws_kendra_index.example.id
  name        = "Example"
  file_format = "CSV"
  role_arn    = aws_iam_role.example.arn

  s3_path {
    bucket = aws_s3_bucket.example.id
    key    = aws_s3_object.example.key
  }
}
```

### With Language Code

```terraform
resource "aws_kendra_faq" "example" {
  index_id      = aws_kendra_index.example.id
  name          = "Example"
  language_code = "en"
  role_arn      = aws_iam_role.example.arn

  s3_path {
    bucket = aws_s3_bucket.example.id
    key    = aws_s3_object.example.key
  }
}
```

## Argument Reference

The following arguments are required:

* `index_id`- (Required, Forces new resource) The identifier of the index for a FAQ.
* `name` - (Required, Forces new resource) The name that should be associated with the FAQ.
* `role_arn` - (Required, Forces new resource) The Amazon Resource Name (ARN) of a role with permission to access the S3 bucket that contains the FAQs. For more information, see [IAM Roles for Amazon Kendra](https://docs.aws.amazon.com/kendra/latest/dg/iam-roles.html).
* `s3_path` - (Required, Forces new resource) The S3 location of the FAQ input data. Detailed below.

The `s3_path` configuration block supports the following arguments:

* `bucket` - (Required, Forces new resource) The name of the S3 bucket that contains the file.
* `key` - (Required, Forces new resource) The name of the file.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `description` - (Optional, Forces new resource) The description for a FAQ.
* `file_format` - (Optional, Forces new resource) The file format used by the input files for the FAQ. Valid Values are `CSV`, `CSV_WITH_HEADER`, `JSON`.
* `language_code` - (Optional, Forces new resource) The code for a language. This shows a supported language for the FAQ document. English is supported by default. For more information on supported languages, including their codes, see [Adding documents in languages other than English](https://docs.aws.amazon.com/kendra/latest/dg/in-adding-languages.html).
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the FAQ.
* `created_at` - The Unix datetime that the FAQ was created.
* `error_message` - When the Status field value is `FAILED`, this contains a message that explains why.
* `faq_id` - The identifier of the FAQ.
* `id` - The unique identifiers of the FAQ and index separated by a slash (`/`)
* `status` - The status of the FAQ. It is ready to use when the status is ACTIVE.
* `updated_at` - The date and time that the FAQ was last updated.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_kendra_faq` using the unique identifiers of the FAQ and index separated by a slash (`/`). For example:

```terraform
import {
  to = aws_kendra_faq.example
  id = "faq-123456780/idx-8012925589"
}
```

Using `terraform import`, import `aws_kendra_faq` using the unique identifiers of the FAQ and index separated by a slash (`/`). For example:

```console
% terraform import aws_kendra_faq.example faq-123456780/idx-8012925589
```
