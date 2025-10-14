---
subcategory: "SES (Simple Email)"
layout: "aws"
page_title: "AWS: aws_ses_event_destination"
description: |-
  Provides an SES event destination
---

# Resource: aws_ses_event_destination

Provides an SES event destination

## Example Usage

### CloudWatch Destination

```terraform
resource "aws_ses_event_destination" "cloudwatch" {
  name                   = "event-destination-cloudwatch"
  configuration_set_name = aws_ses_configuration_set.example.name
  enabled                = true
  matching_types         = ["bounce", "send"]

  cloudwatch_destination {
    default_value  = "default"
    dimension_name = "dimension"
    value_source   = "emailHeader"
  }
}
```

### Kinesis Destination

```terraform
resource "aws_ses_event_destination" "kinesis" {
  name                   = "event-destination-kinesis"
  configuration_set_name = aws_ses_configuration_set.example.name
  enabled                = true
  matching_types         = ["bounce", "send"]

  kinesis_destination {
    stream_arn = aws_kinesis_firehose_delivery_stream.example.arn
    role_arn   = aws_iam_role.example.arn
  }
}
```

### SNS Destination

```terraform
resource "aws_ses_event_destination" "sns" {
  name                   = "event-destination-sns"
  configuration_set_name = aws_ses_configuration_set.example.name
  enabled                = true
  matching_types         = ["bounce", "send"]

  sns_destination {
    topic_arn = aws_sns_topic.example.arn
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `name` - (Required) The name of the event destination
* `configuration_set_name` - (Required) The name of the configuration set
* `enabled` - (Optional) If true, the event destination will be enabled
* `matching_types` - (Required) A list of matching types. May be any of `"send"`, `"reject"`, `"bounce"`, `"complaint"`, `"delivery"`, `"open"`, `"click"`, or `"renderingFailure"`.
* `cloudwatch_destination` - (Optional) CloudWatch destination for the events
* `kinesis_destination` - (Optional) Send the events to a kinesis firehose destination
* `sns_destination` - (Optional) Send the events to an SNS Topic destination

~> **NOTE:** You can specify `"cloudwatch_destination"` or `"kinesis_destination"` but not both

### cloudwatch_destination Argument Reference

* `default_value` - (Required) The default value for the event
* `dimension_name` - (Required) The name for the dimension
* `value_source` - (Required) The source for the value. May be any of `"messageTag"`, `"emailHeader"` or `"linkTag"`.

### kinesis_destination Argument Reference

* `stream_arn` - (Required) The ARN of the Kinesis Stream
* `role_arn` - (Required) The ARN of the role that has permissions to access the Kinesis Stream

### sns_destination Argument Reference

* `topic_arn` - (Required) The ARN of the SNS topic

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The SES event destination name.
* `arn` - The SES event destination ARN.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import SES event destinations using `configuration_set_name` together with the event destination's `name`. For example:

```terraform
import {
  to = aws_ses_event_destination.sns
  id = "some-configuration-set-test/event-destination-sns"
}
```

Using `terraform import`, import SES event destinations using `configuration_set_name` together with the event destination's `name`. For example:

```console
% terraform import aws_ses_event_destination.sns some-configuration-set-test/event-destination-sns
```
