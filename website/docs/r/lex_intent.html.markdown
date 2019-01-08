---
layout: "aws"
page_title: "AWS: aws_lex_intent"
sidebar_current: "docs-aws-resource-lex-intent"
description: |-
  Provides an Amazon Lex intent resource.
---

# aws_lex_intent

Provides an Amazon Lex Intent resource. For more information see
[Amazon Lex: How It Works](https://docs.aws.amazon.com/lex/latest/dg/how-it-works.html)

## Example Usage

```hcl
resource "aws_lex_intent" "order_flowers_intent" {
  confirmation_prompt {
    max_attempts = 2

    message {
      content      = "Okay, your {FlowerType} will be ready for pickup by {PickupTime} on {PickupDate}.  Does this sound okay?"
      content_type = "PlainText"
    }
  }

  description = "Intent to order a bouquet of flowers for pick up"

  fulfillment_activity {
    type = "ReturnIntent"
  }

  rejection_statement {
    message {
      content      = "Okay, I will not place your order."
      content_type = "PlainText"
    }
  }

  sample_utterances = [
    "I would like to order some flowers",
    "I would like to pick up flowers",
  ]

  slot {
    description = "The type of flowers to pick up"
    name        = "FlowerType"
    priority    = 1

    sample_utterances = [
      "I would like to order {FlowerType}",
    ]

    slot_constraint   = "Required"
    slot_type         = "FlowerTypes"
    slot_type_version = "$$LATEST"

    value_elicitation_prompt {
      max_attempts = 2

      message {
        content      = "What type of flowers would you like to order?"
        content_type = "PlainText"
      }
    }
  }

  slot {
    description = "The date to pick up the flowers"
    name        = "PickupDate"
    priority    = 2

    sample_utterances = [
      "I would like to order {FlowerType}",
    ]

    slot_constraint   = "Required"
    slot_type         = "AMAZON.DATE"
    slot_type_version = "$$LATEST"

    value_elicitation_prompt {
      max_attempts = 2

      message {
        content      = "What day do you want the {FlowerType} to be picked up?"
        content_type = "PlainText"
      }
    }
  }

  slot {
    description = "The time to pick up the flowers"
    name        = "PickupTime"
    priority    = 3

    sample_utterances = [
      "I would like to order {FlowerType}",
    ]

    slot_constraint   = "Required"
    slot_type         = "AMAZON.TIME"
    slot_type_version = "$$LATEST"

    value_elicitation_prompt {
      max_attempts = 2

      message {
        content      = "Pick up the {FlowerType} at what time on {PickupDate}?"
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

  The name of the intent. The name is not case sensitive.

  * Type: string
  * Min: 1
  * Max: 100
  * Pattern: ^([A-Za-z]_?)+$

### Optional

* `conclusion_statement`

  The statement that you want Amazon Lex to convey to the user after the intent is successfully
  fulfilled by the Lambda function.

  This element is relevant only if you provide a Lambda function in the `fulfillment_activity`. If
  you return the intent to the client application, you can't specify this element.

  The `follow_up_prompt` and `conclusion_statement` are mutually exclusive. You can specify only one.

  * Type: [Lex Statement](/docs/providers/aws/r/lex_statement.html)

* `confirmation_prompt`

  Prompts the user to confirm the intent. This question should have a yes or no answer.

  You you must provide both the `rejection_statement` and `confirmation_prompt`, or neither.

  * Type: [Lex Prompt](/docs/providers/aws/r/lex_statement.html)

* `description`

  A description of the intent.

  * Type: string
  * Min: 0
  * Max: 200

* `dialog_code_hook`

  Specifies a Lambda function to invoke for each user input. You can invoke this Lambda function to
  personalize user interaction.

  * Type: [Lex CodeHook](/docs/providers/aws/r/lex_code_hook.html)

* `follow_up_prompt`

  Amazon Lex uses this prompt to solicit additional activity after fulfilling an intent. For example,
  after the OrderPizza intent is fulfilled, you might prompt the user to order a drink.

  The `follow_up_prompt` field and the `conclusion_statement` field are mutually exclusive. You can
  specify only one.

  * Type: [Lex FollowUpPrompt](/docs/providers/aws/r/lex_follow_up_prompt.html)

* `fulfillment_activity`

  Describes how the intent is fulfilled. For example, after a user provides all of the information
  for a pizza order, `fulfillment_activity` defines how the bot places an order with a local pizza store.

  * Type: [Lex FulfillmentActivity](/docs/providers/aws/r/lex_fulfillment_activity.html)

* `parent_intent_signature`

  A unique identifier for the built-in intent to base this intent on. To find the signature for an
  intent, see [Standard Built-in Intents](https://developer.amazon.com/public/solutions/alexa/alexa-skills-kit/docs/built-in-intent-ref/standard-intents)
  in the Alexa Skills Kit.

  * Type: string

* `rejection_statement`

  When the user answers "no" to the question defined in `confirmation_prompt`, Amazon Lex responds
  with this statement to acknowledge that the intent was canceled.

  You must provide both the `rejection_statement` and the `confirmation_prompt`, or neither.

  * Type: [Lex Statement](/docs/providers/aws/r/lex_statement.html)

* `sample_utterances`

  An array of utterances (strings) that a user might say to signal the intent. For example, "I want
  {PizzaSize} pizza", "Order {Quantity} {PizzaSize} pizzas".

  In each utterance, a slot name is enclosed in curly braces.

  * Type: List of strings
  * Min: 0
  * Max: 1500
  * Min Length: 1
  * Max Length: 200

* `slot`

  An array of intent slots. At runtime, Amazon Lex elicits required slot values from the user using
  prompts defined in the slots.

  * Type: List of [Lex Intent Slots](/docs/providers/aws/r/lex_intent_slot.html)
  * Min: 0
  * Max: 100

## Attributes Reference

The following attributes are exported in addition to the arguments listed above:

* `checksum`

	Checksum identifying the version of the intent that was created. The checksum is not included as
	an argument because the resource will add it automatically when updating the intent.

* `created_date`

	The date when the intent version was created.

* `last_updated_date`

	The date when the $LATEST version of this intent was updated.

* `version`

	The version of the bot.

## Import

Intents can be imported using their name.

```
$ terraform import aws_lex_intent.order_flowers_intent OrderFlowers
```
