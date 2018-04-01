---
layout: "aws"
page_title: "AWS: aws_lex_bot_alias"
sidebar_current: "docs-aws-lex-bot-alias"
description: |-
    Provides details about a specific Lex bot alias
---

# Data Source: aws_lex_bot_alias

`aws_lex_bot_alias` provides details about a specific Lex bot alias.

## Example Usage

```hcl
data "aws_lex_bot_alias" "florist_bot_prod_alias" {
  bot_name = "FloristBot"
  name     = "FloristBotProd"
}
```

## Argument Reference

### Required

* `bot_name`

    The name of the bot.

* `name`

    The name of the bot alias. The name is case sensitive.

## Attributes Reference

All attributes are exported. See the aws_lex_bot_alias resource for the full list.
