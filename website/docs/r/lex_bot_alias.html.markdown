---
layout: "aws"
page_title: "AWS: aws_lex_bot_alias"
sidebar_current: "docs-aws-resource-lex-bot-alias"
description: |-
  Provides an Amazon Lex Bot Alias resource.
---

# Resource: aws_lex_bot_alias

Provides an Amazon Lex Bot Alias resource. For more information see
[Amazon Lex: How It Works](https://docs.aws.amazon.com/lex/latest/dg/how-it-works.html)

## Example Usage

```hcl
resource "aws_lex_bot_alias" "order_flowers_prod" {
  bot_name    = "OrderFlowers"
  bot_version = "1"
  description = "Production Version of the OrderFlowers Bot."
  name        = "OrderFlowersProd"
}
```

## Argument Reference

The following arguments are supported:

* `bot_name` _(**Required**, Type: string, Min: 1, Max: 100, Regex: \^([A-Za-z]\_?)+$)_:

	The name of the bot.

* `bot_version` _(**Required**, Type: string, Min: 1, Max: 64, Regex: \$LATEST|[0-9]+)_:

	The name of the bot.

* `description` _(Optional, Type: string, Min: 0, Max: 200)_:

	A description of the alias.

* `name` _(**Required**, Type: string, Min: 1, Max: 100, Regex: \^([A-Za-z]\_?)+$)_:

	The name of the alias. The name is not case sensitive.

## Attributes Reference

The following attributes are exported in addition to the arguments listed above:

* `checksum`

	The checksum for the current version of the alias. Note: The checksum is not included as an
	argument because the resource will add it automatically when updating the bot alias.

* `created_date`

	The date that the bot alias was created.

* `last_updated_date`

	The date that the bot alias was updated. When you create a resource, the creation date and the
	last updated date are the same.

## Import

Bot aliases can be imported using an ID with the format BotName.BotAliasName.

```
$ terraform import aws_lex_bot_alias.order_flowers_prod OrderFlowers.OrderFlowersProd
```
