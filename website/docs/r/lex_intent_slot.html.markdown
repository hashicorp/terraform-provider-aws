---
layout: "aws"
page_title: "AWS: aws_lex_intent_slot"
sidebar_current: "docs-aws-resource-lex-intent-slot"
description: |-
  Definition of an Amazon Lex Intent Slot used as an attribute in other Lex resources.
---

# aws_lex_intent_slot

Identifies the version of a specific slot.

## Example Usage

```hcl
resource "aws_lex_intent" "order_flowers" {
  slot {
    description = "The type of flowers to pick up"
    name        = "FlowerType"
    priority    = 1

    sample_utterances = [
      "I would like to order {FlowerType}",
    ]

    slot_constraint   = "Required"
    slot_type         = "FlowerTypes"
    slot_type_version = "$LATEST"

    value_elicitation_prompt {
      max_attempts = 2

      message {
        content      = "What type of flowers would you like to order?"
        content_type = "PlainText"
      }
    }
  }
}
```

## Argument Reference

The following arguments are supported:

### Required

* `name`

	The name of the intent slot that you want to create. The name is case sensitive.

    * Type: string
    * Min: 2
    * Max: 50
    * Pattern: ^([A-Za-z]_?)+$

* `slot_constraint`

	Specifies whether the slot is required or optional.

    * Type: string
    * Values: Required | Optional

### Optional

* `description`

	A description of the bot.

    * Type: string
    * Min: 0
    * Max: 200

* `priority`

	Directs Lex the order in which to elicit this slot value from the user. For example, if the
	intent has two slots with priorities 1 and 2, AWS Lex first elicits a value for the slot
	with priority 1.

	If multiple slots share the same priority, the order in which Lex elicits values is arbitrary.

    * Type: number
    * Min: 0
    * Max: 100

* `response_card`

    The response card. Amazon Lex will substitute session attributes and slot values into the
    response card. For more information, see
	[Example: Using a Response Card](https://docs.aws.amazon.com/lex/latest/dg/ex-resp-card.html).

    * Type: string
    * Min: 1
    * Max: 50000

* `sample_utterances`

    If you know a specific pattern with which users might respond to an Amazon Lex request
    for a slot value, you can provide those utterances to improve accuracy. This is optional.
    In most cases, Amazon Lex is capable of understanding user utterances.

    * Type: List<string>
    * Min: 0
    * Max: 10
    * Min Length: 1
    * Max Length: 200

* `slot_type`

    The type of the slot, either a custom slot type that you defined or one of the built-in slot types.

    * Type: string
    * Min: 1
    * Max: 100
    * Regex: ^((AMAZON\.)_?|[A-Za-z]_?)+

* `slot_type_version`

    The version of the slot type.

    * Type: string
    * Min: 1
    * Max: 64
    * Regex: \$LATEST|[0-9]+

* `value_elicitation_prompt`

    The prompt that Amazon Lex uses to elicit the slot value from the user.

    * Type: [Lex Prompt](/docs/providers/aws/r/lex_prompt.html)
