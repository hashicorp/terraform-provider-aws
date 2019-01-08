---
layout: "aws"
page_title: "AWS: aws_lex_statement"
sidebar_current: "docs-aws-resource-lex-statement"
description: |-
  Definition of an Amazon Lex Statement used as an attribute in other Lex resources.
---

# aws_lex_statement

A statement is a map with a set of message maps and an optional response card string. Messages
convey information to the user. At runtime, Amazon Lex selects the message to convey.

## Example Usage

```hcl
resource "aws_lex_bot" "florist_bot" {
  abort_statement {
    message {
      content      = "Sorry, I am not able to assist at this time"
      content_type = "PlainText"
    }

    message {
      content      = "Sorry, I do not understand"
      content_type = "PlainText"
    }

    response_card = "..."
  }
}
```

## Argument Reference

The following arguments are supported:

### Required

* `message`

	A set of messages, each of which provides a message string and its type. You can specify the
	message string in plain text or in Speech Synthesis Markup Language (SSML).

    * Type: Set of [Lex Messages](/docs/providers/aws/r/lex_messages.html)
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
