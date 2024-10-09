---
subcategory: "Lex Model Building"
layout: "aws"
page_title: "AWS: aws_lex_bot_alias"
description: |-
  Provides details about a specific Lex Bot Alias
---

# Data Source: aws_lex_bot_alias

Provides details about a specific Amazon Lex Bot Alias.

## Example Usage

```terraform
data "aws_lex_bot_alias" "order_flowers_prod" {
  bot_name = "OrderFlowers"
  name     = "OrderFlowersProd"
}
```

## Argument Reference

This data source supports the following arguments:

* `bot_name` - (Required) Name of the bot.
* `name` - (Required) Name of the bot alias. The name is case sensitive.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the bot alias.
* `bot_name` - Name of the bot.
* `bot_version` - Version of the bot that the alias points to.
* `checksum` - Checksum of the bot alias.
* `created_date` - Date that the bot alias was created.
* `description` - Description of the alias.
* `last_updated_date` - Date that the bot alias was updated. When you create a resource, the creation date and the last updated date are the same.
* `name` - Name of the alias. The name is not case sensitive.
