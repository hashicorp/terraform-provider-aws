---
subcategory: "EventBridge Schemas"
layout: "aws"
page_title: "AWS: aws_schemas_discoverer"
description: |-
  Provides an EventBridge Schema Discoverer resource.
---

# Resource: aws_schemas_discoverer

Provides an EventBridge Schema Discoverer resource.

~> **Note:** EventBridge was formerly known as CloudWatch Events. The functionality is identical.

## Example Usage

```terraform
resource "aws_cloudwatch_event_bus" "messenger" {
  name = "chat-messages"
}

resource "aws_schemas_discoverer" "test" {
  source_arn  = aws_cloudwatch_event_bus.messenger.arn
  description = "Auto discover event schemas"
}
```

## Argument Reference

This resource supports the following arguments:

* `source_arn` - (Required) The ARN of the event bus to discover event schemas on.
* `description` - (Optional) The description of the discoverer. Maximum of 256 characters.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The Amazon Resource Name (ARN) of the discoverer.
* `id` - The ID of the discoverer.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import EventBridge discoverers using the `id`. For example:

```terraform
import {
  to = aws_schemas_discoverer.test
  id = "123"
}
```

Using `terraform import`, import EventBridge discoverers using the `id`. For example:

```console
% terraform import aws_schemas_discoverer.test 123
```
