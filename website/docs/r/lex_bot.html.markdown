---
layout: "aws"
page_title: "AWS: aws_lex_bot"
sidebar_current: "docs-aws-resource-lex-bot"
description: |-
  Provides an Amazon Lex bot resource.
---

# aws_lex_bot

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

### Required

* `abort_statement`

	The message that Amazon Lex uses to abort a conversation.

    * Type: [Lex Statement](/docs/providers/aws/r/lex_statement.html)

* `child_directed`

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

    * Type: bool

* `clarification_prompt`

	The message that Amazon Lex uses when it doesn't understand the user's request.

    * Type: [Lex Prompt](/docs/providers/aws/r/lex_prompt.html)

* `intent`

	A set of Intent objects. Each intent represents a command that a user can express.

    * Type: Set of [Lex Bot Intents](/docs/providers/aws/r/lex_bot_intent.html)

* `name`

	The name of the bot that you want to create, case sensitive.

    * Type: string
    * Min: 2
    * Max: 50
    * Pattern: ^([A-Za-z]_?)+$

### Optional

* `description`

	A description of the bot.

    * Type: string
    * Min: 0
    * Max: 200

* `idle_session_ttl_in_seconds`

	The maximum time in seconds that Amazon Lex retains the data gathered in a conversation.

    * Type: number
    * Min: 60
    * Max: 86400
    * Default: 300

* `locale`

	Specifies the target locale for the bot. Any intent used in the bot must be compatible with
	the locale of the bot. *[String, values=en-US,en-GB,de-DE]*

    * Type: string
    * Values: en-US | en-GB | de-DE
    * Default: en-US

* `process_behavior`

	If you set the process_behavior element to BUILD , Amazon Lex builds the bot so that it can be
	run. If you set the element to SAVE Amazon Lex saves the bot, but doesn't build it.

    * Type: string
    * Values: SAVE | BUILD
    * Default: SAVE

* `voice_id`

	The Amazon Polly voice ID that you want Amazon Lex to use for voice interactions with the
	user. The locale configured for the voice must match the locale of the bot. For more
	information, see [Available Voices](http://docs.aws.amazon.com/polly/latest/dg/voicelist.html)
	in the Amazon Polly Developer Guide .

    * Type: string

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

* `status`

	When you send a request to create or update a bot, Amazon Lex sets the status response element
	to BUILDING. After Amazon Lex builds the bot, it sets status to READY. If Amazon Lex can't
	build the bot, it sets status to FAILED. Amazon Lex returns the reason for the failure in the
	failure_reason response element.

    * Values: BUILDING | READY | FAILED | NOT_BUILT

* `version`

	The version of the bot.

## Import

Bots can be imported using their name.

```
$ terraform import aws_lex_bot.order_flowers_bot OrderFlowers
```
