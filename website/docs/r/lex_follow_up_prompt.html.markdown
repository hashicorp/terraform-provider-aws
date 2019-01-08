---
layout: "aws"
page_title: "AWS: aws_lex_follow_up_prompt"
sidebar_current: "docs-aws-resource-lex-follow-up-prompt"
description: |-
  Definition of an Amazon Lex Follow Up Prompt used as an attribute in other Lex resources.
---

# aws_lex_follow_up_prompt

A prompt for additional activity after an intent is fulfilled. For example, after the OrderPizza
intent is fulfilled, you might prompt the user to find out whether the user wants to order drinks.

## Example Usage

```hcl
resource "aws_lex_intent" "order_flowers" {
  follow_up_prompt {
    prompt {
      max_attempts = 2

      message {
        content      = "Would you like to place another order?"
        content_type = "PlainText"
      }
    }

    rejection_statement {
      message {
        content      = "Your order for {FlowerType} has been placed and will be ready by {PickupTime} on {PickupDate}"
        content_type = "PlainText"
      }
    }
  }
}
```

## Argument Reference

The following arguments are supported:

### Required

* `prompt`

    * Type: [Lex Prompt](/docs/providers/aws/r/lex_prompt.html)

### Optional

* `rejectionStatement`

	If the user answers "no" to the question defined in the prompt field, Amazon Lex responds with
	this statement to acknowledge that the intent was canceled.

    * Type: [Lex Statement](/docs/providers/aws/r/lex_statement.html)
