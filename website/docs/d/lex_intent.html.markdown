---
subcategory: "Lex Model Building"
layout: "aws"
page_title: "AWS: aws_lex_intent"
description: |-
  Provides details about a specific Amazon Lex Intent
---

# Data Source: aws_lex_intent

Provides details about a specific Amazon Lex Intent.

## Example Usage

```terraform
data "aws_lex_intent" "order_flowers" {
  name    = "OrderFlowers"
  version = "$LATEST"
}
```

## Argument Reference

This data source supports the following arguments:

* `name` - (Required) Name of the intent. The name is case sensitive.
* `version` - (Optional) Version of the intent.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Lex intent.
* `checksum` - Checksum identifying the version of the intent that was created. The checksum is not
included as an argument because the resource will add it automatically when updating the intent.
* `created_date` - Date when the intent version was created.
* `description` - Description of the intent.
* `last_updated_date` - Date when the $LATEST version of this intent was updated.
* `name` - Name of the intent, not case sensitive.
* `parent_intent_signature` - A unique identifier for the built-in intent to base this
intent on. To find the signature for an intent, see
[Standard Built-in Intents](https://developer.amazon.com/public/solutions/alexa/alexa-skills-kit/docs/built-in-intent-ref/standard-intents)
in the Alexa Skills Kit.
* `version` - Version of the bot.
