---
layout: "aws"
page_title: "AWS: aws_lex_enumeration_value"
sidebar_current: "docs-aws-resource-lex-enumeration-value"
description: |-
  Definition of an Amazon Lex Enumeration Value used as an attribute in other Lex resources.
---

# aws_lex_enumeration_value

Each slot type can have a set of values. Each enumeration value represents a value the slot type
can take.

## Example Usage

```hcl
resource "aws_lex_slot_type" "flowers" {
  enumeration_value {
    synonyms = [
      "Lirium",
      "Martagon",
    ]

    value = "lilies"
  }
}
```

## Argument Reference

The following arguments are supported:

### Required

* `value`

	The version of the intent.

    * Type: string
    * Min: 1
    * Max: 140

### Optional

* `synonyms`

	The name of the intent.

    * Type: string
    * Min: 1
    * Max: 140
