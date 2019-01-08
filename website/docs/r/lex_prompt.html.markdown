---
layout: "aws"
page_title: "AWS: aws_lex_prompt"
sidebar_current: "docs-aws-resource-lex-prompt"
description: |-
  Definition of an Amazon Lex Prompt used as an attribute in other Lex resources.
---

# aws_lex_prompt

Obtains information from the user. To define a prompt, provide one or more messages and specify the
number of attempts to get information from the user. If you provide more than one message, Amazon
Lex chooses one of the messages to use to prompt the user.

## Example Usage

```hcl
resource "aws_lex_bot" "order_flower_bot" {
  clarification_prompt {
    max_attempts = 2

    message {
      content      = "I didn't understand you, what would you like to do?"
      content_type = "PlainText"
    }

    response_card = "..."
  }
}
```

## Argument Reference

The following arguments are supported:

### Required

* `max_attempts`

    The number of times to prompt the user for information.

    * Type: number
    * Min: 1
    * Max: 5

* `message`

	A set of messages, each of which provides a message string and its type. You can specify the
	message string in plain text or in Speech Synthesis Markup Language (SSML).

    * Type: Set of [Lex Messages](/docs/providers/aws/r/lex_message.html)
    * Min: 1
    * Max: 15

### Optional

* `response_card`

    The response card. Amazon Lex will substitute session attributes and slot values into the
    response card. For more information, see
	[Example: Using a Response Card](https://docs.aws.amazon.com/lex/latest/dg/ex-resp-card.html).

    * Type: string
    * Min: 1
    * Max: 50000
