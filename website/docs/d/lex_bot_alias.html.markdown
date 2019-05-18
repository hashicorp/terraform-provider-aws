---
layout: "aws"
page_title: "AWS: aws_lex_bot_alias"
sidebar_current: "docs-aws-lex-bot-alias"
description: |-
    Provides details about a specific Lex Bot Alias
---

# Data Source: aws_lex_bot_alias

Provides details about a specific Amazon Lex Bot Alias.

## Example Usage

```hcl
data "aws_lex_bot_alias" "order_flowers_prod" {
  bot_name = "OrderFlowers"
  name     = "OrderFlowersProd"
}
```

## Argument Reference

The following arguments are supported:

* `bot_name` _(Required)_:

    The name of the bot.

* `name` _(Required)_:

    The name of the bot alias. The name is case sensitive.

## Attributes Reference

All attributes are exported. See the [aws_lex_bot_alias](/docs/providers/aws/r/lex_bot_alias.html)
resource for the full list.
