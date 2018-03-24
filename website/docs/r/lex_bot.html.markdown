---
layout: "aws"
page_title: "AWS: aws_lex_bot"
sidebar_current: "docs-aws-resource-lex-bot"
description: |-
  Provides an Amazon Lex bot resource.
---

# aws_lex_bot

Provides an [Amazon Lex](https://docs.aws.amazon.com/lex/latest/dg/what-is.html) bot resource.

## Example Usage

```hcl
resource "aws_lex_bot" "florist_bot" {
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

  locale = "en-US"
  name = "FloristBot"
  process_behavior = "BUILD"
  voice_id = "Salli"
}
```

## Argument Reference

The following arguments are supported:

### Required

* `abort_statement, type=Statement`

	The message that Amazon Lex uses to abort a conversation.

* `child_directed, type=bool`

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

* `clarification_prompt, type=Prompt`

	The message that Amazon Lex uses when it doesn't understand the user's request.

* `intent, type=set<Intent>`

	A set of Intent objects. Each intent represents a command that a user can express.

* `name, type=string, min=2, max=50, pattern=^([A-Za-z]_?)+$`

	The name of the bot that you want to create a new version of. The name is case sensitive.

### Optional

* `description, type=string, min=0, max=200`

	A description of the bot.

* `idle_session_ttl_in_seconds, type=number, min=60, max=86400, default=300`

	The maximum time in seconds that Amazon Lex retains the data gathered in a conversation.

* `locale, type=string, values=[en-US | en-GB | de-DE]`

	Specifies the target locale for the bot. Any intent used in the bot must be compatible with
	the locale of the bot. *[String, values=en-US,en-GB,de-DE]*

* `process_behavior, type=string, values=[SAVE | BUILD], default=BUILD`

	If you set the process_behavior element to BUILD , Amazon Lex builds the bot so that it can be
	run. If you set the element to SAVE Amazon Lex saves the bot, but doesn't build it.

* `voice_id, type=string, `

	The Amazon Polly voice ID that you want Amazon Lex to use for voice interactions with the
	user. The locale configured for the voice must match the locale of the bot. For more
	information, see [Available Voices](http://docs.aws.amazon.com/polly/latest/dg/voicelist.html)
	in the Amazon Polly Developer Guide .

### Statement

A statement is a map with a set of message maps and an optional response card string. Messages
convey information to the user. At runtime, Amazon Lex selects the message to convey.

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

    response_card = ""
  }
}
```

#### Required

* `message, type=set<Message>, min=1, max=15`

	A set of message maps. See the Message specification below.

#### Optional

* `response_card, type=string, min=1, max=50000`

    The response card. Amazon Lex will substitute session attributes and slot values into the
    response card. For more information, see
	[Example: Using a Response Card](https://docs.aws.amazon.com/lex/latest/dg/ex-resp-card.html).

### Prompt

A prompt is a map with a set of message maps, max attempts int, and an optional response card
string. Prompts obtain information from the user. For more information, see
[Amazon Lex: How It Works](https://docs.aws.amazon.com/lex/latest/dg/how-it-works.html).

```hcl
resource "aws_lex_bot" "florist_bot" {
  clarification_prompt {
    max_attempts = 2

    message {
      content      = "I didn't understand you, what would you like to do?"
      content_type = "PlainText"
    }

    message {
      content      = "Sorry, I don't understand, what would you like to do?"
      content_type = "PlainText"
    }

    response_card = ""
  }
}
```

#### Required

* `max_attempts, type=number, min=1, max=5`

	The number of times to prompt the user for information.

* `message, type=set<Message>, min=1, max=15`

	A set of message maps. See the Message specification below.

#### Optional

* `response_card, type=string, min=1, max=50000`

    The response card. Amazon Lex will substitute session attributes and slot values into the
    response card. For more information, see
	[Example: Using a Response Card](https://docs.aws.amazon.com/lex/latest/dg/ex-resp-card.html).

### Message

A message is a map with content and content type strings.

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

    response_card = ""
  }
}
```

#### Required

* `content, type=string, min=1, max=1000`

	The text of the message.

* `content_type, type=string, values=[PlainText | SSML | CustomPayload]`

	The content type of the message string.

#### Optional

* `group_number, type=number, min=1, max=5`

	Identifies the message group that the message belongs to. When a group is assigned to a
	message, Amazon Lex returns one message from each group in the response.

### Intent

An intent is a map with intent name and intent version strings.

```hcl
resource "aws_lex_bot" "florist_bot" {
  intent {
    intent_name    = "OrderFlowers"
    intent_version = "1"
  }

  intent {
    intent_name    = "CheckOrderStatus"
    intent_version = "1"
  }
}
```

#### Required

* `intent_name, type=string, min=1, max=100, pattern=^([A-Za-z]_?)+$`

	The name of the intent.

* `intent_version, type=string, min=1, max=64, pattern=\$LATEST|[0-9]+`

	The version of the intent.

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

* `status, values=[BUILDING | READY | FAILED | NOT_BUILT]`

	When you send a request to create or update a bot, Amazon Lex sets the status response element
	to BUILDING. After Amazon Lex builds the bot, it sets status to READY. If Amazon Lex can't
	build the bot, it sets status to FAILED. Amazon Lex returns the reason for the failure in the
	failure_reason response element.

* `version`

	The version of the bot.

## Import

Bots can be imported using their name.

```
$ terraform import aws_lex_bot.florist_bot FloristBot
```
