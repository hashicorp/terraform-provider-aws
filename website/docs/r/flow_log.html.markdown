---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_flow_log"
description: |-
  Provides a VPC/Subnet/ENI Flow Log
---

# Resource: aws_flow_log

Provides a VPC/Subnet/ENI/Transit Gateway/Transit Gateway Attachment Flow Log to capture IP traffic for a specific network
interface, subnet, or VPC. Logs are sent to a CloudWatch Log Group, a S3 Bucket, or Amazon Kinesis Data Firehose

## Example Usage

### CloudWatch Logging

```terraform
resource "aws_flow_log" "example" {
  iam_role_arn    = aws_iam_role.example.arn
  log_destination = aws_cloudwatch_log_group.example.arn
  traffic_type    = "ALL"
  vpc_id          = aws_vpc.example.id
}

resource "aws_cloudwatch_log_group" "example" {
  name = "example"
}

data "aws_iam_policy_document" "assume_role" {
  statement {
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["vpc-flow-logs.amazonaws.com"]
    }

    actions = ["sts:AssumeRole"]
  }
}

resource "aws_iam_role" "example" {
  name               = "example"
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

data "aws_iam_policy_document" "example" {
  statement {
    effect = "Allow"

    actions = [
      "logs:CreateLogGroup",
      "logs:CreateLogStream",
      "logs:PutLogEvents",
      "logs:DescribeLogGroups",
      "logs:DescribeLogStreams",
    ]

    resources = ["*"]
  }
}

resource "aws_iam_role_policy" "example" {
  name   = "example"
  role   = aws_iam_role.example.id
  policy = data.aws_iam_policy_document.example.json
}
```

### Amazon Kinesis Data Firehose logging

```terraform
resource "aws_flow_log" "example" {
  log_destination      = aws_kinesis_firehose_delivery_stream.example.arn
  log_destination_type = "kinesis-data-firehose"
  traffic_type         = "ALL"
  vpc_id               = aws_vpc.example.id
}

resource "aws_kinesis_firehose_delivery_stream" "example" {
  name        = "kinesis_firehose_test"
  destination = "extended_s3"

  extended_s3_configuration {
    role_arn   = aws_iam_role.example.arn
    bucket_arn = aws_s3_bucket.example.arn
  }

  tags = {
    "LogDeliveryEnabled" = "true"
  }
}

resource "aws_s3_bucket" "example" {
  bucket = "example"
}

resource "aws_s3_bucket_acl" "example" {
  bucket = aws_s3_bucket.example.id
  acl    = "private"
}

data "aws_iam_policy_document" "assume_role" {
  statement {
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["firehose.amazonaws.com"]
    }

    actions = ["sts:AssumeRole"]
  }
}

resource "aws_iam_role" "example" {
  name               = "firehose_test_role"
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

data "aws_iam_policy_document" "example" {
  effect = "Allow"

  actions = [
    "logs:CreateLogDelivery",
    "logs:DeleteLogDelivery",
    "logs:ListLogDeliveries",
    "logs:GetLogDelivery",
    "firehose:TagDeliveryStream",
  ]

  resources = ["*"]
}

resource "aws_iam_role_policy" "example" {
  name   = "test"
  role   = aws_iam_role.example.id
  policy = data.aws_iam_policy_document.example.json
}
```

### S3 Logging

```terraform
resource "aws_flow_log" "example" {
  log_destination      = aws_s3_bucket.example.arn
  log_destination_type = "s3"
  traffic_type         = "ALL"
  vpc_id               = aws_vpc.example.id
}

resource "aws_s3_bucket" "example" {
  bucket = "example"
}
```

### S3 Logging in Apache Parquet format with per-hour partitions

```terraform
resource "aws_flow_log" "example" {
  log_destination      = aws_s3_bucket.example.arn
  log_destination_type = "s3"
  traffic_type         = "ALL"
  vpc_id               = aws_vpc.example.id
  destination_options {
    file_format        = "parquet"
    per_hour_partition = true
  }
}

resource "aws_s3_bucket" "example" {
  bucket = "example"
}
```

## Argument Reference

~> **NOTE:** One of `eni_id`, `subnet_id`, `transit_gateway_id`, `transit_gateway_attachment_id`, or `vpc_id` must be specified.

This resource supports the following arguments:

* `traffic_type` - (Required) The type of traffic to capture. Valid values: `ACCEPT`,`REJECT`, `ALL`.
* `deliver_cross_account_role` - (Optional) ARN of the IAM role that allows Amazon EC2 to publish flow logs across accounts.
* `eni_id` - (Optional) Elastic Network Interface ID to attach to
* `iam_role_arn` - (Optional) The ARN for the IAM role that's used to post flow logs to a CloudWatch Logs log group
* `log_destination_type` - (Optional) The type of the logging destination. Valid values: `cloud-watch-logs`, `s3`, `kinesis-data-firehose`. Default: `cloud-watch-logs`.
* `log_destination` - (Optional) The ARN of the logging destination. Either `log_destination` or `log_group_name` must be set.
* `log_group_name` - (Optional) **Deprecated:** Use `log_destination` instead. The name of the CloudWatch log group. Either `log_group_name` or `log_destination` must be set.
* `subnet_id` - (Optional) Subnet ID to attach to
* `transit_gateway_id` - (Optional) Transit Gateway ID to attach to
* `transit_gateway_attachment_id` - (Optional) Transit Gateway Attachment ID to attach to
* `vpc_id` - (Optional) VPC ID to attach to
* `log_format` - (Optional) The fields to include in the flow log record. Accepted format example: `"$${interface-id} $${srcaddr} $${dstaddr} $${srcport} $${dstport}"`.
* `max_aggregation_interval` - (Optional) The maximum interval of time
  during which a flow of packets is captured and aggregated into a flow
  log record. Valid Values: `60` seconds (1 minute) or `600` seconds (10
  minutes). Default: `600`. When `transit_gateway_id` or `transit_gateway_attachment_id` is specified, `max_aggregation_interval` *must* be 60 seconds (1 minute).
* `destination_options` - (Optional) Describes the destination options for a flow log. More details below.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### destination_options

Describes the destination options for a flow log.

* `file_format` - (Optional) The format for the flow log. Default value: `plain-text`. Valid values: `plain-text`, `parquet`.
* `hive_compatible_partitions` - (Optional) Indicates whether to use Hive-compatible prefixes for flow logs stored in Amazon S3. Default value: `false`.
* `per_hour_partition` - (Optional) Indicates whether to partition the flow log per hour. This reduces the cost and response time for queries. Default value: `false`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The Flow Log ID
* `arn` - The ARN of the Flow Log.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Flow Logs using the `id`. For example:

```terraform
import {
  to = aws_flow_log.test_flow_log
  id = "fl-1a2b3c4d"
}
```

Using `terraform import`, import Flow Logs using the `id`. For example:

```console
% terraform import aws_flow_log.test_flow_log fl-1a2b3c4d
```
