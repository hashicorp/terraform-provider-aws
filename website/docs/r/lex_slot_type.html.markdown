---
subcategory: "Lex Model Building"
layout: "aws"
page_title: "AWS: aws_lex_slot_type"
description: |-
  Provides details about a specific Amazon Lex Slot Type
---

# Resource: aws_lex_slot_type

Provides an Amazon Lex Slot Type resource. For more information see
[Amazon Lex: How It Works](https://docs.aws.amazon.com/lex/latest/dg/how-it-works.html)

## Example Usage

```terraform
resource "aws_lex_slot_type" "flower_types" {
  create_version = true
  description    = "Types of flowers to order"

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

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `enumeration_value` - (Required) A list of EnumerationValue objects that defines the values that
the slot type can take. Each value can have a list of synonyms, which are additional values that help
train the machine learning model about the values that it resolves for a slot. Attributes are
documented under [enumeration_value](#enumeration_value).
* `name` - (Required) The name of the slot type. The name is not case sensitive. Must be less than or equal to 100 characters in length.
* `create_version` - (Optional)
Determines if a new slot type version is created when the initial resource is created and on each
update. Defaults to `false`.
* `description` - (Optional) A description of the slot type. Must be less than or equal to 200 characters in length.
* `value_selection_strategy` - (Optional) Determines the slot resolution strategy that Amazon Lex
uses to return slot type values. `ORIGINAL_VALUE` returns the value entered by the user if the user
value is similar to the slot value. `TOP_RESOLUTION` returns the first value in the resolution list
if there is a resolution list for the slot, otherwise null is returned. Defaults to `ORIGINAL_VALUE`.

### enumeration_value

Each slot type can have a set of values. Each enumeration value represents a value the slot type
can take.

For example, a pizza ordering bot could have a slot type that specifies the type of crust that the
pizza should have. The slot type could include the values: thick, thin, stuffed.

* `synonyms` - (Optional) Additional values related to the slot type value. Each item must be less than or equal to 140 characters in length.
* `value` - (Required) The value of the slot type. Must be less than or equal to 140 characters in length.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `1m`)
* `update` - (Default `1m`)
* `delete` - (Default `5m`)

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `checksum` - Checksum identifying the version of the slot type that was created. The checksum is
not included as an argument because the resource will add it automatically when updating the slot type.
* `created_date` - The date when the slot type version was created.
* `last_updated_date` - The date when the `$LATEST` version of this slot type was updated.
* `version` - The version of the slot type.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import slot types using their name. For example:

```terraform
import {
  to = aws_lex_slot_type.flower_types
  id = "FlowerTypes"
}
```

Using `terraform import`, import slot types using their name. For example:

```console
% terraform import aws_lex_slot_type.flower_types FlowerTypes
```
