---
subcategory: "CloudWatch Logs"
layout: "aws"
page_title: "AWS: aws_cloudwatch_log_destinations"
description: |-
  Terraform data source for managing an AWS CloudWatch Logs Destinations.
---

# Data Source: aws_cloudwatch_log_destinations

Terraform data source for managing an AWS CloudWatch Logs Destinations.

## Example Usage

### Basic Usage

```terraform
data "aws_cloudwatch_log_destinations" "example" {
}

data "aws_cloudwatch_log_destinations" "example_with_prefix" {
  destination_name_prefix = "example"
}
```

## Argument Reference

The following arguments are optional:

* `destination_name_prefix` - (Optional) Prefix to match. If you don't specify a value, no prefix filter is applied.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `destinations` - A list of CloudWatch Log Destinations. Check the [Cloudwatch Log Destination](#cloudwatch-log-destination) below for details.

### Cloudwatch Log Destination

The following attributes are exported:

* `arn` - ARN of the CloudWatch Log Destination.
* `creation_time` - Creation time of the CloudWatch Log Destination.
* `destination_name` - Name of the CloudWatch Log Destination.
* `role_arn` - ARN of the IAM role that grants Amazon CloudWatch Logs permissions to put data into the target.
* `target_arn` - ARN of the physical target where the log data will be delivered (eg. ARN of a Kinesis stream).
* `access_policy` - IAM policy document that governs which AWS accounts can create subscription filters against this destination.
