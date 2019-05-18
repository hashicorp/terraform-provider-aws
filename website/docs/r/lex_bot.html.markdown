---
layout: "aws"
page_title: "AWS: aws_lex_bot"
sidebar_current: "docs-aws-resource-lex-bot"
description: |-
  Provides an Amazon Lex bot resource.
---

# Resource: aws_lex_bot

Provides an Amazon Lex Bot resource. For more information see
[Amazon Lex: How It Works](https://docs.aws.amazon.com/lex/latest/dg/how-it-works.html)

## Example Usage

```hcl
resource "aws_lex_bot" "order_flowers_bot" {
  abort_statement {
    message {
      content      = "Sorry, I am not able to assist at this time"
      content_type = "PlainText"
    }
  }

  child_directed = false

  clarification_prompt {
    max_attempts = 2

    message {
      content      = "I didn't understand you, what would you like to do?"
      content_type = "PlainText"
    }
  }

  description = "Bot to order flowers on the behalf of a user"

  idle_session_ttl_in_seconds = 600

  intent {
    intent_name    = "OrderFlowers"
    intent_version = "1"
  }

  locale           = "en-US"
  name             = "OrderFlowers"
  process_behavior = "BUILD"
  voice_id         = "Salli"
}
```

## Argument Reference

The following arguments are supported:

* `abort_statement` _(**Required**, Type: map)_:

	The message that Amazon Lex uses to abort a conversation. Attributes are documented under 
  [statement](#statement)..

* `child_directed` _(**Required**, Type: bool)_:

	By specifying true, you confirm that your use of Amazon Lex is related to a website, program,
	or other application that is directed or targeted, in whole or in part, to children under age
	13 and subject to COPPA.

	By specifying false, you confirm that your use of Amazon Lex is not related to a website,
	program, or other application that is directed or targeted, in whole or in part, to children
	under age 13 and subject to COPPA.

    If your use of Amazon Lex relates to a website, program, or other application that is directed
	in whole or in part, to children under age 13, you must obtain any required verifiable
	parental consent under COPPA. For information regarding the use of Amazon Lex in connection
	with websites, programs, or other applications that are directed or targeted, in whole or in
	part, to children under age 13, see the
	[Amazon Lex FAQ](https://aws.amazon.com/lex/faqs#data-security).

* `clarification_prompt` _(**Required**, Type: map)_:

	The message that Amazon Lex uses when it doesn't understand the user's request. Attributes 
  are documented under [prompt](#prompt)..

* `description` _(Optional, Type: string, Min: 0, Max: 200)_:

	A description of the bot.

* `idle_session_ttl_in_seconds` _(Optional, Type: number, Min: 60, Max: 86400, Default: 300)_:

	The maximum time in seconds that Amazon Lex retains the data gathered in a conversation.

* `locale` _(Optional, Type: string, Values: en-US | en-GB | de-DE, Default: en-US)_:

	Specifies the target locale for the bot. Any intent used in the bot must be compatible with
	the locale of the bot. *[String, values=en-US,en-GB,de-DE]*

* `intent` _(**Required**, Type: set)_:

	A set of Intent objects. Each intent represents a command that a user can express. Attributes 
  are documented under [intent](#intent-1)..

* `name` _(**Required**, Type: string, Min: 2, Max: 50, Regex: \^([A-Za-z]\_?)+$)_:

	The name of the bot that you want to create, case sensitive.

* `process_behavior` _(Optional, Type: string, Values: SAVE | BUILD, Default: SAVE)_:

	If you set the process_behavior element to BUILD , Amazon Lex builds the bot so that it can be
	run. If you set the element to SAVE Amazon Lex saves the bot, but doesn't build it.

* `voice_id` _(Optional, Type: string)_:

	The Amazon Polly voice ID that you want Amazon Lex to use for voice interactions with the
	user. The locale configured for the voice must match the locale of the bot. For more
	information, see [Available Voices](http://docs.aws.amazon.com/polly/latest/dg/voicelist.html)
	in the Amazon Polly Developer Guide.

### intent

Identifies the specific version of an intent.

* `intent_name` _(**Required**, Type: string, Min: 1, Max: 100, Regex: \^([A-Za-z]\_?)+$)_:

    The name of the intent.

* `intent_version` _(**Required**, Type: string, Min: 1, Max: 64, Regex: \$LATEST|[0-9]+)_:

    The version of the intent.

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
    documented under [message](#message-2).

* `response_card` _(Optional, Type: string, Min: 1, Max: 50000)_:

    The response card. Amazon Lex will substitute session attributes and slot values into the
    response card. For more information, see 
    [Example: Using a Response Card](https://docs.aws.amazon.com/lex/latest/dg/ex-resp-card.html).

### statement

A statement is a map with a set of message maps and an optional response card string. Messages
convey information to the user. At runtime, Amazon Lex selects the message to convey.

* `message` _(**Required**, Type: Set, Min: 1, Max: 15)_:

    A set of messages, each of which provides a message string and its type. You can specify the 
    message string in plain text or in Speech Synthesis Markup Language (SSML). Attributes are 
    documented under [message](#message-2).

* `response_card` _(Optional, Type: string, Min: 1, Max: 50000)_:

    The response card. Amazon Lex will substitute session attributes and slot values into the
    response card. For more information, see
    [Example: Using a Response Card](https://docs.aws.amazon.com/lex/latest/dg/ex-resp-card.html).

## Attributes Reference

The following attributes are exported in addition to the arguments listed above:

* `checksum`

	Checksum identifying the version of the bot that was created. The checksum is not included as
	an argument because the resource will add it automatically when updating the bot.

* `created_date`

	The date when the bot version was created.

* `failure_reason`

	If status is FAILED, Amazon Lex provides the reason that it failed to build the bot.

* `last_updated_date`

	The date when the $LATEST version of this bot was updated.

* `status` _(Values: BUILDING | READY | FAILED | NOT_BUILT)_:

	When you send a request to create or update a bot, Amazon Lex sets the status response element
	to BUILDING. After Amazon Lex builds the bot, it sets status to READY. If Amazon Lex can't
	build the bot, it sets status to FAILED. Amazon Lex returns the reason for the failure in the
	failure_reason response element.

* `version`

	The version of the bot.

## Import

Bots can be imported using their name.

```
$ terraform import aws_lex_bot.order_flowers_bot OrderFlowers
```
