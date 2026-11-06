---
subcategory: "EventBridge V2"
layout: "aws"
page_title: "AWS: aws_eventbridgev2_event_buses"
description: |-
  Lists Amazon EventBridge V2 event buses.
---

# Data Source: aws_eventbridgev2_event_buses

Lists Amazon EventBridge V2 event buses visible to this account, including event buses shared with this account through AWS Resource Access Manager (AWS RAM). Event buses that another account grants access to only through a [resource policy](../r/eventbridgev2_resource_policy.html) are not listed; use [`aws_eventbridgev2_event_bus`](eventbridgev2_event_bus.html) with the event bus ARN instead.

## Example Usage

### Basic Usage

```terraform
data "aws_eventbridgev2_event_buses" "example" {
  name_prefix = "example"
}
```

### Event Buses Shared from Another Account

```terraform
data "aws_eventbridgev2_event_buses" "example" {
  event_bus_account_id = "111122223333"
}
```

## Argument Reference

This data source supports the following arguments:

* `event_bus_account_id` - (Optional) ID of the AWS account that owns the event buses. If omitted, both owned event buses and event buses shared with this account are listed.
* `name_prefix` - (Optional) Name prefix used to filter event buses.
* `region` - (Optional) Region where this data source will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `event_buses` - List of event buses. For event buses shared with this account, only `account_id`, `arn`, and `name` are set. See [`event_buses` Block](#event_buses-block) below.

### `event_buses` Block

* `account_id` - ID of the AWS account that owns the event bus.
* `arn` - ARN of the event bus.
* `creation_time` - Time the event bus was created, in RFC3339 format.
* `description` - Description of the event bus.
* `last_modified_time` - Time the event bus was last modified, in RFC3339 format.
* `name` - Name of the event bus.
* `state` - Lifecycle state of the event bus.
* `state_reason` - Explanation of why the event bus is in its current state. Only set when the event bus is not `ACTIVE`.
