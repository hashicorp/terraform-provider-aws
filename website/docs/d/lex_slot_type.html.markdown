---
subcategory: "Lex"
layout: "aws"
page_title: "AWS: aws_lex_slot_type"
description: |-
  Provides details about a specific Amazon Lex Slot Type
---

# Data Source: aws_lex_slot_type

Provides details about a specific Amazon Lex Slot Type.

## Example Usage

```hcl
data "aws_lex_slot_type" "flower_types" {
  name    = "FlowerTypes"
  version = "1"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the slot type. The name is case sensitive.
* `version` - (Required) The version of the slot type.

## Attributes Reference

All attributes are exported. See the [aws_lex_slot_type](/docs/providers/aws/r/lex_slot_type.html) 
resource for the full list.
