---
subcategory: "Lex V2 Models"
layout: "aws"
page_title: "AWS: aws_lexv2models_bot_alias"
description: |-
  Manages an AWS Lex V2 Models Bot Alias.
---

# Resource: aws_lexv2models_bot_alias

Manages an AWS Lex V2 Models Bot Alias. An alias points to a specific version of a bot so client applications can use a stable name to invoke that version.

## Example Usage

### Basic Usage

```terraform
resource "aws_lexv2models_bot_alias" "example" {
  bot_id         = aws_lexv2models_bot.example.id
  bot_alias_name = "production"
}
```

### Lambda Code Hook and CloudWatch Conversation Logs

```terraform
resource "aws_cloudwatch_log_group" "example" {
  name = "lex/example"
}

resource "aws_lexv2models_bot_alias" "example" {
  bot_id         = aws_lexv2models_bot.example.id
  bot_alias_name = "production"
  bot_version    = aws_lexv2models_bot_version.example.bot_version
  description    = "Production alias for the example bot"

  bot_alias_locale_settings {
    locale_id = "en_US"
    enabled   = true

    code_hook_specification {
      lambda_code_hook {
        code_hook_interface_version = "1.0"
        lambda_arn                  = aws_lambda_function.example.arn
      }
    }
  }

  conversation_log_settings {
    text_log_settings {
      enabled = true

      destination {
        cloudwatch {
          cloudwatch_log_group_arn = aws_cloudwatch_log_group.example.arn
          log_prefix               = "lex/"
        }
      }
    }
  }

  sentiment_analysis_settings {
    detect_sentiment = true
  }
}
```

## Argument Reference

The following arguments are required:

* `bot_alias_name` - (Required) Name of the bot alias. Must be unique for the bot.
* `bot_id` - (Required) Identifier of the bot that the alias applies to. Forces replacement on change.

The following arguments are optional:

* `bot_alias_locale_settings` - (Optional) Per-locale settings that override the bot's locale defaults. See [`bot_alias_locale_settings` Block](#bot_alias_locale_settings-block) for details.
* `bot_version` - (Optional) Version of the bot that this alias points to. When omitted, Lex creates the alias without a bot version and the alias must be updated before it can be used to converse with the bot.
* `conversation_log_settings` - (Optional) Conversation logging configuration. See [`conversation_log_settings` Block](#conversation_log_settings-block) for details.
* `description` - (Optional) Description of the alias.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `sentiment_analysis_settings` - (Optional) Sentiment analysis configuration. See [`sentiment_analysis_settings` Block](#sentiment_analysis_settings-block) for details.
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `bot_alias_locale_settings` Block

The `bot_alias_locale_settings` configuration block supports the following arguments:

* `code_hook_specification` - (Optional) Lambda code hook to invoke for this locale. See [`code_hook_specification` Block](#code_hook_specification-block) for details.
* `enabled` - (Required) Whether to enable the locale for this alias.
* `locale_id` - (Required) String used to identify the locale (for example, `en_US`).

### `code_hook_specification` Block

The `code_hook_specification` configuration block supports the following arguments:

* `lambda_code_hook` - (Required) Lambda function configuration. See [`lambda_code_hook` Block](#lambda_code_hook-block) for details.

### `lambda_code_hook` Block

The `lambda_code_hook` configuration block supports the following arguments:

* `code_hook_interface_version` - (Required) Version of the request-response interface that Lex uses to invoke the Lambda function.
* `lambda_arn` - (Required) ARN of the Lambda function.

### `conversation_log_settings` Block

The `conversation_log_settings` configuration block supports the following arguments. At least one of `audio_log_settings` or `text_log_settings` must be configured for logging to be active.

* `audio_log_settings` - (Optional) One or more audio log destinations. See [`audio_log_settings` Block](#audio_log_settings-block) for details.
* `text_log_settings` - (Optional) One or more text log destinations. See [`text_log_settings` Block](#text_log_settings-block) for details.

### `audio_log_settings` Block

The `audio_log_settings` configuration block supports the following arguments:

* `destination` - (Required) S3 destination for audio logs. See [`audio_log_settings` `destination` Block](#audio_log_settings-destination-block) for details.
* `enabled` - (Required) Whether to enable audio logging.

### `audio_log_settings` `destination` Block

The `destination` configuration block supports the following arguments:

* `s3_bucket` - (Required) S3 bucket configuration for audio logs. See [`s3_bucket` Block](#s3_bucket-block) for details.

### `s3_bucket` Block

The `s3_bucket` configuration block supports the following arguments:

* `kms_key_arn` - (Optional) ARN of a KMS key used to encrypt the audio log files.
* `log_prefix` - (Required) S3 key prefix to apply to audio log files.
* `s3_bucket_arn` - (Required) ARN of the S3 bucket where audio logs are stored.

### `text_log_settings` Block

The `text_log_settings` configuration block supports the following arguments:

* `destination` - (Required) CloudWatch destination for text logs. See [`text_log_settings` `destination` Block](#text_log_settings-destination-block) for details.
* `enabled` - (Required) Whether to enable text logging.

### `text_log_settings` `destination` Block

The `destination` configuration block supports the following arguments:

* `cloudwatch` - (Required) CloudWatch Logs configuration for text logs. See [`cloudwatch` Block](#cloudwatch-block) for details.

### `cloudwatch` Block

The `cloudwatch` configuration block supports the following arguments:

* `cloudwatch_log_group_arn` - (Required) ARN of the CloudWatch Logs log group that receives text logs.
* `log_prefix` - (Required) Prefix applied to the log stream name within the log group.

### `sentiment_analysis_settings` Block

The `sentiment_analysis_settings` configuration block supports the following arguments:

* `detect_sentiment` - (Required) Whether to use Amazon Comprehend to detect the sentiment of user utterances.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the bot alias.
* `bot_alias_id` - Unique identifier of the bot alias.
* `id` - Comma-delimited string concatenating `bot_id` and `bot_alias_id`.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Lex V2 Models Bot Alias using the `id`. For example:

```terraform
import {
  to = aws_lexv2models_bot_alias.example
  id = "ABCDEF1234,GHIJKL5678"
}
```

Using `terraform import`, import Lex V2 Models Bot Alias using the `id`. For example:

```console
% terraform import aws_lexv2models_bot_alias.example ABCDEF1234,GHIJKL5678
```
