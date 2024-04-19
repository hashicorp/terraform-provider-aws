---
subcategory: "Transcribe"
layout: "aws"
page_title: "AWS: aws_transcribe_language_model"
description: |-
  Terraform resource for managing an AWS Transcribe LanguageModel.
---

# Resource: aws_transcribe_language_model

Terraform resource for managing an AWS Transcribe LanguageModel.

-> This resource can take a significant amount of time to provision. See Language Model [FAQ](https://aws.amazon.com/transcribe/faqs/) for more details.

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

resource "aws_s3_object" "object" {
  bucket = aws_s3_bucket.example.id
  key    = "transcribe/test1.txt"
  source = "test1.txt"
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

* `base_model_name` - (Required) Name of reference base model.
* `input_data_config` - (Required) The input data config for the LanguageModel. See [Input Data Config](#input-data-config) for more details.
* `language_code` - (Required) The language code you selected for your language model. Refer to the [supported languages](https://docs.aws.amazon.com/transcribe/latest/dg/supported-languages.html) page for accepted codes.
* `model_name` - (Required) The model name.

### Input Data Config

* `data_access_role_arn` - (Required) IAM role with access to S3 bucket.
* `s3_uri` - (Required) S3 URI where training data is located.
* `tuning_data_s3_uri` - (Optional) S3 URI where tuning data is located.

The following arguments are optional:

* `tags` - (Optional) A map of tags to assign to the LanguageModel. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - LanguageModel name.
* `arn` - ARN of the LanguageModel.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `600m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Transcribe LanguageModel using the `model_name`. For example:

```terraform
import {
  to = aws_transcribe_language_model.example
  id = "example-name"
}
```

Using `terraform import`, import Transcribe LanguageModel using the `model_name`. For example:

```console
% terraform import aws_transcribe_language_model.example example-name
```
