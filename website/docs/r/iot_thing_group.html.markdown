---
subcategory: "IoT Core"
layout: "aws"
page_title: "AWS: aws_iot_thing_group"
description: |-
    Manages an AWS IoT Thing Group.
---

# Resource: aws_iot_thing_group

Manages an AWS IoT Thing Group.

## Example Usage

```terraform
resource "aws_iot_thing_group" "parent" {
  name = "parent"
}

resource "aws_iot_thing_group" "example" {
  name = "example"

  parent_group_name = aws_iot_thing_group.parent.name

  properties {
    attribute_payload {
      attributes = {
        One = "11111"
        Two = "TwoTwo"
      }
    }
    description = "This is my thing group"
  }

  tags = {
    terraform = "true"
  }
}
```

## Argument Reference

* `name` - (Required) The name of the Thing Group.
* `parent_group_name` - (Optional) The name of the parent Thing Group.
* `properties` - (Optional) The Thing Group properties. Defined below.
* `tags` - (Optional) Key-value mapping of resource tags

### properties Reference

* `attribute_payload` - (Optional) The Thing Group attributes. Defined below.
* `description` - (Optional) A description of the Thing Group.

### attribute_payload Reference

* `attributes` - (Optional) Key-value map.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The ARN of the Thing Group.
* `id` - The Thing Group ID.
* `version` - The current version of the Thing Group record in the registry.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import IoT Things Groups using the name. For example:

```terraform
import {
  to = aws_iot_thing_group.example
  id = "example"
}
```

Using `terraform import`, import IoT Things Groups using the name. For example:

```console
% terraform import aws_iot_thing_group.example example
```
