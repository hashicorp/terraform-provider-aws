---
subcategory: "Network Firewall"
layout: "aws"
page_title: "AWS: aws_networkfirewall_logging_configuration"
description: |-
  Provides an AWS Network Firewall Logging Configuration resource.
---

# Resource: aws_networkfirewall_logging_configuration

Provides an AWS Network Firewall Logging Configuration Resource

## Example Usage

### Logging to S3

```terraform
resource "aws_networkfirewall_logging_configuration" "example" {
  firewall_arn = aws_networkfirewall_firewall.example.arn
  logging_configuration {
    log_destination_config {
      log_destination = {
        bucketName = aws_s3_bucket.example.bucket
        prefix     = "example"
      }
      log_destination_type = "S3"
      log_type             = "FLOW"
    }
  }
}
```

### Logging to CloudWatch

```terraform
resource "aws_networkfirewall_logging_configuration" "example" {
  firewall_arn = aws_networkfirewall_firewall.example.arn
  logging_configuration {
    log_destination_config {
      log_destination = {
        logGroup = aws_cloudwatch_log_group.example.name
      }
      log_destination_type = "CloudWatchLogs"
      log_type             = "ALERT"
    }
  }
}
```

### Logging to Kinesis Data Firehose

```terraform
resource "aws_networkfirewall_logging_configuration" "example" {
  firewall_arn = aws_networkfirewall_firewall.example.arn
  logging_configuration {
    log_destination_config {
      log_destination = {
        deliveryStream = aws_kinesis_firehose_delivery_stream.example.name
      }
      log_destination_type = "KinesisDataFirehose"
      log_type             = "TLS"
    }
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `firewall_arn` - (Required, Forces new resource) The Amazon Resource Name (ARN) of the Network Firewall firewall.

* `logging_configuration` - (Required) A configuration block describing how AWS Network Firewall performs logging for a firewall. See [Logging Configuration](#logging-configuration) below for details.

### Logging Configuration

The `logging_configuration` block supports the following arguments:

* `log_destination_config` - (Required) Set of configuration blocks describing the logging details for a firewall. See [Log Destination Config](#log-destination-config) below for details. At most, only Three blocks can be specified; one for `FLOW` logs and one for `ALERT` logs and one for `TLS` logs.

### Log Destination Config

The `log_destination_config` block supports the following arguments:

* `log_destination` - (Required) A map describing the logging destination for the chosen `log_destination_type`.
    * For an Amazon S3 bucket, specify the key `bucketName` with the name of the bucket and optionally specify the key `prefix` with a path (Do not add a leading / in the `prefix` as the configuration will have two // when applied).
    * For a CloudWatch log group, specify the key `logGroup` with the name of the CloudWatch log group.
    * For a Kinesis Data Firehose delivery stream, specify the key `deliveryStream` with the name of the delivery stream.

* `log_destination_type` - (Required) The location to send logs to. Valid values: `S3`, `CloudWatchLogs`, `KinesisDataFirehose`.

* `log_type` - (Required) The type of log to send. Valid values: `ALERT` or `FLOW` or `TLS`. Alert logs report traffic that matches a `StatefulRule` with an action setting that sends a log message. Flow logs are standard network traffic flow logs.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The Amazon Resource Name (ARN) of the associated firewall.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Network Firewall Logging Configurations using the `firewall_arn`. For example:

```terraform
import {
  to = aws_networkfirewall_logging_configuration.example
  id = "arn:aws:network-firewall:us-west-1:123456789012:firewall/example"
}
```

Using `terraform import`, import Network Firewall Logging Configurations using the `firewall_arn`. For example:

```console
% terraform import aws_networkfirewall_logging_configuration.example arn:aws:network-firewall:us-west-1:123456789012:firewall/example
```
