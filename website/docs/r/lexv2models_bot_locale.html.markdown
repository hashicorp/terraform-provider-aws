---
subcategory: "Lex V2 Models"
layout: "aws"
page_title: "AWS: aws_lexv2models_bot_locale"
description: |-
  Terraform resource for managing an AWS Lex V2 Models Bot Locale.
---

# Resource: aws_lexv2models_bot_locale

Terraform resource for managing an AWS Lex V2 Models Bot Locale.

## Example Usage

### Basic Usage

```terraform
resource "aws_lexv2models_bot_locale" "example" {
  bot_id                           = aws_lexv2models_bot.example.id
  bot_version                      = "DRAFT"
  locale_id                        = "en_US"
  n_lu_intent_confidence_threshold = 0.70
}
```

### Voice Settings

```terraform
resource "aws_lexv2models_bot_locale" "example" {
  bot_id                           = aws_lexv2models_bot.example.id
  bot_version                      = "DRAFT"
  locale_id                        = "en_US"
  n_lu_intent_confidence_threshold = 0.70

  voice_settings {
    voice_id = "Kendra"
    engine   = "standard"
  }
}
```

## Argument Reference

The following arguments are required:

* `bot_id` - Identifier of the bot to create the locale for.
* `bot_version` - Version of the bot to create the locale for. This can only be the draft version of the bot.
* `locale_id` - Identifier of the language and locale that the bot will be used in. The string must match one of the supported locales. All of the intents, slot types, and slots used in the bot must have the same locale. For more information, see Supported languages (https://docs.aws.amazon.com/lexv2/latest/dg/how-languages.html)
* `n_lu_intent_confidence_threshold` - Determines the threshold where Amazon Lex will insert the AMAZON.FallbackIntent, AMAZON.KendraSearchIntent, or both when returning alternative intents.

The following arguments are optional:

* `description` - Description of the bot locale. Use this to help identify the bot locale in lists.
* `voice_settings` - Amazon Polly voice ID that Amazon Lex uses for voice interaction with the user. See [`voice_settings`](#voice-settings).

### Voice Settings

* `voice_id` - (Required) Identifier of the Amazon Polly voice to use.
* `engine` - (Optional) Indicates the type of Amazon Polly voice that Amazon Lex should use for voice interaction with the user. Valid values are `standard` and `neural`. If not specified, the default is `standard`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Comma-delimited string joining `locale_id`, `bot_id`, and `bot_version`.
* `name` - Specified locale name.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Lex V2 Models Bot Locale using the `id`. For example:

```terraform
import {
  to = aws_lexv2models_bot_locale.example
  id = "en_US,abcd-12345678,1"
}
```

Using `terraform import`, import Lex V2 Models Bot Locale using the `id`. For example:

```console
% terraform import aws_lexv2models_bot_locale.example en_US,abcd-12345678,1
```
