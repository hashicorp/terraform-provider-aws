---
subcategory: "EventBridge Schemas"
layout: "aws"
page_title: "AWS: aws_schemas_registry"
description: |-
  Provides an EventBridge Custom Schema Registry resource.
---

# Resource: aws_schemas_registry

Provides an EventBridge Custom Schema Registry resource.

~> **Note:** EventBridge was formerly known as CloudWatch Events. The functionality is identical.


## Example Usage

```terraform
resource "aws_schemas_registry" "test" {
  name        = "my_own_registry"
  description = "A custom schema registry"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the custom event schema registry. Maximum of 64 characters consisting of lower case letters, upper case letters, 0-9, ., -, _.
* `description` - (Optional) The description of the discoverer. Maximum of 256 characters.
* `tags` - (Optional)  A map of tags to assign to the resource.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The Amazon Resource Name (ARN) of the discoverer.


## Import

EventBridge schema registries can be imported using the `name`, e.g.

```console
$ terraform import aws_schemas_registry.test my_own_registry
```
