---
layout: "aws"
page_title: "AWS: aws_lex_slot_type"
sidebar_current: "docs-aws-resource-lex-slot-type"
description: |-
  Provides an Amazon Lex Slot Type resource.
---

# Resource: aws_lex_slot_type

Provides an Amazon Lex Slot Type resource. For more information see
[Amazon Lex: How It Works](https://docs.aws.amazon.com/lex/latest/dg/how-it-works.html)

## Example Usage

```hcl
resource "aws_lex_slot_type" "flower_types" {
  description = "Types of flowers to order"

  enumeration_value {
    synonyms = [
      "Lirium",
      "Martagon",
    ]

    value = "lilies"
  }

  enumeration_value {
    synonyms = [
      "Eduardoregelia",
      "Podonix",
    ]

    value = "tulips"
  }

  name                     = "FlowerTypes"
  value_selection_strategy = "ORIGINAL_VALUE"
}
```

## Argument Reference

The following arguments are supported:

* `description` _(Optional, Type: string, Min: 0, Max: 200)_:

	A description of the slot type.

* `enumeration_value` _(**Required**, Type: list, Min: 1, Max: 10000)_:

	A list of EnumerationValue objects that defines the values that the slot type can take. Each
	value can have a list of synonyms, which are additional values that help train the machine
	learning model about the values that it resolves for a slot. Attributes are documented under
  [enumeration_value](#enumeration_value-1).

* `name` _(**Required**, Type: string, Min: 1, Max: 100, Regex: \^([A-Za-z]\_?)+$)_:

	The name of the slot type. The name is not case sensitive.

* `value_selection_strategy` _(Optional, Type: string, Values: ORIGINAL_VALUE | TOP_RESOLUTION, Default: ORIGINAL_VALUE)_:

	Determines the slot resolution strategy that Amazon Lex uses to return slot type values.
	*ORIGINAL_VALUE* returns the value entered by the user if the user value is similar to the slot
	value. *TOP_RESOLUTION* returns the first value in the resolution list if there is a resolution
	list for the slot, otherwise null is returned.

### enumeration_value

Each slot type can have a set of values. Each enumeration value represents a value the slot type
can take.

For example, a pizza ordering bot could have a slot type that specifies the type of crust that the pizza should have. The slot type could include the values: thick, thin, stuffed.

* `synonyms` _(Optional, Type: list of strings, Min: 1, Max:140)_:

    Additional values related to the slot type value.

* `value` _(Required, Type: string, Min: 1, Max: 140)_:

	  The value of the slot type.

### Timeouts

The `timeouts` block allows you to specify [timeouts](https://www.terraform.io/docs/configuration/resources.html#timeouts) for certain actions:

* `delete` - (Defaults to 5 mins) Used when deleting the slot type

## Attributes Reference

The following attributes are exported in addition to the arguments listed above:

* `checksum`

	Checksum identifying the version of the slot type that was created. The checksum is not included
	as an argument because the resource will add it automatically when updating the slot type.

* `created_date`

	The date when the slot type version was created.

* `last_updated_date`

	The date when the $LATEST version of this slot type was updated.

* `version`

	The version of the slot type.

## Import

Slot types can be imported using their name.

```
$ terraform import aws_lex_slot_type.flower_types FlowerTypes
```
