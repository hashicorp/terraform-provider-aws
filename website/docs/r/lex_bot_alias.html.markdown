---
layout: "aws"
page_title: "AWS: aws_lex_bot_alias"
sidebar_current: "docs-aws-resource-lex-bot-alias"
description: |-
  Provides an Amazon Lex bot alias resource.
---

# aws_lex_bot_alias

Provides an [Amazon Lex](https://docs.aws.amazon.com/lex/latest/dg/what-is.html) bot alias resource.

## Example Usage

```hcl
resource "aws_lex_bot_alias" "florist_bot_v1" {
  bot_name = "FloristBot"
  bot_version = "1"
  description = "Version 1 of the Florist Bot."
  name = "FloristBotV1"
}
```

## Argument Reference

The following arguments are supported:

### Required

* `bot_name, type=string, min=1, max=100, pattern=^([A-Za-z]_?)+$`

	The name of the bot.

* `bot_version, type=string, min=1, max=64, pattern=\$LATEST|[0-9]+`

	The name of the bot.

* `name, type=string, min=1, max=100, pattern=^([A-Za-z]_?)+$`

	The name of the alias. The name is not case sensitive.

### Optional

* `description, type=string, min=0, max=200`

	A description of the alias.

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
$ terraform import aws_lex_bot_alias.florist_bot_v1 FloristBot.FloristBotV1
```
