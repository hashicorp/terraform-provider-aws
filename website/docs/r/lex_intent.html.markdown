---
layout: "aws"
page_title: "AWS: aws_lex_intent"
sidebar_current: "docs-aws-resource-lex-intent"
description: |-
  Provides an Amazon Lex intent resource.
---

# Resource: aws_lex_intent

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

* `conclusion_statement` _(Optional, Type: map)_:

    The statement that you want Amazon Lex to convey to the user after the intent is successfully
    fulfilled by the Lambda function.

    This element is relevant only if you provide a Lambda function in the `fulfillment_activity`. If
    you return the intent to the client application, you can't specify this element.

    The `follow_up_prompt` and `conclusion_statement` are mutually exclusive. You can specify only one.

    Attributes are documented under [statement](#statement).

* `confirmation_prompt` _(Optional, Type: map)_:

    Prompts the user to confirm the intent. This question should have a yes or no answer. You you 
    must provide both the `rejection_statement` and `confirmation_prompt`, or neither. Attributes 
    are documented under [prompt](#prompt-1).

* `description` _(Optional, Type: string, Min: 0, Max: 200)_:

    A description of the intent.

* `dialog_code_hook` _(Optional, Type: map)_:

    Specifies a Lambda function to invoke for each user input. You can invoke this Lambda function to
    personalize user interaction. Attributes are documented under [code_hook](#code_hook-1).

* `follow_up_prompt` _(Optional, Type: map)_:

    Amazon Lex uses this prompt to solicit additional activity after fulfilling an intent. For 
    example, after the OrderPizza intent is fulfilled, you might prompt the user to order a drink.

    The `follow_up_prompt` field and the `conclusion_statement` field are mutually exclusive. You 
    can specify only one.

    Attributes are documented under [follow_up_prompt](#follow_up_prompt-1).

* `fulfillment_activity` _(Optional, Type: map)_:

    Describes how the intent is fulfilled. For example, after a user provides all of the information
    for a pizza order, `fulfillment_activity` defines how the bot places an order with a local pizza store.

    Attributes are documented under [fulfillment_activity](#fulfillment_activity-1).

* `name` _(**Required**, Type: string, Min: 1, Max: 100, Regex: \^([A-Za-z]\_?)+$)_:

    The name of the intent, not case sensitive.

* `parent_intent_signature` _(Optional, Type: string)_:

    A unique identifier for the built-in intent to base this intent on. To find the signature for an
    intent, see [Standard Built-in Intents](https://developer.amazon.com/public/solutions/alexa/alexa-skills-kit/docs/built-in-intent-ref/standard-intents)
    in the Alexa Skills Kit.

* `rejection_statement` _(Optional, Type: map)_:

    When the user answers "no" to the question defined in `confirmation_prompt`, Amazon Lex responds
    with this statement to acknowledge that the intent was canceled.

    You must provide both the `rejection_statement` and the `confirmation_prompt`, or neither.

    Attributes are documented under [statement](#statement).

* `sample_utterances` _(Optional, Type: list of strings, Min: 0, Max: 1500, Min Length: 1, Max Length: 200)_:

    An array of utterances (strings) that a user might say to signal the intent. For example, "I want
    {PizzaSize} pizza", "Order {Quantity} {PizzaSize} pizzas".

    In each utterance, a slot name is enclosed in curly braces.

* `slot` _(Optional, Type: list, Min: 0, Max 100)_:

    An list of intent slots. At runtime, Amazon Lex elicits required slot values from the user using
    prompts defined in the slots. Attributes are documented under [slot](#slot-1).

### code_hook

Specifies a Lambda function that verifies requests to a bot or fulfills the user's request to a bot.

* `message_version` _(Required, Type: string, Min: 1, Max: 5)_:

    The version of the request-response that you want Amazon Lex to use to invoke your Lambda 
    function. For more information, see 
    [Using Lambda Functions](https://docs.aws.amazon.com/lex/latest/dg/using-lambda.html).

* `uri` _(Required, Type: string, Min: 20, Max: 2048, Regex:_ arn:aws:lambda:[a-z]+-[a-z]+-[0-9]:[0-9]{12}:function:[a-zA-Z0-9-\_]+(/[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})?(:[a-zA-Z0-9-\_]+)?):-

    The Amazon Resource Name (ARN) of the Lambda function.

### follow_up_prompt

A prompt for additional activity after an intent is fulfilled. For example, after the OrderPizza
intent is fulfilled, you might prompt the user to find out whether the user wants to order drinks.

* `prompt` _(Required, Type: map)_:

    Prompts for information from the user. Attributes are documented under [prompt](#prompt-1).

* `rejectionStatement` _(Optional, Type: map)_:

    If the user answers "no" to the question defined in the prompt field, Amazon Lex responds with
    this statement to acknowledge that the intent was canceled. Attributes are documented below 
    under [statement](#statement).

### fulfillment_activity

Describes how the intent is fulfilled after the user provides all of the information required for the intent.

* `type` _(Required, Type: string, Values: ReturnIntent | CodeHook)_:

    How the intent should be fulfilled, either by running a Lambda function or by returning the
    slot data to the client application.

* `code_hook` _(Optional, Type: map)_:

    A description of the Lambda function that is run to fulfill the intent. Required if type is CodeHook. Attributes are documented under [code_hook](#code_hook-1).

### message

The message object that provides the message text and its type.

* `content` _(**Required**, Type: string, Min: 1, Max: 1000)_:

	  The text of the message.

* `content_type` _(**Required**, Type: string, Values: PlainText | SSML | CustomPayload)_:

	  The content type of the message string.

* `group_number` _(Optional, Type: number, Min: 1, Max: 5)_:

    Identifies the message group that the message belongs to. When a group is assigned to a message,
    Amazon Lex returns one message from each group in the response.

### prompt

Obtains information from the user. To define a prompt, provide one or more messages and specify the
number of attempts to get information from the user. If you provide more than one message, Amazon
Lex chooses one of the messages to use to prompt the user.

* `max_attempts` _(**Required**, Type: number, Min: 1, Max: 5)_:

    The number of times to prompt the user for information.

* `message` _(**Required**, Type: Set, Min: 1, Max: 15)_:

    A set of messages, each of which provides a message string and its type. You can specify the
	  message string in plain text or in Speech Synthesis Markup Language (SSML). Attributes are 
    documented under [message](#message-1).

* `response_card` _(Optional, Type: string, Min: 1, Max: 50000)_:

    The response card. Amazon Lex will substitute session attributes and slot values into the
    response card. For more information, see 
    [Example: Using a Response Card](https://docs.aws.amazon.com/lex/latest/dg/ex-resp-card.html).

### slot

Identifies the version of a specific slot.

* `name` _(Required, Type: string, Min: 2, Max: 50, Regex: \^([A-Za-z]\_?)+$)_:

	  The name of the intent slot that you want to create. The name is case sensitive.

* `slot_constraint` _(Required, Type: string, Values: Required | Optional)_:

	  Specifies whether the slot is required or optional.

* `description` _(Optional, Type: string, Min: 0, Max: 200)_:

	  A description of the bot.

* `priority` _(Optional, Type: number, Min: 0, Max: 100)_:

    Directs Lex the order in which to elicit this slot value from the user. For example, if the
    intent has two slots with priorities 1 and 2, AWS Lex first elicits a value for the slot
    with priority 1.

	  If multiple slots share the same priority, the order in which Lex elicits values is arbitrary.

* `response_card` _(Optional, Type: string, Min: 1, Max: 50000)_:

    The response card. Amazon Lex will substitute session attributes and slot values into the
    response card. For more information, see
    [Example: Using a Response Card](https://docs.aws.amazon.com/lex/latest/dg/ex-resp-card.html).

* `sample_utterances` _(Optional, Type: list of strings, Min: 0, Max: 10, Min Length: 1, Max Length: 200)_:

    If you know a specific pattern with which users might respond to an Amazon Lex request
    for a slot value, you can provide those utterances to improve accuracy. This is optional.
    In most cases, Amazon Lex is capable of understanding user utterances.

* `slot_type` _(Optional, Type: string, Min: 1, Max: 100, Regex: \^((AMAZON\\.)\_?|[A-Za-z]\_?)+)_:

    The type of the slot, either a custom slot type that you defined or one of the built-in slot types.

* `slot_type_version` _(Optional, Type: string, Min: 1, Max: 64, Regex: \$LATEST|[0-9]+)_:

    The version of the slot type.

* `value_elicitation_prompt` _(Optional, Type: map)_:

    The prompt that Amazon Lex uses to elicit the slot value from the user. Attributes are 
    documented under [prompt](#prompt-1).

### statement

A statement is a map with a set of message maps and an optional response card string. Messages
convey information to the user. At runtime, Amazon Lex selects the message to convey.

* `message` _(**Required**, Type: Set, Min: 1, Max: 15)_:

    A set of messages, each of which provides a message string and its type. You can specify the
	  message string in plain text or in Speech Synthesis Markup Language (SSML). Attributes are 
    documented under [message](#message-1).

* `response_card` _(Optional, Type: string, Min: 1, Max: 50000)_:

    The response card. Amazon Lex will substitute session attributes and slot values into the
    response card. For more information, see 
    [Example: Using a Response Card](https://docs.aws.amazon.com/lex/latest/dg/ex-resp-card.html).

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
