---
subcategory: "Lex Model Building"
layout: "aws"
page_title: "AWS: aws_lex_slot_type"
description: |-
  Provides details about a specific Amazon Lex Slot Type
---

# Data Source: aws_lex_slot_type

Provides details about a specific Amazon Lex Slot Type.

## Example Usage

```terraform
data "aws_lex_slot_type" "flower_types" {
  name    = "FlowerTypes"
  version = "1"
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `name` - (Required) Name of the slot type. The name is case sensitive.
* `version` - (Optional) Version of the slot type.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `checksum` - Checksum identifying the version of the slot type that was created. The checksum is
not included as an argument because the resource will add it automatically when updating the slot type.
* `created_date` - Date when the slot type version was created.
* `description` - Description of the slot type.
* `enumeration_value` - Set of EnumerationValue objects that defines the values that
the slot type can take. Each value can have a set of synonyms, which are additional values that help
train the machine learning model about the values that it resolves for a slot.
* `last_updated_date` - Date when the $LATEST version of this slot type was updated.
* `name` - Name of the slot type. The name is not case sensitive.
* `value_selection_strategy` - Determines the slot resolution strategy that Amazon Lex
uses to return slot type values. `ORIGINAL_VALUE` returns the value entered by the user if the user
value is similar to the slot value. `TOP_RESOLUTION` returns the first value in the resolution list
if there is a resolution list for the slot, otherwise null is returned.
* `version` - Version of the slot type.
