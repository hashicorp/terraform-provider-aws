---
subcategory: "Connect"
layout: "aws"
page_title: "AWS: aws_connect_bot_association"
description: |-
  Associates an Amazon Connect instance to an Amazon Lex (V1) bot
---

# Resource: aws_connect_bot_association

Allows the specified Amazon Connect instance to access the specified Amazon Lex (V1) bot. For more information see
[Amazon Connect: Getting Started](https://docs.aws.amazon.com/connect/latest/adminguide/amazon-connect-get-started.html) and [Add an Amazon Lex bot](https://docs.aws.amazon.com/connect/latest/adminguide/amazon-lex.html).

~> **NOTE:** This resource only currently supports Amazon Lex (V1) Associations.

## Example Usage

### Basic

```terraform
resource "aws_connect_bot_association" "example" {
  instance_id = aws_connect_instance.example.id
  lex_bot {
    lex_region = "us-west-2"
    name       = "Test"

  }
}
```

### Including a sample Lex bot

```terraform
data "aws_region" "current" {}

resource "aws_lex_intent" "example" {
  create_version = true
  name           = "connect_lex_intent"
  fulfillment_activity {
    type = "ReturnIntent"
  }
  sample_utterances = [
    "I would like to pick up flowers.",
  ]
}

resource "aws_lex_bot" "example" {
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
    intent_name    = aws_lex_intent.example.name
    intent_version = "1"
  }

  child_directed   = false
  name             = "connect_lex_bot"
  process_behavior = "BUILD"
}

resource "aws_connect_bot_association" "example" {
  instance_id = aws_connect_instance.example.id
  lex_bot {
    lex_region = data.aws_region.current.name
    name       = aws_lex_bot.example.name
  }
}
```

## Argument Reference

The following arguments are supported:

* `instance_id` - (Required) The identifier of the Amazon Connect instance. You can find the instanceId in the ARN of the instance.
* `lex_bot` - (Required) Configuration information of an Amazon Lex (V1) bot. Detailed below.

### lex_bot

The `lex_bot` configuration block supports the following:

* `name` - (Required) The name of the Amazon Lex (V1) bot.
* `lex_region` - (Optional) The Region that the Amazon Lex (V1) bot was created in. Defaults to current region.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The Amazon Connect instance ID, Lex (V1) bot name, and Lex (V1) bot region separated by colons (`:`).

## Import

`aws_connect_bot_association` can be imported by using the Amazon Connect instance ID, Lex (V1) bot name, and Lex (V1) bot region separated by colons (`:`), e.g.

```
$ terraform import aws_connect_bot_association.example aaaaaaaa-bbbb-cccc-dddd-111111111111:Example:us-west-2
```
