---
layout: "aws"
page_title: "AWS: aws_lex_message"
sidebar_current: "docs-aws-resource-lex-message"
description: |-
  Definition of an Amazon Lex Message used as an attribute in other Lex resources.
---

# aws_lex_message

The message object that provides the message text and its type.

## Example Usage

```hcl
resource "aws_lex_bot" "order_flowers_bot" {
  abort_statement {
    message {
      content      = "Sorry, I am not able to assist at this time"
      content_type = "PlainText"
      group_number = 1
    }
  }
}
```

## Argument Reference

The following arguments are supported:

### Required

* `content`

	The text of the message.

    * Type: string
    * Min: 1
    * Max: 1000

* `content_type`

	The content type of the message string.

    * Type: string
    * Values: PlainText | SSML | CustomPayload

### Optional

* `group_number`

    Identifies the message group that the message belongs to. When a group is assigned to a message,
    Amazon Lex returns one message from each group in the response.

    * Type: number
    * Min: 1
    * Max: 5
