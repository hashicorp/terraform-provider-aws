---
subcategory: "IVS (Interactive Video) Chat"
layout: "aws"
page_title: "AWS: aws_ivschat_room"
description: |-
  Terraform resource for managing an AWS IVS (Interactive Video) Chat Room.
---

# Resource: aws_ivschat_room

Terraform resource for managing an AWS IVS (Interactive Video) Chat Room.

## Example Usage

### Basic Usage

```terraform
resource "aws_ivschat_room" "example" {
  name = "tf-room"
}
```

### Usage with Logging Configuration to S3 Bucket

```terraform
resource "aws_s3_bucket" "example" {
  bucket_prefix = "tf-ivschat-logging-bucket-"
  force_destroy = true
}

resource "aws_ivschat_logging_configuration" "example" {
  name = "tf-ivschat-loggingconfiguration"

  lifecycle {
    create_before_destroy = true
  }

  destination_configuration {
    s3 {
      bucket_name = aws_s3_bucket.example.id
    }
  }
}

resource "aws_ivschat_room" "example" {
  name                              = "tf-ivschat-room"
  logging_configuration_identifiers = [aws_ivschat_logging_configuration.example.arn]
}
```

### Usage with Message Review Handler

The message review handler is a Lambda function that receives each message's `Content` and returns a `ReviewResult` (`ALLOW` or `DENY`) and, optionally, modified `Content`. IVS Chat must be granted permission to invoke the function before the room is created.

```terraform
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}
data "aws_region" "current" {}

data "aws_iam_policy_document" "assume_role" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "example" {
  name               = "example-ivschat-message-review-handler"
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

data "archive_file" "example" {
  type                    = "zip"
  source_content_filename = "index.js"
  source_content          = <<-EOT
    exports.handler = async function ({ Content }) {
      return { ReviewResult: "ALLOW", Content: Content + " - edited by Lambda" };
    };
  EOT
  output_path             = "${path.module}/message_review_handler.zip"
}

resource "aws_lambda_function" "example" {
  filename      = data.archive_file.example.output_path
  function_name = "example-ivschat-message-review-handler"
  role          = aws_iam_role.example.arn
  handler       = "index.handler"
  runtime       = "nodejs24.x"
  code_sha256   = data.archive_file.example.output_base64sha256
}

resource "aws_lambda_permission" "example" {
  action         = "lambda:InvokeFunction"
  function_name  = aws_lambda_function.example.function_name
  principal      = "ivschat.amazonaws.com"
  source_account = data.aws_caller_identity.current.account_id
  source_arn     = "arn:${data.aws_partition.current.partition}:ivschat:${data.aws_region.current.region}:${data.aws_caller_identity.current.account_id}:room/*"
}

resource "aws_ivschat_room" "example" {
  name = "example"

  message_review_handler {
    uri             = aws_lambda_function.example.arn
    fallback_result = "ALLOW"
  }

  depends_on = [aws_lambda_permission.example]
}
```

## Argument Reference

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `logging_configuration_identifiers` - (Optional) List of Logging Configuration
  ARNs to attach to the room.
* `maximum_message_length` - (Optional) Maximum number of characters in a single
  message. Messages are expected to be UTF-8 encoded and this limit applies
  specifically to rune/code-point count, not number of bytes.
* `maximum_message_rate_per_second` - (Optional) Maximum number of messages per
  second that can be sent to the room (by all clients).
* `message_review_handler` - (Optional) Configuration information for optional
  review of messages.
    * `fallback_result` - (Optional) The fallback behavior (whether the message
    is allowed or denied) if the handler does not return a valid response,
    encounters an error, or times out. Valid values: `ALLOW`, `DENY`.
    * `uri` - (Optional) ARN of the lambda message review handler function.
* `name` - (Optional) Room name.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Room.
* `id` - Room ID
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts):

* `create` - (Default `5m`)
* `update` - (Default `5m`)
* `delete` - (Default `5m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_ivschat_room.example
  identity = {
    "arn" = "arn:aws:ivschat:us-west-2:123456789012:room/g1H2I3j4k5L6"
  }
}

resource "aws_ivschat_room" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

- `arn` (String) ARN of the IVS Chat room.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import IVS (Interactive Video) Chat Room using the ARN. For example:

```terraform
import {
  to = aws_ivschat_room.example
  id = "arn:aws:ivschat:us-west-2:326937407773:room/GoXEXyB4VwHb"
}
```

Using `terraform import`, import IVS (Interactive Video) Chat Room using the ARN. For example:

```console
% terraform import aws_ivschat_room.example arn:aws:ivschat:us-west-2:326937407773:room/GoXEXyB4VwHb
```
