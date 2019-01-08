---
layout: "aws"
page_title: "AWS: aws_lex_bot_alias"
sidebar_current: "docs-aws-lex-bot-alias"
description: |-
    Provides details about a specific Lex Bot Alias
---

# Data Source: aws_lex_bot_alias

`aws_lex_bot_alias` provides details about a specific Lex Bot Alias.

## Example Usage

```hcl
data "aws_lex_bot_alias" "order_flowers_prod" {
  bot_name = "OrderFlowers"
  name     = "OrderFlowersProd"
}
```

## Argument Reference

### Required

* `bot_name`

    The name of the bot.

* `name`

    The name of the bot alias. The name is case sensitive.

## Attributes Reference

All attributes are exported. See the [aws_lex_bot_alias](/docs/providers/aws/r/lex_bot_alias.html)
resource for the full list.
