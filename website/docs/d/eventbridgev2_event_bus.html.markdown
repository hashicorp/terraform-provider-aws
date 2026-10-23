---
subcategory: "EventBridge V2"
layout: "aws"
page_title: "AWS: aws_eventbridgev2_event_bus"
description: |-
  Provides details about an Amazon EventBridge V2 event bus.
---

# Data Source: aws_eventbridgev2_event_bus

Provides details about an Amazon EventBridge V2 event bus, including an event bus shared with this account through AWS Resource Access Manager (AWS RAM).

## Example Usage

### Basic Usage

```terraform
data "aws_eventbridgev2_event_bus" "example" {
  arn = "arn:aws:events:us-east-1:123456789012:event-busv2/example/abcdefghij0123456789abcde"
}
```

### Event Bus Shared from Another Account

Use the [`aws_eventbridgev2_event_buses`](eventbridgev2_event_buses.html) data source to find the ARN of an event bus shared by its owner account.

```terraform
data "aws_eventbridgev2_event_buses" "example" {
  event_bus_account_id = "111122223333"
  name_prefix          = "orders"
}

data "aws_eventbridgev2_event_bus" "example" {
  arn = one([for v in data.aws_eventbridgev2_event_buses.example.event_buses : v.arn if v.name == "orders"])
}

resource "aws_eventbridgev2_subscriber" "example" {
  name          = "example"
  event_bus_arn = data.aws_eventbridgev2_event_bus.example.arn

  filter_configuration {
    filter {
      scope   = "METADATA"
      pattern = jsonencode({ source = ["example"] })
    }
  }

  invoke_configuration {
    role_arn   = aws_iam_role.example.arn
    target_arn = aws_sqs_queue.example.arn
  }
}
```

## Argument Reference

The following arguments are required:

* `arn` - (Required) ARN of the event bus. Can be an event bus in another account that is shared with this account through AWS RAM or whose [resource policy](../r/eventbridgev2_resource_policy.html) allows `events:DescribeEventBus`.

The following arguments are optional:

* `region` - (Optional) Region where this data source will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `creation_time` - Time the event bus was created, in RFC3339 format.
* `description` - Description of the event bus.
* `encryption_configuration` - Encryption configuration of the event bus. See [`encryption_configuration` Block](#encryption_configuration-block) below.
* `last_modified_time` - Time the event bus was last modified, in RFC3339 format.
* `name` - Name of the event bus.
* `state` - Lifecycle state of the event bus.
* `state_reason` - Explanation of why the event bus is in its current state. Only set when the event bus is not `ACTIVE`.
* `storage_configuration` - Event storage configuration of the event bus. See [`storage_configuration` Block](#storage_configuration-block) below.

### `encryption_configuration` Block

* `kms_key_identifier` - ARN of the AWS KMS customer managed key used to encrypt events.

### `storage_configuration` Block

* `retention_period_in_days` - Number of days events are retained on the event bus.
