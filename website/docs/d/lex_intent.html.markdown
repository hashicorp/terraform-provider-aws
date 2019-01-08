---
layout: "aws"
page_title: "AWS: aws_lex_intent"
sidebar_current: "docs-aws-lex-intent"
description: |-
    Provides details about a specific Amazon Lex Intent
---

# Data Source: aws_lex_intent

`aws_lex_intent` provides details about a specific Amazon Lex Intent.

## Example Usage

```hcl
data "aws_lex_intent" "order_flowers" {
  name    = "OrderFlowers"
  version = "$LATEST"
}
```

## Argument Reference

### Required

* `name`

    The name of the slot type. The name is case sensitive.

### Optional

* `version`

    The version or alias of the slot type.

## Attributes Reference

The following attributes are exported. See the [aws_lex_intent](/docs/providers/aws/r/lex_intent.html)
resource for attribute descriptions.

* `checksum`
* `created_date`
* `description`
* `last_updated_date`
* `name`
* `parent_intent_signature`
* `version`
