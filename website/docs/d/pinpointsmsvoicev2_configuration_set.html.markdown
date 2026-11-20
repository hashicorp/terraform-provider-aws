---
subcategory: "End User Messaging SMS"
layout: "aws"
page_title: "AWS: aws_pinpointsmsvoicev2_configuration_set"
description: |-
  Provides details about an AWS End User Messaging SMS Configuration Set.
---

# Data Source: aws_pinpointsmsvoicev2_configuration_set

Provides details about an AWS End User Messaging SMS Configuration Set.

## Example Usage

### Basic Usage

```terraform
data "aws_pinpointsmsvoicev2_configuration_set" "example" {
  name = "example-configuration-set"
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the configuration set.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the configuration set.
* `created_timestamp` - Time when the configuration set was created, in [RFC3339 format](https://tools.ietf.org/html/rfc3339#section-5.8).
* `default_message_feedback_enabled` - Whether message feedback is enabled by default for messages sent with this configuration set.
* `default_message_type` - Default message type of the configuration set. Either `TRANSACTIONAL` or `PROMOTIONAL`.
* `default_sender_id` - Default sender ID of the configuration set.
* `event_destinations` - List of event destinations of the configuration set. See [`event_destinations` Block](#event_destinations-block) for details.
* `protect_configuration_id` - ID of the protect configuration associated with the configuration set.
* `tags` - Map of tags assigned to the configuration set.

### `event_destinations` Block

* `cloudwatch_logs_destination` - Amazon CloudWatch Logs destination of the event destination. See [`event_destinations.cloudwatch_logs_destination` Block](#event_destinationscloudwatch_logs_destination-block) for details.
* `enabled` - Whether the event destination is enabled.
* `event_destination_name` - Name of the event destination.
* `kinesis_firehose_destination` - Amazon Data Firehose destination of the event destination. See [`event_destinations.kinesis_firehose_destination` Block](#event_destinationskinesis_firehose_destination-block) for details.
* `matching_event_types` - Event types for which the destination receives records.
* `sns_destination` - Amazon SNS destination of the event destination. See [`event_destinations.sns_destination` Block](#event_destinationssns_destination-block) for details.

### `event_destinations.cloudwatch_logs_destination` Block

* `iam_role_arn` - ARN of the IAM role that End User Messaging SMS assumes to write to the log group.
* `log_group_arn` - ARN of the Amazon CloudWatch log group that receives the events.

### `event_destinations.kinesis_firehose_destination` Block

* `delivery_stream_arn` - ARN of the Amazon Data Firehose delivery stream that receives the events.
* `iam_role_arn` - ARN of the IAM role that End User Messaging SMS assumes to write to the delivery stream.

### `event_destinations.sns_destination` Block

* `topic_arn` - ARN of the Amazon SNS topic that receives the events.
