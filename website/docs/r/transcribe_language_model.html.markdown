---
subcategory: "Transcribe"
layout: "aws"
page_title: "AWS: aws_transcribe_language_model"
description: |-
  Terraform resource for managing an AWS Transcribe LanguageModel.
---

# Resource: aws_transcribe_language_model

Terraform resource for managing an AWS Transcribe LanguageModel.

## Example Usage

### Basic Usage

```terraform
data "aws_iam_policy_document" "example" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["transcribe.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "example" {
  name               = "example"
  assume_role_policy = data.aws_iam_policy_document.example.json
}

resource "aws_iam_role_policy" "test_policy" {
  name = "example"
  role = aws_iam_role.example.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = [
          "s3:GetObject",
          "s3:ListBucket",
        ]
        Effect   = "Allow"
        Resource = ["*"]
      },
    ]
  })
}

resource "aws_s3_bucket" "example" {
  bucket        = "example-transcribe"
  force_destroy = true
}

resource "aws_transcribe_language_model" "example" {
  model_name      = "example"
  base_model_name = "NarrowBand"

  input_data_config {
    data_access_role_arn = aws_iam_role.example.arn
    s3_uri               = "s3://${aws_s3_bucket.example.id}/transcribe/"
  }

  language_code = "en-US"

  tags = {
    ENVIRONMENT = "development"
  }
}
```

## Argument Reference

The following arguments are required:

* `example_arg` - (Required) Concise argument description.

The following arguments are optional:

* `optional_arg` - (Optional) Concise argument description.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - ARN of the LanguageModel.
* `example_attribute` - Concise description.

## Timeouts

`aws_transcribe_language_model` provides the following [Timeouts](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts) configuration options:

* `create` - (Optional, Default: `90m`)
* `delete` - (Optional, Default: `90m`)

## Import

Transcribe LanguageModel can be imported using the `model_name`, e.g.,

```
$ terraform import aws_transcribe_language_model.example rft-8012925589
```
