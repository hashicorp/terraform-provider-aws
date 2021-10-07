---
subcategory: "Connect"
layout: "aws"
page_title: "AWS: aws_connect_bot_association"
description: |-
  Provides details about a specific Connect Bot Association.
---

# Resource: aws_connect_bot_association

Allows the specified Amazon Connect instance to access the specified Amazon V1 Lex bot. For more information see
[Amazon Connect: Getting Started](https://docs.aws.amazon.com/connect/latest/adminguide/amazon-connect-get-started.html)

[Add an Amazon Lex bot](https://docs.aws.amazon.com/connect/latest/adminguide/amazon-lex.html)

~> **NOTE:** This resource only currently supports Amazon Lex (Classic) Associations.

## Example Usage
### Basic

```hcl
resource "aws_connect_bot_association" "test" {
  bot_name    = "Test"
  instance_id = aws_connect_instance.test.id
  lex_region  = "us-west-2"
}
```

### Including a sample Lex bot

```hcl
data "aws_region" "current" {}

resource "aws_lex_intent" "test" {
  create_version = true
  name           = "connect_lex_intent"
  fulfillment_activity {
    type = "ReturnIntent"
  }
  sample_utterances = [
    "I would like to pick up flowers.",
  ]
}

resource "aws_lex_bot" "test" {
  abort_statement {
    message {
      content      = "Sorry, I am not able to assist at this time."
      content_type = "PlainText"
    }
  }
  clarification_prompt {
    max_attempts = 2
    message {
      content      = "I didn't understand you, what would you like to do?"
      content_type = "PlainText"
    }
  }
  intent {
    intent_name    = aws_lex_intent.test.name
    intent_version = "1"
  }

  child_directed   = false
  name             = "connect_lex_bot"
  process_behavior = "BUILD"
}

resource "aws_connect_bot_association" "test" {
  bot_name    = "connect_lex_bot"
  instance_id = aws_connect_instance.test.id
  lex_region  = data.aws_region.current.name
}
```

## Argument Reference

The following arguments are supported:

* `bot_name` - (Required) The name of the Amazon V1 Lex bot.
* `instance_id` - (Required) The identifier of the Amazon Connect instance. You can find the instanceId in the ARN of the instance.
* `lex_region` - (Required) The Region in which the Amazon V1 Lex bot has been created.

### Timeouts

The `timeouts` block allows you to specify [timeouts](https://www.terraform.io/docs/configuration/resources.html#timeouts) for certain actions:

* `create` - (Defaults to 5 mins) Used when creating the association.
* `delete` - (Defaults to 5 mins) Used when creating the association.

## Attributes Reference

No additional attributes are exported.
