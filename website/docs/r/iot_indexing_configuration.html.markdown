---
subcategory: "IoT Core"
layout: "aws"
page_title: "AWS: aws_iot_indexing_configuration"
description: |-
    Managing IoT Thing indexing.
---

# Resource: aws_iot_indexing_configuration

Managing [IoT Thing indexing](https://docs.aws.amazon.com/iot/latest/developerguide/managing-index.html).

## Example Usage

```terraform
resource "aws_iot_indexing_configuration" "example" {
  thing_indexing_configuration {
    thing_indexing_mode              = "REGISTRY_AND_SHADOW"
    thing_connectivity_indexing_mode = "STATUS"
    device_defender_indexing_mode    = "VIOLATIONS"
    named_shadow_indexing_mode       = "ON"

    custom_field {
      name = "shadow.desired.power"
      type = "Boolean"
    }
    custom_field {
      name = "attributes.version"
      type = "Number"
    }
    custom_field {
      name = "shadow.name.thing1shadow.desired.DefaultDesired"
      type = "String"
    }
    custom_field {
      name = "deviceDefender.securityProfile1.NUMBER_VALUE_BEHAVIOR.lastViolationValue.number"
      type = "Number"
    }
  }
}
```

## Argument Reference

* `thing_group_indexing_configuration` - (Optional) Thing group indexing configuration. See below.
* `thing_indexing_configuration` - (Optional) Thing indexing configuration. See below.

### thing_group_indexing_configuration

The `thing_group_indexing_configuration` configuration block supports the following:

* `custom_field` - (Optional) A list of thing group fields to index. This list cannot contain any managed fields. See below.
* `managed_field` - (Optional) Contains fields that are indexed and whose types are already known by the Fleet Indexing service. See below.
* `thing_group_indexing_mode` - (Required) Thing group indexing mode. Valid values: `OFF`, `ON`.

### thing_indexing_configuration

The `thing_indexing_configuration` configuration block supports the following:

* `custom_field` - (Optional) Contains custom field names and their data type. See below.
* `device_defender_indexing_mode` - (Optional) Device Defender indexing mode. Valid values: `VIOLATIONS`, `OFF`. Default: `OFF`.
* `managed_field` - (Optional) Contains fields that are indexed and whose types are already known by the Fleet Indexing service. See below.
* `named_shadow_indexing_mode` - (Optional) [Named shadow](https://docs.aws.amazon.com/iot/latest/developerguide/iot-device-shadows.html) indexing mode. Valid values: `ON`, `OFF`. Default: `OFF`.
* `thing_connectivity_indexing_mode` - (Optional) Thing connectivity indexing mode. Valid values: `STATUS`, `OFF`. Default: `OFF`.
* `thing_indexing_mode` - (Required) Thing indexing mode. Valid values: `REGISTRY`, `REGISTRY_AND_SHADOW`, `OFF`.

### field

The `custom_field` and `managed_field` configuration blocks supports the following:

* `name` - (Optional) The name of the field.
* `type` - (Optional) The data type of the field. Valid values: `Number`, `String`, `Boolean`.

## Attribute Reference

This resource exports no additional attributes.
