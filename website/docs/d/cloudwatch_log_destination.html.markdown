---
subcategory: "CloudWatch Logs"
layout: "aws"
page_title: "AWS: aws_logs_destination"
description: |-
  Terraform data source for managing an AWS CloudWatch Logs Destination.
---

# Data Source: aws_cloudwatch_log_destination

Terraform data source for managing an AWS CloudWatch Logs Destination.

## Example Usage

### Basic Usage

```terraform
data "aws_cloudwatch_log_destination" "example" {
  destination_name = "example"
}
```

## Argument Reference

The following arguments are required:

* `destination_name` - (Required) The name of the Log Destination.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the CloudWatch Log Destination.
* `creation_time` - Creation time of the CloudWatch Log Destination.
* `destination_name` - Name of the CloudWatch Log Destination.
* `role_arn` - ARN of the IAM role that grants Amazon CloudWatch Logs permissions to put data into the target.
* `target_arn` - ARN of the physical target where the log data will be delivered (eg. ARN of a Kinesis stream).
* `access_policy` - IAM policy document that governs which AWS accounts can create subscription filters against this destination.