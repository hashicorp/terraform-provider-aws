---
layout: "aws"
page_title: "AWS: aws_lex_bot"
sidebar_current: "docs-aws-lex-bot"
description: |-
    Provides details about a specific Lex bot
---

# Data Source: aws_lex_bot

`aws_lex_bot` provides details about a specific Lex bot.

## Example Usage

```hcl
data "aws_lex_bot" "florist_bot_latest" {
  name             = "FloristBot"
  version_or_alias = "$LATEST"
}
```

## Argument Reference

### Required

* `name`

    The name of the bot. The name is case sensitive.

* `version_or_alias`

    The version or alias of the bot.

## Attributes Reference

The following attributes are exported. See the [aws_lex_bot](/docs/providers/aws/r/lex_bot.html)
resource for attribute descriptions.

- `name`
- `checksum`
- `child_directed`
- `created_date`
- `description`
- `failure_reason`
- `idle_session_ttl_in_seconds`
- `last_updated_date`
- `locale`
- `status`
- `version`
- `voice_id`
