---
subcategory: "SESv2 (Simple Email V2)"
layout: "aws"
page_title: "AWS: aws_sesv2_configuration_set_event_destination"
description: |-
  Terraform resource for managing an AWS SESv2 (Simple Email V2) Configuration Set Event Destination.
---

# Resource: aws_sesv2_configuration_set_event_destination

Terraform resource for managing an AWS SESv2 (Simple Email V2) Configuration Set Event Destination.

## Example Usage

### CloudWatch Destination

```terraform
resource "aws_sesv2_configuration_set" "example" {
  configuration_set_name = "example"
}

resource "aws_sesv2_configuration_set_event_destination" "example" {
  configuration_set_name = aws_sesv2_configuration_set.example.configuration_set_name
  event_destination_name = "example"

  event_destination {
    cloud_watch_destination {
      dimension_configuration {
        default_dimension_value = "example"
        dimension_name          = "example"
        dimension_value_source  = "MESSAGE_TAG"
      }
    }

    enabled              = true
    matching_event_types = ["SEND"]
  }
}
```

### EventBridge Destination

```terraform
data "aws_cloudwatch_event_bus" "default" {
  name = "default"
}

resource "aws_sesv2_configuration_set_event_destination" "example" {
  configuration_set_name = aws_sesv2_configuration_set.example.configuration_set_name
  event_destination_name = "example"

  event_destination {
    event_bridge_destination {
      event_bus_arn = data.aws_cloudwatch_event_bus.default.arn
    }

    enabled              = true
    matching_event_types = ["SEND"]
  }
}
```

### Kinesis Firehose Destination

```terraform
resource "aws_sesv2_configuration_set" "example" {
  configuration_set_name = "example"
}

resource "aws_sesv2_configuration_set_event_destination" "example" {
  configuration_set_name = aws_sesv2_configuration_set.example.configuration_set_name
  event_destination_name = "example"

  event_destination {
    kinesis_firehose_destination {
      delivery_stream_arn = aws_kinesis_firehose_delivery_stream.example.arn
      iam_role_arn        = aws_iam_role.example.arn
    }

    enabled              = true
    matching_event_types = ["SEND"]
  }
}
```

### Pinpoint Destination

```terraform
resource "aws_sesv2_configuration_set" "example" {
  configuration_set_name = "example"
}

resource "aws_sesv2_configuration_set_event_destination" "example" {
  configuration_set_name = aws_sesv2_configuration_set.example.configuration_set_name
  event_destination_name = "example"

  event_destination {
    pinpoint_destination {
      application_arn = aws_pinpoint_app.example.arn
    }

    enabled              = true
    matching_event_types = ["SEND"]
  }
}
```

### SNS Destination

```terraform
resource "aws_sesv2_configuration_set" "example" {
  configuration_set_name = "example"
}

resource "aws_sesv2_configuration_set_event_destination" "example" {
  configuration_set_name = aws_sesv2_configuration_set.example.configuration_set_name
  event_destination_name = "example"

  event_destination {
    sns_destination {
      topic_arn = aws_sns_topic.example.arn
    }

    enabled              = true
    matching_event_types = ["SEND"]
  }
}
```

## Argument Reference

The following arguments are required:

* `configuration_set_name` - (Required) The name of the configuration set.
* `event_destination` - (Required) A name that identifies the event destination within the configuration set.
* `event_destination_name` - (Required) An object that defines the event destination. See [`event_destination` Block](#event_destination-block) for details.

### `event_destination` Block

The `event_destination` configuration block supports the following arguments:

* `matching_event_types` - (Required) - An array that specifies which events the Amazon SES API v2 should send to the destinations. Valid values: `SEND`, `REJECT`, `BOUNCE`, `COMPLAINT`, `DELIVERY`, `OPEN`, `CLICK`, `RENDERING_FAILURE`, `DELIVERY_DELAY`, `SUBSCRIPTION`.
* `cloud_watch_destination` - (Optional) An object that defines an Amazon CloudWatch destination for email events. See [`cloud_watch_destination` Block](#cloud_watch_destination-block) for details.
* `enabled` - (Optional) When the event destination is enabled, the specified event types are sent to the destinations. Default: `false`.
* `event_bridge_configuration` - (Optional) An object that defines an Amazon EventBridge destination for email events. You can use Amazon EventBridge to send notifications when certain email events occur. See [`event_bridge_configuration` Block](#event_bridge_configuration-block) for details.
* `kinesis_firehose_destination` - (Optional) An object that defines an Amazon Kinesis Data Firehose destination for email events. See [`kinesis_firehose_destination` Block](#kinesis_firehose_destination-block) for details.
* `pinpoint_destination` - (Optional) An object that defines an Amazon Pinpoint project destination for email events. See [`pinpoint_destination` Block](#pinpoint_destination-block) for details.
* `sns_destination` - (Optional) An object that defines an Amazon SNS destination for email events. See [`sns_destination` Block](#sns_destination-block) for details.

### `cloud_watch_destination` Block

The `cloud_watch_destination` configuration block supports the following arguments:

* `dimension_configuration` - (Required) An array of objects that define the dimensions to use when you send email events to Amazon CloudWatch. See [`dimension_configuration` Block](#dimension_configuration-block) for details.

### `dimension_configuration` Block

The `dimension_configuration` configuration block supports the following arguments:

* `default_dimension_value` - (Required) The default value of the dimension that is published to Amazon CloudWatch if you don't provide the value of the dimension when you send an email.
* `dimension_name` - (Required) The name of an Amazon CloudWatch dimension associated with an email sending metric.
* `dimension_value_source` - (Required) The location where the Amazon SES API v2 finds the value of a dimension to publish to Amazon CloudWatch. Valid values: `MESSAGE_TAG`, `EMAIL_HEADER`, `LINK_TAG`.

### `event_bridge_configuration` Block

The `event_bridge_configuration` configuration block supports the following arguments:

* `event_bus_arn` - (Required) The Amazon Resource Name (ARN) of the Amazon EventBridge bus to publish email events to. Only the default bus is supported.

### `kinesis_firehose_destination` Block

The `kinesis_firehose_destination` configuration block supports the following arguments:

* `delivery_stream_arn` - (Required) The Amazon Resource Name (ARN) of the Amazon Kinesis Data Firehose stream that the Amazon SES API v2 sends email events to.
* `iam_role_arn` - (Required) The Amazon Resource Name (ARN) of the IAM role that the Amazon SES API v2 uses to send email events to the Amazon Kinesis Data Firehose stream.

### `pinpoint_destination` Block

The `pinpoint_destination` configuration block supports the following arguments:

* `pinpoint_application_arn` - (Required) The Amazon Resource Name (ARN) of the Amazon Pinpoint project to send email events to.

### `sns_destination` Block

The `sns_destination` configuration block supports the following arguments:

* `topic_arn` - (Required) The Amazon Resource Name (ARN) of the Amazon SNS topic to publish email events to.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - A pipe-delimited string combining `configuration_set_name` and `event_destination_name`.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import SESv2 (Simple Email V2) Configuration Set Event Destination using the `id` (`configuration_set_name|event_destination_name`). For example:

```terraform
import {
  to = aws_sesv2_configuration_set_event_destination.example
  id = "example_configuration_set|example_event_destination"
}
```

Using `terraform import`, import SESv2 (Simple Email V2) Configuration Set Event Destination using the `id` (`configuration_set_name|event_destination_name`). For example:

```console
% terraform import aws_sesv2_configuration_set_event_destination.example example_configuration_set|example_event_destination
```
