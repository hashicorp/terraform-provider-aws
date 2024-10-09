---
subcategory: "EventBridge"
layout: "aws"
page_title: "AWS: aws_cloudwatch_event_bus"
description: |-
  Provides an EventBridge event bus resource.
---

# Resource: aws_cloudwatch_event_bus

Provides an EventBridge event bus resource.

~> **Note:** EventBridge was formerly known as CloudWatch Events. The functionality is identical.

## Example Usage

```terraform
resource "aws_cloudwatch_event_bus" "messenger" {
  name = "chat-messages"
}
```

```terraform
data "aws_cloudwatch_event_source" "examplepartner" {
  name_prefix = "aws.partner/examplepartner.com"
}

resource "aws_cloudwatch_event_bus" "examplepartner" {
  name              = data.aws_cloudwatch_event_source.examplepartner.name
  event_source_name = data.aws_cloudwatch_event_source.examplepartner.name
}
```

## Argument Reference

This resource supports the following arguments:

* `name` - (Required) The name of the new event bus. The names of custom event buses can't contain the / character. To create a partner event bus, ensure the `name` matches the `event_source_name`.
* `event_source_name` - (Optional) The partner event source that the new event bus will be matched with. Must match `name`.
* `kms_key_identifier` - (Optional) The identifier of the AWS KMS customer managed key for EventBridge to use, if you choose to use a customer managed key to encrypt events on this event bus. The identifier can be the key Amazon Resource Name (ARN), KeyId, key alias, or key alias ARN.
* `tags` - (Optional)  A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The Amazon Resource Name (ARN) of the event bus.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import EventBridge event buses using the `name` (which can also be a partner event source name). For example:

```terraform
import {
  to = aws_cloudwatch_event_bus.messenger
  id = "chat-messages"
}
```

Using `terraform import`, import EventBridge event buses using the `name` (which can also be a partner event source name). For example:

```console
% terraform import aws_cloudwatch_event_bus.messenger chat-messages
```
