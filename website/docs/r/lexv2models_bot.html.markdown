---
subcategory: "Lex V2 Models"
layout: "aws"
page_title: "AWS: aws_lexv2models_bot"
description: |-
  Terraform resource for managing an AWS Lex V2 Models Bot.
---

# Resource: aws_lexv2models_bot

Terraform resource for managing an AWS Lex V2 Models Bot.

## Example Usage

### Basic Usage

```terraform
resource "aws_lexv2models_bot" "example" {
  name        = "example"
  description = "Example description"
  data_privacy {
    child_directed = false
  }
  idle_session_ttl_in_seconds = 60
  role_arn                    = aws_iam_role.example.arn
  type                        = "Bot"

  tags = {
    foo = "bar"
  }
}

resource "aws_iam_role" "example" {
  name = "example"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Sid    = ""
        Principal = {
          Service = "lexv2.amazonaws.com"
        }
      },
    ]
  })

  tags = {
    created_by = "aws"
  }
}
```

## Argument Reference

The following arguments are required:

* `name` - Name of the bot. The bot name must be unique in the account that creates the bot. Type String. Length Constraints: Minimum length of 1. Maximum length of 100.
* `data_privacy` - Provides information on additional privacy protections Amazon Lex should use with the bot's data. See [`data_privacy`](#data-privacy)
* `idle_session_ttl_in_seconds` - Time, in seconds, that Amazon Lex should keep information about a user's conversation with the bot. You can specify between 60 (1 minute) and 86,400 (24 hours) seconds.
* `role_arn` - ARN of an IAM role that has permission to access the bot.

The following arguments are optional:

* `members` - List of bot members in a network to be created. See [`bot_members`](#bot-members).
* `tags` - List of tags to add to the bot. You can only add tags when you create a bot.
* `type` - Type of a bot to create. Possible values are `"Bot"` and `"BotNetwork"`.
* `description` - Description of the bot. It appears in lists to help you identify a particular bot.
* `test_bot_alias_tags` - List of tags to add to the test alias for a bot. You can only add tags when you create a bot.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Unique identifier for a particular bot.

### Data Privacy

* `child_directed` (Required) -  For each Amazon Lex bot created with the Amazon Lex Model Building Service, you must specify whether your use of Amazon Lex is related to a website, program, or other application that is directed or targeted, in whole or in part, to children under age 13 and subject to the Children's Online Privacy Protection Act (COPPA) by specifying true or false in the childDirected field.

### Bot Members

* `alias_id` (Required) - Alias ID of a bot that is a member of this network of bots.
* `alias_name` (Required) - Alias name of a bot that is a member of this network of bots.
* `id` (Required) - Unique ID of a bot that is a member of this network of bots.
* `name` (Required) - Unique name of a bot that is a member of this network of bots.
* `version` (Required) - Version of a bot that is a member of this network of bots.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Lex V2 Models Bot using the `id`. For example:

```terraform
import {
  to = aws_lexv2models_bot.example
  id = "bot-id-12345678"
}
```

Using `terraform import`, import Lex V2 Models Bot using the `id`. For example:

```console
% terraform import aws_lexv2models_bot.example bot-id-12345678
```
