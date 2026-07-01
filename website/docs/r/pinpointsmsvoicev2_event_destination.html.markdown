---
subcategory: "End User Messaging SMS"
layout: "aws"
page_title: "AWS: aws_pinpointsmsvoicev2_event_destination"
description: |-
  Manages an AWS End User Messaging SMS Event Destination.
---

# Resource: aws_pinpointsmsvoicev2_event_destination

Manages an AWS End User Messaging SMS Event Destination.

An event destination is a location where messaging events are published. Exactly one of `cloudwatch_logs_destination`, `kinesis_firehose_destination`, or `sns_destination` must be configured per event destination. Changing the sink type (e.g., from `sns_destination` to `cloudwatch_logs_destination`) forces resource replacement — AWS's `UpdateEventDestination` rejects sink-type changes with `ConflictException`.

## Example Usage

### CloudWatch Logs Destination

```terraform
resource "aws_pinpointsmsvoicev2_configuration_set" "example" {
  name = "example-configuration-set"
}

resource "aws_pinpointsmsvoicev2_event_destination" "example" {
  configuration_set_name = aws_pinpointsmsvoicev2_configuration_set.example.name
  event_destination_name = "example"
  matching_event_types   = ["ALL"]

  cloudwatch_logs_destination {
    iam_role_arn  = aws_iam_role.example.arn
    log_group_arn = aws_cloudwatch_log_group.example.arn
  }
}
```

### Kinesis Firehose Destination

```terraform
resource "aws_pinpointsmsvoicev2_configuration_set" "example" {
  name = "example-configuration-set"
}

resource "aws_pinpointsmsvoicev2_event_destination" "example" {
  configuration_set_name = aws_pinpointsmsvoicev2_configuration_set.example.name
  event_destination_name = "example"
  matching_event_types   = ["ALL"]

  kinesis_firehose_destination {
    delivery_stream_arn = aws_kinesis_firehose_delivery_stream.example.arn
    iam_role_arn        = aws_iam_role.example.arn
  }
}
```

### SNS Destination

```terraform
resource "aws_pinpointsmsvoicev2_configuration_set" "example" {
  name = "example-configuration-set"
}

resource "aws_pinpointsmsvoicev2_event_destination" "example" {
  configuration_set_name = aws_pinpointsmsvoicev2_configuration_set.example.name
  event_destination_name = "example"
  matching_event_types   = ["ALL"]

  sns_destination {
    topic_arn = aws_sns_topic.example.arn
  }
}
```

## Argument Reference

The following arguments are required:

* `configuration_set_name` - (Required) Name of the configuration set this event destination belongs to. Changing this forces a new resource.
* `event_destination_name` - (Required) Name of the event destination. Changing this forces a new resource.
* `matching_event_types` - (Required) Event types for which the destination receives records. See the [AWS API reference](https://docs.aws.amazon.com/pinpoint/latest/apireference_smsvoicev2/API_CreateEventDestination.html#pinpoint-CreateEventDestination-request-MatchingEventTypes) for valid values.

The following arguments are optional:

* `cloudwatch_logs_destination` - (Optional) Send events to Amazon CloudWatch Logs. Exactly one of `cloudwatch_logs_destination`, `kinesis_firehose_destination`, or `sns_destination` must be configured. See [`cloudwatch_logs_destination` Block](#cloudwatch_logs_destination-block) for details.
* `enabled` - (Optional) Whether the event destination is enabled. Defaults to `true`.
* `kinesis_firehose_destination` - (Optional) Send events to Amazon Data Firehose. Exactly one of `cloudwatch_logs_destination`, `kinesis_firehose_destination`, or `sns_destination` must be configured. See [`kinesis_firehose_destination` Block](#kinesis_firehose_destination-block) for details.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `sns_destination` - (Optional) Send events to Amazon SNS. Exactly one of `cloudwatch_logs_destination`, `kinesis_firehose_destination`, or `sns_destination` must be configured. See [`sns_destination` Block](#sns_destination-block) for details.

### `cloudwatch_logs_destination` Block

The `cloudwatch_logs_destination` configuration block supports the following arguments:

* `iam_role_arn` - (Required) ARN of the IAM role that End User Messaging SMS assumes to write to the log group.
* `log_group_arn` - (Required) ARN of the Amazon CloudWatch log group that receives the events.

### `kinesis_firehose_destination` Block

The `kinesis_firehose_destination` configuration block supports the following arguments:

* `delivery_stream_arn` - (Required) ARN of the Amazon Data Firehose delivery stream that receives the events.
* `iam_role_arn` - (Required) ARN of the IAM role that End User Messaging SMS assumes to write to the delivery stream.

### `sns_destination` Block

The `sns_destination` configuration block supports the following arguments:

* `topic_arn` - (Required) ARN of the Amazon SNS topic that receives the events.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `configuration_set_arn` - ARN of the parent configuration set.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_pinpointsmsvoicev2_event_destination.example
  identity = {
    configuration_set_name = "example-configuration-set"
    event_destination_name = "example-event-destination"
  }
}

resource "aws_pinpointsmsvoicev2_event_destination" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `configuration_set_name` (String) Name of the configuration set this event destination belongs to.
* `event_destination_name` (String) Name of the event destination.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import an event destination using the `configuration_set_name` and `event_destination_name`, separated by a comma. For example:

```terraform
import {
  to = aws_pinpointsmsvoicev2_event_destination.example
  id = "example-configuration-set,example-event-destination"
}
```

Using `terraform import`, import an event destination using the `configuration_set_name` and `event_destination_name`, separated by a comma. For example:

```console
% terraform import aws_pinpointsmsvoicev2_event_destination.example "example-configuration-set,example-event-destination"
```
