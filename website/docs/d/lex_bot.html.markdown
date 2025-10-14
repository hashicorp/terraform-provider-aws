---
subcategory: "Lex Model Building"
layout: "aws"
page_title: "AWS: aws_lex_bot"
description: |-
  Provides details about a specific Lex Bot
---

# Data Source: aws_lex_bot

Provides details about a specific Amazon Lex Bot.

## Example Usage

```terraform
data "aws_lex_bot" "order_flowers_bot" {
  name    = "OrderFlowers"
  version = "$LATEST"
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `name` - (Required) Name of the bot. The name is case sensitive.
* `version` - (Optional) Version or alias of the bot.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the bot.
* `checksum` - Checksum of the bot used to identify a specific revision of the bot's `$LATEST` version.
* `child_directed` - If this Amazon Lex Bot is related to a website, program, or other application that is directed or targeted, in whole or in part, to children under age 13 and subject to COPPA.
* `created_date` - Date that the bot was created.
* `description` - Description of the bot.
* `detect_sentiment` - When set to true user utterances are sent to Amazon Comprehend for sentiment analysis.
* `enable_model_improvements` - Set to true if natural language understanding improvements are enabled.
* `failure_reason` - If the `status` is `FAILED`, the reason why the bot failed to build.
* `idle_session_ttl_in_seconds` - The maximum time in seconds that Amazon Lex retains the data gathered in a conversation.
* `last_updated_date` - Date that the bot was updated.
* `locale` - Target locale for the bot. Any intent used in the bot must be compatible with the locale of the bot.
* `name` - Name of the bot, case sensitive.
* `nlu_intent_confidence_threshold` - The threshold where Amazon Lex will insert the AMAZON.FallbackIntent, AMAZON.KendraSearchIntent, or both when returning alternative intents in a PostContent or PostText response. AMAZON.FallbackIntent and AMAZON.KendraSearchIntent are only inserted if they are configured for the bot.
* `status` - Status of the bot.
* `version` - Version of the bot. For a new bot, the version is always `$LATEST`.
* `voice_id` - Amazon Polly voice ID that the Amazon Lex Bot uses for voice interactions with the user.
