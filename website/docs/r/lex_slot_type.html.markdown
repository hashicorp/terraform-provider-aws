---
layout: "aws"
page_title: "AWS: aws_lex_slot_type"
sidebar_current: "docs-aws-resource-lex-slot-type"
description: |-
  Provides an Amazon Lex slot type resource.
---

# aws_lex_slot_type

Provides an [Amazon Lex](https://docs.aws.amazon.com/lex/latest/dg/what-is.html) slot type resource.
This creates a custom slot type.

## Example Usage

```hcl
resource "aws_lex_slot_type" "flowers" {
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

### Required

* `enumeration_value, type=list<EnumerationValue>, min=1, max=10000`

	A list of EnumerationValue objects that defines the values that the slot type can take. Each
	value can have a list of synonyms, which are additional values that help train the machine
	learning model about the values that it resolves for a slot.

* `name, type=string, min=1, max=100, pattern=^([A-Za-z]_?)+$`

	The name of the slot type. The name is not case sensitive.

### Optional

* `description, type=string, min=0, max=200`

	A description of the slot type.

* `value_selection_strategy, type=string, values=[ORIGINAL_VALUE | TOP_RESOLUTION], default=ORIGINAL_VALUE`

	Determines the slot resolution strategy that Amazon Lex uses to return slot type values.
	*ORIGINAL_VALUE* returns the value entered by the user if the user value is similar to the slot
	value. *TOP_RESOLUTION* returns the first value in the resolution list if there is a resolution
	list for the slot, otherwise null is returned.

### EnumerationValue

An enumeration value is a map with a value and an optional synonyms list.

#### Required

* `value, type=string, min=1, max=140`

	The value of the slot type.

#### Optional

* `synonyms, type=list, min=1, max=140`

    Additional values related to the slot type value.

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
$ terraform import aws_lex_slot_type.flowers FlowerTypes
```
