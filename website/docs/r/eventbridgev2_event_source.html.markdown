---
subcategory: "EventBridge V2"
layout: "aws"
page_title: "AWS: aws_eventbridgev2_event_source"
description: |-
  Manages an Amazon EventBridge V2 event source.
---

# Resource: aws_eventbridgev2_event_source

Manages an Amazon EventBridge V2 event source. An event source forwards events from an AWS service or a partner event source to an EventBridge V2 event bus.

## Example Usage

### AWS Service Events

```terraform
resource "aws_eventbridgev2_event_bus" "example" {
  name = "example"
}

resource "aws_eventbridgev2_event_source" "example" {
  name          = "example"
  event_bus_arn = aws_eventbridgev2_event_bus.example.arn

  configuration {
    aws_service_events_configuration {
      aws_service = "aws.s3"
      pattern = jsonencode({
        "detail-type" = ["Object Created"]
      })
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `configuration` - (Required) Configuration of the events forwarded to the event bus. See [`configuration` Block](#configuration-block) below.
* `event_bus_arn` - (Required) ARN of the event bus to forward events to. Can be an event bus shared with this account through AWS RAM. Changing this value creates a new resource.
* `name` - (Required) Name of the event source. Must start with an alphanumeric character, contain only alphanumeric characters, periods (`.`), hyphens (`-`), and underscores (`_`), and must not begin with `aws.`. Changing this value creates a new resource.

The following arguments are optional:

* `description` - (Optional) Description of the event source.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Map of tags assigned to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `configuration` Block

The `configuration` block supports the following. Exactly one of `aws_service_events_configuration` or `partner_events_configuration` must be specified. Changing between them creates a new resource.

* `aws_service_events_configuration` - (Optional) Forwards events from an AWS service. See [`aws_service_events_configuration` Block](#aws_service_events_configuration-block) below.
* `partner_events_configuration` - (Optional) Forwards events from a partner event source. See [`partner_events_configuration` Block](#partner_events_configuration-block) below.

### `aws_service_events_configuration` Block

The `aws_service_events_configuration` block supports the following:

* `aws_service` - (Required) AWS service source identifier, for example `aws.s3`. Wildcards and lists are not supported.
* `on_failure_configuration` - (Optional) Destination for events that could not be forwarded. See [`on_failure_configuration` Block](#on_failure_configuration-block) below.
* `pattern` - (Optional) JSON event pattern that selects which of the service's events are forwarded. Do not include `source`, `account`, or `region` as top-level fields. If omitted, all events from the service are forwarded.

### `partner_events_configuration` Block

The `partner_events_configuration` block supports the following:

* `on_failure_configuration` - (Optional) Destination for events that could not be forwarded. See [`on_failure_configuration` Block](#on_failure_configuration-block) below.
* `partner_bus_kms_key_identifier` - (Optional) Identifier of the AWS KMS customer managed key used to encrypt events on the managed partner event bus. Can be a key ID, key ARN, alias name, or alias ARN.
* `partner_event_source_arn` - (Required) ARN of the partner event source to forward.
* `pattern` - (Optional) JSON event pattern that selects which of the partner event source's events are forwarded. If omitted, all events are forwarded.

### `on_failure_configuration` Block

The `on_failure_configuration` block supports the following:

* `arn` - (Required) ARN of the Amazon SQS queue that receives events that could not be forwarded.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the event source.
* `creation_time` - Time the event source was created, in RFC3339 format.
* `last_modified_time` - Time the event source was last modified, in RFC3339 format.
* `revoked` - Whether the event bus owner has revoked the event source.
* `state` - Lifecycle state of the event source.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_eventbridgev2_event_source.example
  identity = {
    "arn" = "arn:aws:events:us-east-1:123456789012:event-sourcev2/aws.service/example/abcdefghij0123456789abcde"
  }
}

resource "aws_eventbridgev2_event_source" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

- `arn` (String) ARN of the event source.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import EventBridge V2 event sources using the `arn`. For example:

```terraform
import {
  to = aws_eventbridgev2_event_source.example
  id = "arn:aws:events:us-east-1:123456789012:event-sourcev2/aws.service/example/abcdefghij0123456789abcde"
}
```

Using `terraform import`, import EventBridge V2 event sources using the `arn`. For example:

```console
% terraform import aws_eventbridgev2_event_source.example arn:aws:events:us-east-1:123456789012:event-sourcev2/aws.service/example/abcdefghij0123456789abcde
```
