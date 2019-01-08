---
layout: "aws"
page_title: "AWS: aws_lex_bot_intent"
sidebar_current: "docs-aws-resource-lex-bot-intent"
description: |-
  Definition of an Amazon Lex Bot Intent used as an attribute in other Lex resources.
---

# aws_lex_bot_intent

Identifies the specific version of an intent.

## Example Usage

```hcl
resource "aws_lex_bot" "florist_bot" {
  intent {
    intent_name    = "OrderFlowers"
    intent_version = "1"
  }
}
```

## Argument Reference

The following arguments are supported:

### Required

* `intent_name`

	The name of the intent.

    * Type: string
    * Min: 1
    * Max: 100
    * Regex: ^([A-Za-z]_?)+$

* `intent_version`

	The version of the intent.

    * Type: string
    * Min: 1
    * Max: 64
    * Regex: \$LATEST|[0-9]+
