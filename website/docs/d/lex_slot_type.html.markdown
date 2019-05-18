---
layout: "aws"
page_title: "AWS: aws_lex_slot_type"
sidebar_current: "docs-aws-lex-slot-type"
description: |-
    Provides details about a specific Amazon Lex Slot Type
---

# Data Source: aws_lex_slot_type

Provides details about a specific Amazon Lex Slot Type.

## Example Usage

```hcl
data "aws_lex_slot_type" "flower_types" {
  name    = "FlowerTypes"
  version = "$LATEST"
}
```

## Argument Reference

The following arguments are supported:

* `name` _(Required)_:

    The name of the slot type. The name is case sensitive.

* `version` _(Required)_:

    The version or alias of the slot type.

## Attributes Reference

The following attributes are exported. See the [aws_lex_slot_type](/docs/providers/aws/r/lex_slot_type.html)
resource for attribute descriptions.

* `checksum`
* `created_date`
* `description`
* `last_updated_date`
* `name`
* `version`
