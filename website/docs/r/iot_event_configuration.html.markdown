---
subcategory: "IoT"
layout: "aws"
page_title: "AWS: aws_iot_event_configuration"
description: |-
    Creates and manages an AWS IoT Event Configuration.
---

# Resource: aws_iot_event_configuration

Creates and manages an AWS IoT Event Configuration

## Example Usage

```hcl
resource "aws_iot_event_configuration" "example" {
  name = "example"
  configurations_map {
    attribute_name = "example_name"
    enabled        = true
  }

  configurations_map {
    attribute_name = "example_name_2"
    enabled        = false
  }
}
```

## Argument Reference

* `name` - (Required) The name of the event configuration.
* `configurations_map` - (Required) Map of attributes of the event.

## Import

IOT Things can be imported using the name, e.g.

```
$ terraform import aws_iot_event_configuration.example example
```
