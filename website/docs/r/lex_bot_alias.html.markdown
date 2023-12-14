---
subcategory: "Lex Model Building"
layout: "aws"
page_title: "AWS: aws_lex_bot_alias"
description: |-
  Provides an Amazon Lex Bot Alias resource.
---

# Resource: aws_lex_bot_alias

Provides an Amazon Lex Bot Alias resource. For more information see
[Amazon Lex: How It Works](https://docs.aws.amazon.com/lex/latest/dg/how-it-works.html)

## Example Usage

```terraform
resource "aws_lex_bot_alias" "order_flowers_prod" {
  bot_name    = "OrderFlowers"
  bot_version = "1"
  description = "Production Version of the OrderFlowers Bot."
  name        = "OrderFlowersProd"
}
```

## Argument Reference

This resource supports the following arguments:

* `bot_name` - (Required) The name of the bot.
* `bot_version` - (Required) The version of the bot.
* `conversation_logs` - (Optional) The settings that determine how Amazon Lex uses conversation logs for the alias. Attributes are documented under [conversation_logs](#conversation_logs).
* `description` - (Optional) A description of the alias. Must be less than or equal to 200 characters in length.
* `name` - (Required) The name of the alias. The name is not case sensitive. Must be less than or equal to 100 characters in length.

### conversation_logs

Contains information about conversation log settings.

* `iam_role_arn` - (Required) The Amazon Resource Name (ARN) of the IAM role used to write your logs to CloudWatch Logs or an S3 bucket. Must be between 20 and 2048 characters in length.
* `log_settings` - (Optional) The settings for your conversation logs. You can log text, audio, or both. Attributes are documented under [log_settings](#log_settings).

### log_settings

The settings for conversation logs.

* `destination` - (Required) The destination where logs are delivered. Options are `CLOUDWATCH_LOGS` or `S3`.
* `kms_key_arn` - (Optional) The Amazon Resource Name (ARN) of the key used to encrypt audio logs in an S3 bucket. This can only be specified when `destination` is set to `S3`. Must be between 20 and 2048 characters in length.
* `log_type` - (Required) The type of logging that is enabled. Options are `AUDIO` or `TEXT`.
* `resource_arn` - (Required) The Amazon Resource Name (ARN) of the CloudWatch Logs log group or S3 bucket where the logs are delivered. Must be less than or equal to 2048 characters in length.
* `resource_prefix` - (Computed) The prefix of the S3 object key for `AUDIO` logs or the log stream name for `TEXT` logs.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The ARN of the bot alias.
* `checksum` - Checksum of the bot alias.
* `created_date` - The date that the bot alias was created.
* `last_updated_date` - The date that the bot alias was updated. When you create a resource, the creation date and the last updated date are the same.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `1m`)
* `update` - (Default `1m`)
* `delete` - (Default `5m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import bot aliases using an ID with the format `bot_name:bot_alias_name`. For example:

```terraform
import {
  to = aws_lex_bot_alias.order_flowers_prod
  id = "OrderFlowers:OrderFlowersProd"
}
```

Using `terraform import`, import bot aliases using an ID with the format `bot_name:bot_alias_name`. For example:

```console
% terraform import aws_lex_bot_alias.order_flowers_prod OrderFlowers:OrderFlowersProd
```
